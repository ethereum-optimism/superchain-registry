## Hardforks

Each chain config (`superchain/configs/<superchain>/<chain>.toml`) carries a `[hardforks]` table with the activation
times of the OP Stack network upgrades. Every `*_time` field is the L2 block timestamp (Unix seconds, UTC) at which that
fork activates; an absent field means it is not scheduled, and `0` means active from genesis.

Activation times can be inherited superchain-wide from the superchain's `superchain.toml` when a chain sets
`superchain_time`, and are then written into each chain's own config. See
[Hardfork activation inheritance](../docs/hardfork-activation-inheritance.md) for the exact rules.

Two optional `[hardforks]` fields are not forks themselves but toggle a corrective behavior tied to one:

- **`pectra_blob_schedule_time`** (timestamp, optional) — the Pectra blob-schedule fix: the L1-origin timestamp until
  which (exclusive) the L1 block info's blob base-fee calculation keeps using the pre-Prague (Cancun) blob parameters,
  switching to the Prague (Pectra) parameters from then on. It is needed by chains that were running buggy node software
  when Prague activated on their L1: it records the L1 timestamp at which they actually switched, so re-derivation
  reproduces their canonical history. As a timestamp field it participates in superchain inheritance. See the
  [Pectra blob schedule spec](https://specs.optimism.io/protocol/pectra-blob-schedule/derivation.html) for details.

- **`keep_karst_upgrade_gas`** (bool, optional, default `false`) — opts a chain out of the Karst upgrade-gas fix.
  Karst's activation block adds one-time "upgrade gas" to the block gas limit so the network-upgrade transactions fit,
  and the consensus-layer software is meant to subtract it again in the post-activation block so the limit reverts to
  its steady-state value. Some chains activated Karst with a bug in that software that missed the subtraction, leaving
  them with a permanently inflated gas limit; they set this to `true` so the fixed software keeps reproducing that
  inflated limit and their history stays valid (the operator can lower it later with a `setGasLimit` transaction).
  Unlike the timestamp fields, this flag is set per-chain and is never inherited from the superchain.

## Compression

OP Stack genesis files are typically around 9MB, which quickly bloats the binary size of applications that use the
`validation` package. As a result, genesis files are stored in this repository as compressed `.json.zst` files.

### Benchmarks

We investigated a number of different compression methods, including:

1. No compression
2. Gzipping each individual genesis file
3. Manually extracting each contract's bytecode into individual gzipped files, then referencing them in each
   genesis file by their hashes
4. Using zstd with a pre-trained dictionary

| Compression Method | Size   | Compression Ratio |
|--------------------|--------|-------------------|
| None               | 326M   | 0                 |
| gzip               | 5.39MB | 60:1              |
| Manual             | 2.1MB  | 155:1             |
| zstd               | 2.36MB | 138:1             |

While manual compression is most effective, it also requires custom serialization code. Therefore, we opted for zstd
to optimize for small size while still allowing the genesis files to be compressed and decompressed without changes to
their structure.

### Generating the Dictionary

The dictionary is stored in the `extra/dictionary` file. To generate it:

1. Use the `dump_genesis` tool to output full genesis JSON files in a directory.
2. Run `zstd --train <your-dir>/*.json`. This will output a `dictionary` file in the current directory.
3. Move the resulting dictionary to `extra/dictionary`.
