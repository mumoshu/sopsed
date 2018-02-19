## sops-vault

A convenient wrapper for automatically encrypting/decrypting files with sops.
Out-of-box supports for `kube-aws`, `helm` and `kubectl`.
Use this as a golang library to easily add supports for the commands of your choice.

## Pre-requisite

Create a `.sops.yaml` to tell `sops` which key to be used for (re)encrypting files:

For AWS KMS:

```
creation_rules:
    - kms: "arn:aws:kms:<aws region>:<aws account id>:key/<key id>
```

## Usage

```
# Automatically encrypts/decrypts `credentials/*-key.pem` before/after running `kube-aws` a sub-command
sops-vault run kube-aws update ...

# Do the same for `./kubeconfig` before/after running `kubectl` and `helm` sub-commands
sops-vault run helm ...
sops-vault run kubectl ...
```

See [the documentation resides in this repository](https://github.com/mumoshu/sops-vault/blob/master/docs/sops-vault.md) for more detailed usage of each command.
