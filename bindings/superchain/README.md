# `superchain`

The `superchain` is an optionally `no_std` crate that provides core types
and bindings for the Superchain.

It re-exports two crates:
- [`superchain-primitives`][scp]
- [`superchain-registry`][scr] _Only available if `serde` feature flag is enabled_

[`superchain-primitives`][scp] defines core types used in the `superchain-registry`
along with a few default values for core chains.

[`superchain-registry`][scr] provides bindings to all chains in the `superchain`.

## Usage

Add the following to your `Cargo.toml`.

```toml
[dependencies]
superchain = "0.1"
```

To make make `superchain` `no_std`, toggle `default-features` off like so.

```toml
[dependencies]
superchain = { version = "0.1", default-features = false }
```

## Example

[`superchain-registry`][scr] exposes lazily defined mappings from chain id
to chain configurations that the [`superchain`][sup] re-exports. Below
demonstrates getting the `RollupConfig` for OP Mainnet (Chain ID `10`).

```rust
use superchain::ROLLUP_CONFIGS;

let op_chain_id = 10;
let op_rollup_config = ROLLUP_CONFIGS.get(&op_chain_id);
println!("OP Mainnet Rollup Config: {:?}", op_rollup_config);
```

A mapping from chain id to `ChainConfig` is also available.

```rust
use superchain::OPCHAINS;

let op_chain_id = 10;
let op_chain_config = OPCHAINS.get(&op_chain_id);
println!("OP Mainnet Chain Config: {:?}", op_chain_config);
```

## Feature Flags

- `serde`: Enables [`serde`][s] support for types and makes [`superchain-registry`][scr] types available.
- `std`: Uses the standard library to pull in environment variables.

<!-- Hyperlinks -->

[sp]: ../rust-primitives

[s]: https://crates.io/crates/serde
[sr]: https://github.com/ethereum-optimism/superchain-registry
[scr]: https://crates.io/crates/superchain-registry
[sup]: https://crates.io/crates/superchain
[scp]: https://crates.io/crates/superchain-primitives

