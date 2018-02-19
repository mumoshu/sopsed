package main

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
}

// RunOrPanic runs the app with the provided configuration. On any error it panics
func (app *Job) RunOrPanic(context *Context, command string, args ...string) {
	unencryptedVault := app.unencryptedVault()
	encryptedVault := app.encryptedVault()
	insecureFilePatterns := app.entries

	filesInVault := map[string]string{}
	filepathesInVault := []string{}
	encryptedVaultExists := fileExists(encryptedVault)
	newlyRecognizedFiles := []string{}
	if encryptedVaultExists {
		jsonBytes, err := decrypt.File(encryptedVault, "json")
		if err != nil {
			panic(fmt.Errorf("failed to decrypt %s: %v", encryptedVault, err))
		}
		if err := json.Unmarshal(jsonBytes, &filesInVault); err != nil {
			panic(fmt.Errorf("failed to unmarshal %s: %v", encryptedVault, err))
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
			panic(err)
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
				fmt.Printf("skipping %s: already encrypted. you can safely remove it\n", f)
			} else {
				fmt.Printf("adding %s to the vault\n", f)
			}
			raw, err := ioutil.ReadFile(f)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			filesInVault[f] = string(raw)
			newlyRecognizedFiles = append(newlyRecognizedFiles, f)
		}
	}

	vaultSize := len(filesInVault)

	if vaultSize > 0 {
		// We don't use `sops --set key value` here so that we won't expose sensitive values in the process list
		cleartextFilesInJSON, err := json.Marshal(filesInVault)
		if err != nil {
			panic(err)
		}

		defer cleanup(unencryptedVault)
		err = ioutil.WriteFile(unencryptedVault, cleartextFilesInJSON, 0644)
		if err != nil {
			panic(err)
		}

		err = runInBackground("sh", "-c", fmt.Sprintf("sops --encrypt %s > %s", unencryptedVault, encryptedVault))
		if err != nil {
			panic(err)
		}
	} else {
		patterns := []string{}
		for _, p := range insecureFilePatterns {
			patterns = append(patterns, fmt.Sprintf(`"%s"`, p.pathPattern))
		}
		context.err.Printf("you must have \"%s\" or files matching any of [%s] to run %s\n", encryptedVault, strings.Join(patterns, ", "), command)
		os.Exit(1)
	}

	for _, path := range newlyRecognizedFiles {
		fmt.Printf("renaming %s to %s.bak: it is already encrypted into %s. you can safely remove it\n", path, path, encryptedVault)
		err := os.Rename(path, fmt.Sprintf("%s.bak", path))
		if err != nil {
			panic(err)
		}
	}

	jsonBytes, err := decrypt.File(encryptedVault, "json")
	if err != nil {
		panic(err)
	}
	var restoredFiles map[string]string
	if err := json.Unmarshal(jsonBytes, &restoredFiles); err != nil {
		panic(err)
	}

	fmt.Println(restoredFiles)

	restoredFilepathes := []string{}

	for path := range restoredFiles {
		restoredFilepathes = append(restoredFilepathes, path)
	}

	defer cleanup(restoredFilepathes...)

	for path, content := range restoredFiles {
		fmt.Printf("restoring %s", path)
		err := ioutil.WriteFile(path, []byte(content), 0644)
		if err != nil {
			panic(err)
		}
	}

	err = runInForeground(command, args...)
	if err != nil {
		panic(err)
	}
}

func cleanup(restoredFiles ...string) {
	for _, path := range restoredFiles {
		fmt.Printf("removing %s\n", path)
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("failed to remove %s: BEWARE THAT NO CLEARTEXT FILE IS REMAINING!\n", path)
		}
	}
}
