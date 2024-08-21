# `superchain-primitives`

A set of primitive types for the superchain.
These types mirror the golang types defined by the `superchain-registry`.

[`superchain-primitives`][sp] is a `no_std` [crate][rpc] with optional type support for
[`serde`][serde] serialization and deserialization providing a `serde` feature flag.

Standard library support is available by enabling the `std` feature flag on the
[`superchain-primitives`][sp] dependency.

By default, both the `std` and `serde` feature flags **are** enabled.

## Usage

Add the following to your `Cargo.toml`.

```toml
[dependencies]
superchain-primitives = "0.2"
```

To disable default feature flags, disable the `default-features` field like so.

```toml
superchain-primitives = { version = "0.2", default-features = false }
```

Features can then be enabled individually.

```toml
superchain-primitives = { version = "0.2", default-features = false, features = [ "std" ] }
```

## Example

Below uses statically defined rollup configs for common chain ids.

```rust
use superchain_primitives::rollup_config_from_chain_id;

let op_mainnet_rollup_config = rollup_config_from_chain_id(10).unwrap();
println!("OP Mainnet Rollup Config:\n{op_mainnet_rollup_config:?}");
```

To inherit rollup configs defined by the `superchain-registry`,
use the `superchain-registry` crate defined in [rust-bindings][rb].
Note, `serde` is required.

## Feature Flags

- `serde`: Implements serialization and deserialization for types.
- `std`: Uses standard library types.

## Hardcoded Rollup Configs

- [`OP_MAINNET_CONFIG`][rcfg]: OP Mainnet (Chain ID: `10`)
- [`OP_SEPOLIA_CONFIG`][rcfg]: OP Sepolia (Chain ID: `11155420`)
- [`BASE_MAINNET_CONFIG`][rcfg]: Base Mainnet (Chain ID: `8453`)
- [`BASE_SEPOLIA_CONFIG`][rcfg]: Base Sepolia (Chain ID: `84532`)

<!-- Hyperlinks -->

[rb]: ../rust-bindings
[rpc]: ./Cargo.toml
[rcfg]: ./src/rollup_config.rs

[serde]: https://crates.io/crates/serde
[sp]: https://crates.io/crates/superchain-primitives


