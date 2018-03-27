package app

import (
	"fmt"
)

// App represents a self-contained instance of this app
type App struct {
	*Context
	vaults []*VaultBuilder
}

// NewApp instantiates a new app from a context
func NewApp(ctx *Context, vaults ...*VaultBuilder) *App {
	return &App{
		Context: ctx,
		vaults:  vaults,
	}
}

// Commands returns the list of all the commands this sops-vault instance is abvle to handles/wraps
func (a *App) Commands() []string {
	cmds := map[string]bool{}
	for _, cfg := range a.vaultConfigs() {
		for _, cmd := range cfg.commands {
			cmds[cmd] = true
		}
	}

	ret := []string{}
	for cmd := range cmds {
		ret = append(ret, cmd)
	}
	return ret
}

// vaultConfigs prepares preconfigured vault config(s)
func (a *App) vaultConfigs() []*VaultConfig {
	cfgs := []*VaultConfig{}
	for _, v := range a.vaults {
		cfgs = append(cfgs, v.VaultConfig)
	}
	return cfgs
}

// Run executes the command provided via the command-line args, with temporarily decrypting necessary files according to the appropriate config
func (a *App) Run(cmd string, args ...string) {
	var cfg *VaultConfig
	for _, c := range a.vaultConfigs() {
		if c.MatchesCommand(cmd, args...) {
			cfg = c
			break
		}
	}
	if cfg == nil {
		a.Context.ExitWithError(fmt.Errorf("no config found for command: %s", cmd))
	}
	a.info.Printf("using vault: %s\n", cfg.vaultName)
	job := &Job{VaultConfig: cfg, context: a.Context}
	if err := job.RunOrPanic(cmd, args...); err != nil {
		a.Context.ExitWithError(err)
	}
}

// Decrypt a named vault
func (a *App) Decrypt(vault string) {
	var cfg *VaultConfig
	for _, c := range a.vaultConfigs() {
		if c.vaultName == vault {
			cfg = c
			break
		}
	}
	if cfg == nil {
		a.Context.ExitWithError(fmt.Errorf("no vault found for command: %s", cmd))
	}
	a.info.Printf("using vault: %s\n", cfg.vaultName)
	job := &Job{VaultConfig: cfg, context: a.Context}
	if err := job.Decrypt(cmd, args...); err != nil {
		a.Context.ExitWithError(err)
	}
}
