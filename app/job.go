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

// RunOrPanic runs the app with the provided configuration. On any error it panics
func (app *Job) RunOrPanic(command string, args ...string) error {
	context := app.context

	unencryptedVault := app.unencryptedVault()
	encryptedVault := app.encryptedVault()
	insecureFilePatterns := app.entries

	filesInVault := map[string]string{}
	filepathesInVault := []string{}
	encryptedVaultExists := fileExists(encryptedVault)
	newlyRecognizedFiles := []string{}
	if encryptedVaultExists {
		// TODO shell-out rather than using sops as a library, so that we can easily supress logs from the library
		jsonBytes, err := decrypt.File(encryptedVault, "json")
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %v", encryptedVault, err)
		}
		if err := json.Unmarshal(jsonBytes, &filesInVault); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %v", encryptedVault, err)
		}
		for k := range filesInVault {
			if k != "sops" {
				filepathesInVault = append(filepathesInVault, k)
			}
		}
	}
	for _, e := range insecureFilePatterns {
		files, err := filepath.Glob(e.pathPattern)
		if err != nil {
			return err
		}
		for _, f := range files {
			fmt.Printf("found %s\n", f)
			alreadyEncrypted := false
			for _, path := range filepathesInVault {
				if path == f {
					alreadyEncrypted = true
				}
			}
			if alreadyEncrypted {
				context.Debug(fmt.Sprintf("skipping %s: already encrypted. you can safely remove it", f))
			} else {
				context.Debug(fmt.Sprintf("adding %s to the vault", f))
			}
			raw, err := ioutil.ReadFile(f)
			if err != nil {
				return err
			}
			filesInVault[f] = string(raw)
			newlyRecognizedFiles = append(newlyRecognizedFiles, f)
		}
	}

	vaultSize := len(filesInVault)

	if vaultSize > 0 {
		{
			// We don't use `sops --set key value` here so that we won't expose sensitive values in the process list
			cleartextFilesInJSON, err := json.Marshal(filesInVault)
			if err != nil {
				return err
			}

			defer app.cleanup(unencryptedVault)
			if err := ioutil.WriteFile(unencryptedVault, cleartextFilesInJSON, 0644); err != nil {
				return err
			}
		}

		{
			out, err := runAndCaptureStdout(context, "sh", "-c", fmt.Sprintf("sops --encrypt %s", unencryptedVault))
			if err != nil {
				if strings.Contains(err.Error(), "config file not found and no keys provided through command line options") {
					app.context.Debug(err.Error())
					return fmt.Errorf("`sops --encrypt` failed: .sops.yaml seems to be missing. please create one following the steps in readme(https://github.com/mumoshu/sops-vault)")
				}
				if strings.Contains(err.Error(), "Failed to call KMS encryption service: ExpiredTokenException: The security token included in the request is expired") {
					app.context.Debug(err.Error())
					return fmt.Errorf("`sops --encrypt` failed: aws credentials seem to be missing or expired")
				}
				return err
			}

			if err := ioutil.WriteFile(encryptedVault, []byte(out), 0644); err != nil {
				return err
			}
		}
	} else {
		patterns := []string{}
		for _, p := range insecureFilePatterns {
			patterns = append(patterns, fmt.Sprintf(`"%s"`, p.pathPattern))
		}
		return fmt.Errorf("you must have \"%s\" or files matching any of [%s] to run %s", encryptedVault, strings.Join(patterns, ", "), command)
	}

	for _, path := range newlyRecognizedFiles {
		context.Info(fmt.Sprintf("renaming %s to %s.bak: it is already encrypted into %s. you can safely remove it", path, path, encryptedVault))
		err := os.Rename(path, fmt.Sprintf("%s.bak", path))
		if err != nil {
			return err
		}
	}

	jsonBytes, err := decrypt.File(encryptedVault, "json")
	if err != nil {
		return err
	}
	var restoredFiles map[string]string
	if err := json.Unmarshal(jsonBytes, &restoredFiles); err != nil {
		return err
	}

	restoredFilepathes := []string{}

	for path := range restoredFiles {
		restoredFilepathes = append(restoredFilepathes, path)
	}

	defer app.cleanup(restoredFilepathes...)

	for path, content := range restoredFiles {
		context.Debug(fmt.Sprintf("restoring %s", path))
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			return err
		}
	}

	context.Debug(fmt.Sprintf("running %s %s", command, strings.Join(args, " ")))
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
