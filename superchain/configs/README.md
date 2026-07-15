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
| **immutable** | Fixed when the chain is created and describes the chain itself; changing one normally means the file now describes a *different* chain. | `chain_id`, `batch_inbox_addr`, `block_time`, `seq_window_size`, `max_sequencer_drift`, `gas_paying_token`, `data_availability_type`, the entire `[genesis]` block (including `[genesis.system_config]`) |
| **append-only** | Grows over time as the protocol evolves. New entries may be added, and a not-yet-active entry may still be re-scheduled, but an entry whose activation is already in the past is frozen. | `[hardforks]` activation times — a new hardfork may be appended and an upcoming activation may be pushed out, but a hardfork that has **already activated** keeps its timestamp forever (it is on-chain history). A non-timestamp modifier such as `keep_karst_upgrade_gas` records how its hardfork activated, so it is adjustable until that hardfork is in the past and frozen thereafter. |
| **mutable** | Tracks live on-chain or operational state and may be updated freely. | `[roles]` and `[addresses]` (key rotations, contract upgrades), `public_rpc`/`sequencer_rpc`/`explorer`, `superchain_level`, `[optimism]` |

This contract is **checked in CI, but the check is advisory**: the `check-immutable-fields`
job (`ops/cmd/check_immutable_fields`, driven by the `lifecycle` tags on the `Chain`
struct) compares every changed config in a pull request against its committed version and
**warns** when an immutable field changed or a past hardfork activation was altered. It is
deliberately **not a required status check**, so a legitimate clean-up — for example
correcting a value that was committed wrong — is not blocked; a reviewer sees the warning
and confirms the change is intentional.

Classify a new field by what it genuinely represents, not by whether it has ever been
edited. Because the check only warns, a field can be `immutable` even though a rare
correction to it is sometimes needed.
