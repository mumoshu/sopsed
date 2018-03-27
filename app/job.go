package app

import (
	"strings"

	"go.mozilla.org/sops/decrypt"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

type Job struct {
	*VaultConfig
	context *Context
}

type assets struct {
	files   map[string]string
	paths   []string
	context *Context
}

func readAssetsFromFile(encryptedVault string, context *Context) (*assets, error) {
	filesInVault := map[string]string{}
	filepathesInVault := []string{}
	encryptedVaultExists := fileExists(encryptedVault)
	if encryptedVaultExists {
		// TODO shell-out rather than using sops as a library, so that we can easily supress logs from the library
		jsonBytes, err := decrypt.File(encryptedVault, "json")
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %v", encryptedVault, err)
		}
		if err := json.Unmarshal(jsonBytes, &filesInVault); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %s: %v", encryptedVault, err)
		}
		for k := range filesInVault {
			if k != "sops" {
				filepathesInVault = append(filepathesInVault, k)
			}
		}
	}
	return &assets{
		context: context,
		files:   filesInVault,
		paths:   filepathesInVault,
	}, nil
}

func (a *assets) writeToFile(unencryptedVault string, encryptedVault string) error {
	// We don't use `sops --set key value` here so that we won't expose sensitive values in the process list
	cleartextFilesInJSON, err := json.Marshal(a.files)
	if err != nil {
		return err
	}

	defer a.cleanup(unencryptedVault)
	if err := ioutil.WriteFile(unencryptedVault, cleartextFilesInJSON, 0644); err != nil {
		return err
	}

	out, err := runAndCaptureStdout(a.context, "sh", "-c", fmt.Sprintf("sops --encrypt %s", unencryptedVault))
	if err != nil {
		if strings.Contains(err.Error(), "config file not found and no keys provided through command line options") {
			a.context.Debug(err.Error())
			return fmt.Errorf("`sops --encrypt` failed: .sops.yaml seems to be missing. please create one following the steps in readme(https://github.com/mumoshu/sopsed)")
		}
		if strings.Contains(err.Error(), "Failed to call KMS encryption service: ExpiredTokenException: The security token included in the request is expired") {
			a.context.Debug(err.Error())
			return fmt.Errorf("`sops --encrypt` failed: aws credentials seem to be missing or expired")
		}
		return err
	}

	if err := ioutil.WriteFile(encryptedVault, []byte(out), 0644); err != nil {
		return err
	}
	return nil
}

func (a *assets) addFilesMatchingPatterns(insecureFilePatterns []entry) (*assets, []string, error) {
	newlyRecognizedFiles := []string{}
	for _, e := range insecureFilePatterns {
		files, err := filepath.Glob(e.pathPattern)
		if err != nil {
			return nil, []string{}, err
		}
		for _, f := range files {
			fmt.Printf("found %s\n", f)
			alreadyEncrypted := false
			for _, path := range a.paths {
				if path == f {
					alreadyEncrypted = true
				}
			}
			if alreadyEncrypted {
				a.context.Debug(fmt.Sprintf("skipping %s: already encrypted. you can safely remove it", f))
			} else {
				a.context.Debug(fmt.Sprintf("adding %s to the vault", f))
			}
			raw, err := ioutil.ReadFile(f)
			if err != nil {
				return nil, []string{}, err
			}
			a.files[f] = string(raw)
			newlyRecognizedFiles = append(newlyRecognizedFiles, f)
		}
	}
	return a, newlyRecognizedFiles, nil
}

func (app *Job) Encrypt() error {
	context := app.context
	unencryptedVault := app.unencryptedVault()
	encryptedVault := app.encryptedVault()
	insecureFilePatterns := app.entries

	assets, err := readAssetsFromFile(encryptedVault, context)

	if err != nil {
		return err
	}

	var newlyRecognizedFiles []string

	assets, newlyRecognizedFiles, err = assets.addFilesMatchingPatterns(insecureFilePatterns)

	if err != nil {
		return err
	}

	if len(assets.files) > 0 {
		if err := assets.writeToFile(unencryptedVault, encryptedVault); err != nil {
			return err
		}
	} else {
		patterns := []string{}
		for _, p := range insecureFilePatterns {
			patterns = append(patterns, fmt.Sprintf(`"%s"`, p.pathPattern))
		}
		return fmt.Errorf("you must have \"%s\" or files matching any of [%s] to run %s", encryptedVault, strings.Join(patterns, ", "), app.vaultName)
	}

	for _, path := range newlyRecognizedFiles {
		context.Info(fmt.Sprintf("renaming %s to %s.bak: it is already encrypted into %s. you can safely remove it", path, path, encryptedVault))
		err := os.Rename(path, fmt.Sprintf("%s.bak", path))
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *Job) Decrypt() (func(), error) {
	context := app.context
	encryptedVault := app.encryptedVault()

	jsonBytes, err := decrypt.File(encryptedVault, "json")
	if err != nil {
		return nil, err
	}
	var restoredFiles map[string]string
	if err := json.Unmarshal(jsonBytes, &restoredFiles); err != nil {
		return nil, err
	}

	restoredFilePathes := []string{}

	for path := range restoredFiles {
		restoredFilePathes = append(restoredFilePathes, path)
	}

	for path, content := range restoredFiles {
		context.Debug(fmt.Sprintf("restoring %s", path))
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return nil, err
		}
	}

	return func() { app.cleanup(restoredFilePathes...) }, nil
}

// RunOrPanic runs the app with the provided configuration. On any error it panics
func (app *Job) RunOrPanic(command string, args ...string) error {
	if err := app.Encrypt(); err != nil {
		return err
	}

	cleanup, err := app.Decrypt()
	if err != nil {
		return err
	}

	defer cleanup()

	app.context.Debug(fmt.Sprintf("running %s %s", command, strings.Join(args, " ")))
	if err := runInForeground(command, args...); err != nil {
		return fmt.Errorf("run: %v", err)
	}
	return nil
}

func (j *Job) cleanup(restoredFiles ...string) {
	context := j.context
	for _, path := range restoredFiles {
		context.Debug(fmt.Sprintf("removing %s", path))
		err := os.Remove(path)
		if err != nil {
			context.Warn(fmt.Sprintf("failed to remove %s: BEWARE THAT NO CLEARTEXT FILE IS REMAINING!", path))
		}
	}
}

func (j *assets) cleanup(restoredFiles ...string) {
	context := j.context
	for _, path := range restoredFiles {
		context.Debug(fmt.Sprintf("removing %s", path))
		err := os.Remove(path)
		if err != nil {
			context.Warn(fmt.Sprintf("failed to remove %s: BEWARE THAT NO CLEARTEXT FILE IS REMAINING!", path))
		}
	}
}
