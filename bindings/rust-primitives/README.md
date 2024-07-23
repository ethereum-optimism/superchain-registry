# superchain-primitives

A set of Superchain Primitive Types.

## Usage

Add this to your `Cargo.toml`:

```toml
[dependencies]
superchain-primitives = "0.2"
```

## Example

```rust
use superchain_primitives::rollup_config_from_chain_id;

let op_mainnet_rollup_config = rollup_config_from_chain_id(10).unwrap();
println!("OP Mainnet Rollup Config:\n{op_mainnet_rollup_config:?}");
```

## Feature Flags

- `serde`: Implements serialization and deserialization for types.
- `std`: Uses standard library types.
