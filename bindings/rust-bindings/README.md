# `superchain-registry`

This crate provides Rust bindings for the [Superchain Registry][sr].

[`serde`][s] is a required dependency unlike [`superchain-primitives`][sp].

[`superchain-registry`][src] is an optionally `no_std` crate, by disabling
the `std` feature flag. By default, `std` is enabled, providing standard
library support.

## Usage

Add the following to your `Cargo.toml`.

```toml
[dependencies]
superchain-registry = "0.2"
```

To disable `std` and make `superchain-registry` `no_std` compatible,
simply toggle `default-features` off.

```toml
[dependencies]
superchain-registry = { version = "0.2", default-features = false }
```

## Example

[`superchain-registry`][src] exposes lazily defined mappings from chain id
to chain configurations. Below demonstrates getting the `RollupConfig` for
OP Mainnet (Chain ID `10`).

```rust
use superchain_registry::ROLLUP_CONFIGS;

let op_chain_id = 10;
let op_rollup_config = ROLLUP_CONFIGS.get(&op_chain_id);
println!("OP Mainnet Rollup Config: {:?}", op_rollup_config);
```

A mapping from chain id to `ChainConfig` is also available.

```rust
use superchain_registry::OPCHAINS;

let op_chain_id = 10;
let op_chain_config = OPCHAINS.get(&op_chain_id);
println!("OP Mainnet Chain Config: {:?}", op_chain_config);
```

## Feature Flags

- `std`: Uses the standard library to pull in environment variables.


<!-- Hyperlinks -->

[sp]: ../rust-primitives

[s]: https://crates.io/crates/serde
[sr]: https://github.com/ethereum-optimism/superchain-registry
[scr]: https://crates.io/crates/superchain-registry
