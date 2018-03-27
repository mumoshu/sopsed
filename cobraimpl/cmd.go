package cobraimpl

import (
	"github.com/mumoshu/sopsed/app"
	"github.com/mumoshu/sopsed/cmd"
	"github.com/spf13/cobra"
)

func CreateCommand() *cobra.Command {
	ctx := app.NewContext()
	ap := app.NewApp(
		ctx,
		app.NewVault("kube-aws").UsedForCommand("kube-aws").StoresFilesMatchingGlob("credentials/*-key.pem", "credentials/tokens.csv", "credentials/kubelet-tls-bootstrap-token"),
		app.NewVault("kubectl").UsedForCommand("kubectl", "helm", "helm-secrets", "helmfile").StoresFilesMatchingGlob("kubeconfig"),
	)

	cmd.Init(ap)

	return cmd.RootCmd
}
