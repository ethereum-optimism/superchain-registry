# superchain-registry

> [!WARNING]
> This repository is a **work in progress**.  At a later date, it will be proposed to, and must be approved by, Optimism Governance.  Until that time, the configuration described here is subject to change.

> [!IMPORTANT]
> We're making some changes to this repository and we've paused adding new chains for now. We'll reopen this process once the repository is ready. The Superchain itself, of course, remains open for business.

The Superchain Registry repository hosts Superchain-configuration data in a minimal human-readable form.

This includes mainnet and testnet Superchain targets, and their respective member chains.

Other configuration, such as contract-permissions and `SystemConfig` parameters are hosted and governed onchain.

A list of chains in the registry can be seen in the top level [`chainList.toml`](./chainList.toml) and [`chainList.json`](./chainList.json) files.

## Superchain Level and Rollup Stage
Chains in the superchain-registry are assigned a `superchain_level` (shown in individual config files as well as the `chainList.tom/json` summaries), depending on the set of validation checks that they pass.

**Frontier** chains have `superchain_level = 1`.

**Standard** chains have `superchain_level = 2`. Because they satisfy a much stricter set of validation checks (see the [Standard Rollup Blockspace Charter](https://gov.optimism.io/t/season-6-draft-standard-rollup-charter/8135)), they also qualify as [Stage 1](https://ethereum-magicians.org/t/proposed-milestones-for-rollups-taking-off-training-wheels/11571) rollups chains.


## Downstream packages
The superchain configs are stored in a minimal form, and are embedded in downstream OP-Stack software ([`op-node`](https://github.com/ethereum-optimism/optimism) and [`op-geth`](https://github.com/ethereum-optimism/op-geth)). This means that, after a chain has been added to the registry, and the dependency on the registry updates in the downstream software, it is possible to start an `op-node` instance [using the `--network` flag](https://docs.optimism.io/builders/node-operators/configuration/consensus-config#network) (and also an `op-geth` instance [using the `--op-network` tag](https://docs.optimism.io/builders/node-operators/configuration/execution-config#op-network-betaop-network)) which will successfully sync with other nodes on that network.

## Adding a Chain

The following are the steps you need to take to add a chain to the registry:
### 0. Fork this repository
You will be raising a Pull Request from your fork to the upstream repo.

### 1. Ensure your chain is listed at ethereum-lists/chains
This is to ensure your chain has a unique chain ID. Our validation suite will check your chain against https://github.com/ethereum-lists/chains.


### 2. Install dependencies
You will need [`jq`](https://jqlang.github.io/jq/download/) and [`foundry`](https://book.getfoundry.sh/getting-started/installation) installed, as well as Go and [`just`](https://just.systems/man/en/).

### 3. Set env vars

To contribute a standard OP-Stack chain configuration, in addition to user-supplied metadata (chain name) the following data is required: contracts deployment, rollup config, L2 genesis. We provide a tool to scrape this information from your local [monorepo](https://github.com/ethereum-optimism/optimism) folder.

First, make a copy of `.env.example` named `.env`, and alter the variables to appropriate values.
#### Frontier chains
To contribute a frontier OP-Stack chain configuration, you set the `SCR_CHAIN_TYPE=frontier` in the `.env` file. This will set `superchain_level = 1` in the registry's config file for this chain.

#### Standard chains
Adding a standard chain is a two-step process. First, the chain should be added as a frontier chain as above, but with `SCR_STANDARD_CHAIN_CANDIDATE=true` in the `.env` file.

### 4. Run script

```shell
just add-chain
```

The remaining steps should then be followed to merge the config data into the registry -- a prerequisite for [promoting the chain](#promote-a-chain-to-standard) to a standard chain.

### 5. Understand output
The tool will write the following data:
- The main configuration source, with genesis data, and address of onchain system configuration. These are written to `superchain/configs/superchain_target/chain_short_name.yaml`.
> **Note**
> Hardfork override times, where they have been set, will be included. If and when a chain becomes a standard chain, a `superchain_time` is set in the chain config. From that time on, future hardfork activation times which are missing from the chain config will be inherited from superchain-wide values in the neighboring `superchain.yaml` file.

- Addresses of L1 contracts. (Note that all L2 addresses are statically known addresses defined in the OP-Stack specification, and thus not configured per chain.) These are written to `extra/addresses/superchain_target/chain_short_name.json`.
- Genesis system config data
- Compressed `genesis.json` definitions (in the `extra/genesis` directory) which pull in the bytecode by hash

The genesis largely consists of contracts common with other chains:
all contract bytecode is deduplicated and hosted in the `extra/bytecodes` directory.

The format is a gzipped JSON `genesis.json` file, with either:
- a `alloc` attribute, structured like a standard `genesis.json`,
  but with `codeHash` (bytes32, `keccak256` hash of contract code) attribute per account,
  instead of the `code` attribute seen in standard Ethereum genesis definitions.
- a `stateHash` attribute: to omit a large state (e.g. for networks with a re-genesis or migration history).
  Nodes can load the genesis block header, and state-sync to complete the node initialization.

### 6. Run tests locally

Run the following command to run the Go validation checks, for only the chain you added (replace the `<chain-id>` accordingly):
```
just validate <chain-id>
```

> [!NOTE]
> If you set `SCR_STANDARD_CHAIN_CANDIDATE`, your chain will be checked against the majority of the standard rollup requirements as outlined by the [Standard Rollup Blockspace Charter](https://gov.optimism.io/t/season-6-draft-standard-rollup-charter/8135) (currently a draft pending Governance approval).
>
> The final requirement to qualify as a standard chain concerns the `ProxyAdminOwner` role. The validation check for this role  will not be run until the chain is [promoted](#promote-a-chain-to-standard) to standard.

The [`validation_test.go`](./validation/validation_test.go) test declaration file defines which checks run on each class of chain. The parameters referenced in each check are recorded in [TOML files in the `standard` directory](./validation/standard).

### 7. Run codegen and check output
This is a tool which will rewrite certain summary files of all the chains in the registry, including the one you are adding. The output will be checked in a continuous integration checks (it is required to pass):

```
just codegen
```

> [!NOTE]
> Please double check the diff to this file. This data may be consumed by external services, e.g. wallets. If anything looks incorrect, please get in touch.

### 8. Open Your Pull Request
When opening a PR:
- Open it from a non-protected branch in your fork (e.g. avoid the `main` branch). This allows maintainers to push to your branch if needed, which streamlines the review and merge process.
- Open one PR per chain you would like to add. This ensures the merge of one chain is not blocked by unexpected issues.

Once the PR is opened, the same automated checks from Step 4 will then run on your PR, and your PR will be reviewed in due course. Once these checks pass the PR will be merged.


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

## License

MIT License, see [`LICENSE` file](./LICENSE).
