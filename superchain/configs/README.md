# configs

This directory contains configuration source files for OP Stack chains, arranged by superchain. They can be used to configure OP Stack consensus clients (e.g. op-node).

Each chain configuration file in this directory contains configuration data that was generated from the [ops](../../ops/) code.

See also extra data stored per-chain in the [extra](../extra) folder

## Interop dependency sets

All chains in an interop dependency set must settle to the same L1 and share a SuperchainConfig
(same `l1.chain_id` and `superchain_config_addr` in their `superchain.toml`). They do not have to
live in the same superchain dir. `check_depsets` enforces this.

Node network names are derived from the dir (`<chain name>-<superchain dir>`) and embedded in node
binaries at build time, so moving a chain to another dir renames its network and requires new
images. Devnets therefore keep one dir each (e.g. `sepolia-devnet-6`), chosen once and never
renamed, even when they join a dependency set with chains in other dirs.
