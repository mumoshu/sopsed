package cmd

import (
	"fmt"
	"os"

	"github.com/mumoshu/sopsed/app"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "sopsed",
	Short: "sopsed is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run",
	Long: `sopsed is a general wrapper for mozilla/sops to transparently encrypt/decrypt files according to the command being run.
				  Complete documentation is available at https://github.com/mumoshu/sopsed`,
	Args: cobra.NoArgs,
}

func Init(app *app.App) {
	runCmd := &cobra.Command{
		Use:   "run wrapped-command [args...]",
		Short: "Run wrapped-command with temporarily decrypting required files from the vault",
		Args:  cobra.NoArgs,
	}
	RootCmd.AddCommand(runCmd)

	for _, cmd := range app.Commands() {
		c := &cobra.Command{
			Use:   fmt.Sprintf("%s [args]", cmd),
			Short: fmt.Sprintf("Run %s with temporarily decrypting required files from the vault", cmd),
			Args:  cobra.ArbitraryArgs,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Printf("running %s\n", cmd.Name())
				app.Run(cmd.Name(), args...)
			},
		}
		c.DisableFlagParsing = true
		runCmd.AddCommand(c)
	}

	decryptCmd := &cobra.Command{
		Use:   "decrypt [vault]",
		Short: "Decrypt a named vault to produce cleartext files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			v := args[0]
			fmt.Printf("decryptiong %s\n", v)
			app.Decrypt(v)
		},
	}
	RootCmd.AddCommand(decryptCmd)

	encryptCmd := &cobra.Command{
		Use:   "encrypt [vault]",
		Short: "Encrypt a named vault to produce cleartext files",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			v := args[0]
			fmt.Printf("encrypting %s\n", v)
			app.Encrypt(v)
		},
	}
	RootCmd.AddCommand(encryptCmd)

}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
