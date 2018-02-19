package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sops-vault",
	Short: "sops-vault is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run",
	Long: `sops-vault is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run.
				  Complete documentation is available at https://github.com/mumoshu/sops-vault`,
}

func Execute() {
	ctx := NewContext()
	NewApp(ctx).RunWithVaults(
		NewVault("kube-aws").UsedForCommand("kube-aws").StoresFilesMatchingGlob("credentials/*-key.pem", "credentials/tokens.csv"),
		NewVault("kubectl").UsedForCommand("kubectl", "helm").StoresFilesMatchingGlob("kubeconfig"),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
