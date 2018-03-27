## sopsed

A convenient wrapper command for automatically encrypting/decrypting files with sops.

Out-of-box supports for `kube-aws`, `helm`, `helmfile` and `kubectl`.

Use this as a golang library to easily add supports for the commands of your choice.

## Install

Grab the latest binary from [the GitHub releases page](https://github.com/mumoshu/sopsed/releases).

## Pre-requisite

Create a `.sops.yaml` to tell `sops` which key to be used for (re)encrypting files:

For AWS KMS:

```
creation_rules:
    - kms: "arn:aws:kms:<aws region>:<aws account id>:key/<key id>
```

To separate keys for different environments:

```
creation_rules:
    - filename_regex: environments/test/.*
      kms: "arn:aws:kms:<aws region>:<aws account id>:key/<key #1 id>
    - filename_regex: environments/prod/.*
      kms: "arn:aws:kms:<aws region>:<aws account id>:key/<key #2 id>
```

## Usage

```
# Automatically encrypts/decrypts `credentials/*-key.pem` before/after running `kube-aws` a sub-command
sopsed run kube-aws update ...

# Do the same for `./kubeconfig` before/after running `kubectl` and `helm` sub-commands
sopsed run helm ...
sopsed run kubectl ...
```

See [the documentation resides in this repository](https://github.com/mumoshu/sopsed/blob/master/docs/sopsed.md) for more detailed usage of each command.

## Inspirations

- [miquella/vaulted](https://github.com/miquella/vaulted): vaulted uses a password from human-input to protect your vault, whereas sopsed utilizes KMS/GPG via sops instead.
