package main

import (
	"github.com/mumoshu/sops-vault/app"
	"github.com/mumoshu/sops-vault/cmd"
)

func main() {
	ctx := app.NewContext()
	ap := app.NewApp(
		ctx,
		app.NewVault("kube-aws").UsedForCommand("kube-aws").StoresFilesMatchingGlob("credentials/*-key.pem", "credentials/tokens.csv"),
		app.NewVault("kubectl").UsedForCommand("kubectl", "helm").StoresFilesMatchingGlob("kubeconfig"),
	)

	cmd.GenerateAndRun(ap)
}
