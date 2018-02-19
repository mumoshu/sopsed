package main

import (
	"log"

	"github.com/mumoshu/sops-vault/cobraimpl"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cobraimpl.CreateCommand(), "./docs")
	if err != nil {
		log.Fatal(err)
	}
}
