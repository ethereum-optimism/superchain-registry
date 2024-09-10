# Adding a Chain

The following are the steps you need to take to add a chain to the registry:

> [!IMPORTANT]
> Ensure your chain is listed at [ethereum-lists/chains](https://github.com/ethereum-lists/chains).
> This is to ensure your chain has a unique chain ID. Our validation suite will
> check against this repository. This is a mandatory prerequisite before being
> added to the registy.

### 0. Fork this repository
You will be raising a Pull Request from your fork to the upstream repo.

We recommend only adding one chain at a time, and starting with a fresh branch of this repo for every chain. 

### 1. Install dependencies

Install the following dependencies

| Dependency                                                            | Version   | Version Check Command |
| --------------------------------------------------------------------- | --------- | --------------------- |
| [git](https://git-scm.com/)                                           | `^2`      | `git --version`       |
| [go](https://go.dev/)                                                 | `^1.21`   | `go version`          |
| [foundry](https://github.com/foundry-rs/foundry#installation)         | `^0.2.0`  | `forge --version`     |
| [jq](https://github.com/jqlang/jq)                                    | `^1.6`    | `jq --version`        |
| [just](https://github.com/casey/just?tab=readme-ov-file#installation) | `^1.28.0` | `just --version`      |

### 2. Set environment variables

Make a copy of `.env.example` named `.env`, and alter the variables to appropriate values. Each value is explained in a comment in `.env.example`.

> [!IMPORTANT]
> Adding a standard chain is a two-step process. For your initial PR, ensure
> `SCR_STANDARD_CHAIN_CANDIDATE=true` in the `.env` file. After you've met all
> of the [Standard chain requirements](./glossary.md), you'll open a new PR following the
> [Promote a chain to standard](#promote-a-chain-to-standard) process.

### 3. Run script

Run the following script from the root of the repository to open the PR.

```shell
just add-chain
```

The remaining steps should then be followed to merge the config data into the registry -- a prerequisite for [promoting the chain](#promote-a-chain-to-standard) to a standard chain.

### 4. Understand output
The tool will write the following data:
- The main configuration source, with genesis data, and address of onchain system configuration. These are written to `uperchain/configs/<superchain-target>/<chain-short-name>.toml`.
- Hardfork override times, where they have been set, will be included. If and when a chain becomes a standard chain, a `superchain_time` is set in the chain config. From that time on, future hardfork activation times which are missing from the chain config will be inherited from superchain-wide values in the neighboring `superchain.toml` file.
- Genesis system config data
- Compressed `genesis.json` definitions (in the `extra/genesis` directory) which pull in the bytecode by hash

The genesis largely consists of contracts common with other chains:
all contract bytecode is deduplicated and hosted in the `extra/bytecodes` directory.

The format is a gzipped JSON `genesis.json` file, with either:
- a `alloc` attribute, structured like a standard `genesis.json`, but with `codeHash` (bytes32, `keccak256` hash of contract code) attribute per account, instead of the `code` attribute seen in standard Ethereum genesis definitions.
- a `stateHash` attribute: to omit a large state (e.g. for networks with a re-genesis or migration history). Nodes can load the genesis block header, and state-sync to complete the node initialization.

### 5. Run tests locally

Run the following command to run the Go validation checks, for only the chain you added (replace the `<chain-id>` accordingly):
```
just validate <chain-id>
```

> [!NOTE]
> If you set `SCR_STANDARD_CHAIN_CANDIDATE`, your chain will be checked against the majority of the standard rollup requirements as outlined by the [Standard Rollup Blockspace Charter](https://gov.optimism.io/t/season-6-draft-standard-rollup-charter/8135) (currently a draft pending Governance approval).
>
> The final requirement to qualify as a standard chain concerns the `ProxyAdminOwner` role. The validation check for this role  will not be run until the chain is [promoted](#promote-a-chain-to-standard) to standard.

The [`validation_test.go`](./validation/validation_test.go) test declaration file defines which checks run on each class of chain. The parameters referenced in each check are recorded in [TOML files in the `standard` directory](./validation/standard).

### 6. Run codegen and check output

This tool will add your chain to the `chainList.toml` and `addresses.json` files, which contain summaries of all chains in the registry.

```
just codegen
```

> [!NOTE]
> Please double check the diff to this file.
> This data may be consumed by external services, e.g. wallets.
> If anything looks incorrect, please get in touch.

### 7. Open Your Pull Request
When opening a PR:
- Open it from a non-protected branch in your fork (e.g. avoid the `main` branch). This allows maintainers to push to your branch if needed, which streamlines the review and merge process.
- Open one PR per chain you would like to add. This ensures the merge of one chain is not blocked by unexpected issues.
- Once the PR is opened, please check the box to [allow edits from maintainers](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork). 

Once the PR is opened, the same automated checks you've run locally will then run on your PR, and your PR will be reviewed in due course. Once these checks pass, the PR will be merged.


## Promote a chain to standard
This process is only possible for chains already in the registry.

Run this command (replace the `<chain-id>` accordingly):
```
just promote-to-standard <chain-id>
```

This command will:
* declare the chain as a standard chain
* set the `superchain_time`, so that the chain receives future hardforks with the rest of the superchain (baked into downstream OPStack software, selected with [network flags](https://docs.optimism.io/builders/node-operators/configuration/base-config#initialization-via-network-flags)).
* activate the full suite of validation checks for standard chains, including checks on the `ProxyAdminOwner`