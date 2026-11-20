# Superchain Registry Contributing Guide

> [!WARNING]
> `CONTRIBUTING.md` contains guidelines for modifying or working with the code in the Superchain Registry, including the validation checks and exported modules.
>
> For guidelines on how to add a chain to the Registry, see the documentation for [adding a new chain](docs/ops.md#adding-a-chain).

The Superchain Registry repository contains:

- raw ["per-chain" config data](./superchain) in `toml`, `json` and `json.zst` files arranged in a semantically meaningful directory structure
- [superchain-wide config data](#superchain-wide-config-data)
- a Go workspace with
  - a [`validation`](#validation-go-module) module
  - an [`ops`](#ops-go-module) module
  - The modules are tracked by the `ops/go.work` file. The associated `go.work.sum` file is gitignored and not important to typical workflows, which should mirror those of the [CI configuration](.circleci/config.yml).
- Automatically generated summary `chainList.json` and `chainList.toml` files.

## Superchain-wide config data

A superchain target defines a set of layer 2 chains which share a `SuperchainConfig` contract deployment on layer 1. It is usually named after the layer 1 chain, possibly with an extra identifier to distinguish devnets.

> **Note**
> Example: `sepolia` and `sepolia-devnet-2` are distinct superchain targets, although they are on the same layer 1 chain.

### Adding a superchain target

A new Superchain Target can be added by creating a new superchain config directory,
with a `superchain.toml` config file.

> **Note**
> This is an infrequent operation and unnecessary if you are just looking to add a chain to an existing superchain.

Here's an example:

```bash
cd superchain-registry

export SUPERCHAIN_TARGET=sepolia-dev-1
mkdir superchain/configs/$SUPERCHAIN_TARGET

cat > superchain/configs/$SUPERCHAIN_TARGET/superchain.toml << EOF
name = "Sepolia Dev 1"
# superchain_config_addr = "0x..." # todo

[l1]
  chain_id = 11155111
  public_rpc = "https://ethereum-sepolia-rpc.publicnode.com"
  explorer = "https://eth-sepolia.blockscout.com"
EOF
```

Superchain-wide configuration, like the `SuperchainConfig` contract address, should be configured here when available.

## `superchain` directory

Per chain and superchain-wide configs are stored in `superchain/configs`, and extra data, such as the compressed genesis files, in `superchain/extra`.
The [Optimism monorepo](https://github.com/ethereum-optimism/optimism) takes the registry as a git submodule and embeds these files in its binaries at build time.

The configs are consumed by downstream OP Stack software, e.g. `op-node`, `kona-node` and `op-reth`.

## `validation` Go Module

This module contains TOML files in `validation/standard` declaring standard configuration and contract versions, and Go bindings for these declarations. It is imported by downstream software such as `op-deployer`.

## `ops` Go module

This module contains the CLI tool for generating `superchain` compliant configs and extra data to the registry.

## Setting up your editor for formatting and linting

If you use VSCode, you can place the following in a `settings.json` file in the gitignored `.vscode` directory:

```json
{
    "go.formatTool": "gofumpt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "gopls": {
        "formatting.gofumpt": true,
    },
}
```

## Links

See [Superchain Upgrades] OP Stack specifications.

[Superchain Upgrades]: https://specs.optimism.io/protocol/superchain-upgrades.html
