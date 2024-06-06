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
use superchain_registry::SUPERCHAINS;

let superchains = SUPERCHAINS.clone();

for (name, superchain) in superchains.iter() {
    println!("Loaded Superchain: {} with {} chain IDs", name, superchain.chain_ids.len());
}
```
