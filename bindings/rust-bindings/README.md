# Superchain Registry Rust Bindings

This package provides Rust bindings for the Superchain Registry configuration.

## Usage

Add this to your `Cargo.toml`:

```toml
[dependencies]
superchain-registry = "0.1.0"
```

## Feature Flags

- `std`: Uses the standard library to pull in environment variables.

## Example

```rust
use superchain_registry::ROLLUP_CONFIGS;

let rollups = ROLLUP_CONFIGS.clone();

for (id, config) in rollups.iter() {
    println!("Loaded rollup config for chain ID {}", id);
}
```
