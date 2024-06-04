# Superchain Registry Rust Bindings

This package provides Rust bindings for the Superchain Registry configuration.

## Usage

Add this to your `Cargo.toml`:

```toml
[dependencies]
superchain_registry = "0.1.0"
```

## Example

```rust
use superchain_registry::SUPERCHAINS;

fn main() {
    let superchains = SUPERCHAINS.clone();

    for superchain in superchains.iter() {
        println!("Loaded Superchain: {}", superchain.name);
    }
}
```
