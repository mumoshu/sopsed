package cmd

import (
	"fmt"
	"os"

	"github.com/mumoshu/sops-vault/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sops-vault",
	Short: "sops-vault is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run",
	Long: `sops-vault is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run.
				  Complete documentation is available at https://github.com/mumoshu/sops-vault`,
}

func GenerateAndRun(app *app.App) {
	runCmd := &cobra.Command{
		Use:   "run wrapped-command [args...]",
		Short: "Run wrapped-command with temporarily decrypting required files from the vault",
		Args:  cobra.MinimumNArgs(1),
	}
	rootCmd.AddCommand(runCmd)

	for _, cmd := range app.Commands() {
		c := &cobra.Command{
			Use:   fmt.Sprintf("%s [args]", cmd),
			Short: fmt.Sprintf("Run %s with temporarily decrypting required files from the vault", cmd),
			Args:  cobra.MinimumNArgs(0),
			Run: func(cmd *cobra.Command, args []string) {
				app.Run(cmd.Name(), args...)
			},
		}
		runCmd.AddCommand(c)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
