# configs

This directory contains configuration source files for OP Stack chains, arranged by superchain. They can be used to configure OP Stack consensus clients (e.g. op-node).

Each chain configuration file in this directory contains configuration data that was generated from the [ops](../../ops/) code.

See also extra data stored per-chain in the [extra](../extra) folder

## Field lifecycle

Not every field in a chain config changes in the same way over the chain's life. Each
field falls into one of three lifecycles, classified by the `lifecycle` struct tag on
the `Chain` type in [`ops/internal/config/chain.go`](../../ops/internal/config/chain.go):

| Lifecycle | Meaning | Examples |
| --- | --- | --- |
| **immutable** | Fixed when the chain is created and can never change. These values are determined by the chain's genesis; changing one describes a *different* chain. | `chain_id`, `block_time`, `seq_window_size`, `max_sequencer_drift`, `gas_paying_token`, the entire `[genesis]` block (including `[genesis.system_config]`) |
| **append-only** | Grows over time as the protocol evolves. New entries may be added, and a not-yet-active entry may still be re-scheduled, but an entry whose activation is already in the past is frozen. | `[hardforks]` activation times — a new hardfork may be appended and an upcoming activation may be pushed out, but a hardfork that has **already activated** keeps its timestamp forever (it is on-chain history) |
| **mutable** | Tracks live on-chain or operational state and may be updated freely. | `[roles]` and `[addresses]` (key rotations, contract upgrades), `public_rpc`/`sequencer_rpc`/`explorer`, `superchain_level`, `data_availability_type` |

This contract is **enforced in CI**: `TestImmutableFieldsUnchanged` in
[`ops/internal/manage`](../../ops/internal/manage) compares every changed config in a
pull request against its committed version and fails if an immutable field changed or
an existing hardfork activation was altered. The check is driven by the `lifecycle`
tags, so the source of truth is the struct definition. When adding a new field,
classify it conservatively — prefer `mutable` unless a change genuinely cannot happen
without producing a different chain.
