# Bindings

Bindings bridge the `superchain-registry` to other languages with associated types.

## Rust Bindings

There are two rust [crates](crates.io) that expose rust bindings.
- [`bindings/rust-primitives`][rp]
- [`bindings/rust-bindings`][rb]

### Rust Primitives

[`bindings/rust-primitives`][rp] contains rust types for the objects defined in the `superchain-registry`.
The `rust-primitives` directory is a `no_std` [crate][rpc] called [`superchain-primitives`][sp].
It defines the `std` and `serde` feature flags that enable `std` library use and `serde` serialization
and deserialization support, respectively. Enabling these feature flags on the crate are as simple as
adding the `features` key to the dependency declaration. For example, in a project's `Cargo.toml`, to
use the `superchain-primitives` crate with `serde` support, add the following to the `[dependencies]`
section.

```toml
# By default, superchain-primitives enables the `std` and `serde` feature flags.
superchain-primitives = { version = "0.2", features = [ "serde" ], default-features = false }
```

[`superchain-primitives`][sp] has minimal dependencies itself and uses [alloy][alloy] Ethereum
types internally. Below provides a breakdown of some core types defined in [`superchain-primitives`][sp].

**`ChainConfig`**

The [`ChainConfig`][cc] is the chain configuration that is output from the `add-chain` command in the
`superchain-registry`. It contains genesis data, addresses for the onchain system config, hardfork
timestamps, rpc information, and other superchain-specific information. Static chain config files
are defined in [../superchain/configs/](../superchain/configs/).

**`RollupConfig`**

The [`RollupConfig`][rc], often confused with the `ChainConfig`, defines the configuration for
a rollup node. The [`RollupConfig`][rc] defines the parameters used for deriving the L2 chain as
well as batch-posting data on the flip side. It effectively defines the rules for the chain used by
consensus and execution clients. (e.g. `op-node`, `op-reth`, `op-geth`, ...)

[`superchain-primitives`][sp] also exposes a few default `RollupConfig`s for convenience, providing an alternative to depending on [`superchain-registry`][sr] with `serde` required.

**`Superchain`**

[`Superchain`][s] defines a superchain for a given L1 network. It holds metadata such as the
name of the superchain, the L1 anchor information (chain id, rpc, explorer), and default
hardfork configuration values. Within the [`Superchain`][s], there's a list of [`ChainConfig`][cc]s
that belong to this superchain.

### Rust Bindings

[`bindings/rust-bindings`][rb] exports rust type definitions for chains in the `superchain-registry`.
The `rust-bindings` directory is a `no_std` [crate][rbc] called [`superchain-registry`][sr].
It requires `serde` and enables `serde` features on dependencies including [`superchain-primitives`][sp],
which it depends on for types. To use the `superchain-regsitry` crate, add the crate as a dependency to
a `Cargo.toml`.

```toml
# By default, superchain-registry enables the `std` feature, disabling `no_std`.
superchain-registry = { version = "0.2", default-features = false }
```

[`superchain-registry`][sr] declares lazy evaluated statics that expose `ChainConfig`s, `RollupConfig`s,
and `Chain` objects for all chains with static definitions in the superchain registry. The way this works
is the the golang side of the superchain registry contains an "internal code generation" script that has
been modified to output configuration files to the [`bindings/rust-bindings`][rb] directory in the `etc`
folder that are read by the [`superchain-registry`][sr] rust crate. These static config files contain
an up-to-date list of all superchain configurations with their chain configs.

There are three core statics exposed by the [`superchain-registry`][sr].
- `CHAINS`: A list of chain objects containing the superchain metadata for this chain.
- `OPCHAINS`: A map from chain id to `ChainConfig`.
- `ROLLUP_CONFIGS`: A map from chain id to `RollupConfig`.

Where the [`superchain-primitives`][sp] crate contains a few hardcoded `RollupConfig` objects, the
[`superchain-registry`][sr] exports the _complete_ list of superchains and their chain's `RollupConfig`s
and `ChainConfig`s, at the expense of requiring `serde`.

[`CHAINS`][chains], [`OPCHAINS`][opchains], and [`ROLLUP_CONFIGS`][rollups] are exported at the top-level
of the [`superchain-primitives`][sp] crate and can be used directly. For example, to get the rollup
config for OP Mainnet, one can use the [`ROLLUP_CONFIGS`][rollups] map to get the `RollupConfig` object.

```rust
use superchain_registry::ROLLUP_CONFIGS;

let op_chain_id = 10;
let op_mainnet_rcfg: &RollupConfig = ROLLUP_CONFIGS.get(&op_chain_id).unwrap();
assert_eq!(op_mainnet_rcfg.chain_id, op_chain_id);
```

# Acknowledgements

[Alloy][alloy] for creating and maintaining high quality Ethereum types in rust.


<!-- Hyperlinks -->

[rbc]: ../bindings/rust-bindings/Cargo.toml
[rb]: ../bindings/rust-bindings/
[rpc]: ../bindings/rust-primitives/Cargo.toml
[rp]: ../bindings/rust-primitives/

[alloy]: https://github.com/alloy-rs/alloy

[sr]: https://crates.io/crates/superchain-registry
[sp]: https://crates.io/crates/superchain-primitives

[chains]: https://docs.rs/superchain-registry/latest/superchain_registry/struct.CHAINS.html
[opchains]: https://docs.rs/superchain-registry/latest/superchain_registry/struct.OPCHAINS.html
[rollups]: https://docs.rs/superchain-registry/latest/superchain_registry/struct.ROLLUP_CONFIGS.html

[s]: https://docs.rs/superchain-primitives/latest/superchain_primitives/superchain/struct.Superchain.html
[cc]: https://docs.rs/superchain-primitives/latest/superchain_primitives/chain_config/struct.ChainConfig.html
[rc]: https://docs.rs/superchain-primitives/latest/superchain_primitives/rollup_config/struct.RollupConfig.html
