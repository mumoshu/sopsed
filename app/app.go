package main

import (
	"fmt"
	"os"
)

// App represents a self-contained instance of this app
type App struct {
	*Context
}

// NewApp instantiates a new app from a context
func NewApp(ctx *Context) *App {
	return &App{
		Context: ctx,
	}
}

// RunWithVaults run the app with preconfigured vault(s)
func (ctx *App) RunWithVaults(vaults ...*VaultBuilder) {
	cfgs := []*VaultConfig{}
	for _, v := range vaults {
		cfgs = append(cfgs, v.VaultConfig)
	}
	ctx.run(cfgs...)
}

// run executes the command provided via the command-line args, with temporarily decrypting necessary files according to the appropriate config
func (ctx *App) run(cfgs ...*VaultConfig) {
	var cfg *VaultConfig
	cmd, args := os.Args[1], os.Args[2:]
	for _, c := range cfgs {
		if c.MatchesCommand(cmd, args...) {
			cfg = c
			break
		}
	}
	if cfg == nil {
		panic(fmt.Errorf("no config found for command: %s", cmd))
	}
	ctx.info.Printf("using vault: %s\n", cfg.vaultName)
	job := &Job{cfg}
	job.RunOrPanic(ctx.Context, cmd, args...)
}
