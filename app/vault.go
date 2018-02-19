package main

import "fmt"

type entry struct {
	pathPattern string
}

// VaultConfig contains all the configuration variables for sops-exec
type VaultConfig struct {
	vaultName string
	entries   []entry
	commands  []string
}

type VaultBuilder struct {
	*VaultConfig
}

func NewVault(name string) *VaultBuilder {
	return &VaultBuilder{
		&VaultConfig{
			vaultName: name,
		},
	}
}

func (b *VaultBuilder) UsedForCommand(commands ...string) *VaultBuilder {
	b.commands = commands
	return b
}

func (b *VaultBuilder) StoresFilesMatchingGlob(globs ...string) *VaultBuilder {
	entries := []entry{}
	for _, g := range globs {
		entries = append(entries, entry{g})
	}
	b.entries = entries
	return b
}

func (b *VaultBuilder) Build() *VaultConfig {
	return b.VaultConfig
}

// MatchesCommand returns true if this config is for the vault of the given command
func (c *VaultConfig) MatchesCommand(command string, args ...string) bool {
	for _, c := range c.commands {
		if c == command {
			return true
		}
	}
	return false
}

func (c *VaultConfig) unencryptedVault() string {
	return fmt.Sprintf(".sops.vault.%s.insecure", c.vaultName)
}

func (c *VaultConfig) encryptedVault() string {
	return fmt.Sprintf(".sops.vault.%s", c.vaultName)
}
