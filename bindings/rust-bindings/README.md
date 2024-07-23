# Superchain Registry Rust Bindings

This package provides Rust bindings for the Superchain Registry configuration.

## Usage

Add the following to your `Cargo.toml`.

```toml
[dependencies]
superchain-registry = "0.2"
```

## Feature Flags

- `std`: Uses the standard library to pull in environment variables.

## Example

```rust
use superchain_registry::ROLLUP_CONFIGS;

let op_chain_id = 10;
let op_rollup_config = ROLLUP_CONFIGS.get(&op_chain_id);
println!("OP Mainnet Rollup Config: {:?}", op_rollup_config);

for (id, _) in ROLLUP_CONFIGS.iter() {
    println!("Loaded rollup config for chain ID {}", id);
}
```
