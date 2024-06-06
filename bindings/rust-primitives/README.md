# superchain-primitives

A set of Superchain Primitive Types.

## Usage

Add this to your `Cargo.toml`:

```toml
[dependencies]
superchain-primitives = "0.1.0"
```

## Example

```rust
use alloy_primitives::b256;
use superchain_primitives::BlockID;

let block_id = BlockID {
    hash: b256!("0000000000000000000000000000000000000000000000000000000000000000"),
    number: 0u64,
};
println!("Block ID: {block_id}");
```

## Feature Flags

- `serde`: Implements serialization and deserialization for types.
- `std`: Uses standard library types.
