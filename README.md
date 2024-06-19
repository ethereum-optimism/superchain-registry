# superchain-registry

> [!WARNING]
> This repository is a **work in progress**.  At a later date, it will be proposed to, and must be approved by, Optimism Governance.  Until that time, the configuration described here is subject to change.

> [!IMPORTANT]
> We're making some changes to this repository and we've paused adding new chains for now. We'll reopen this process once the repository is ready. The Superchain itself, of course, remains open for business.

The Superchain Registry repository hosts Superchain-configuration data in a minimal human-readable form.

This includes mainnet and testnet Superchain targets, and their respective member chains.

Other configuration, such as contract-permissions and `SystemConfig` parameters are hosted and governed onchain.

A list of chains in the registry can be seen in the top level [`chainList.toml`](./chainList.toml) and [`chainList.json`](./chainList.json) files.

## Downstream packages
The superchain configs are stored in a minimal form, and are embedded in downstream OP-Stack software ([`op-node`](https://github.com/ethereum-optimism/optimism) and [`op-geth`](https://github.com/ethereum-optimism/op-geth)).

Full deployment artifacts and genesis-states can be derived from the minimal form
using the reference [`op-chain-ops`] tooling.

[`op-chain-ops`]: https://github.com/ethereum-optimism/optimism/tree/develop/op-chain-ops

## Adding a Chain

The following are the steps you need to take to add a chain to the registry:

### 0. Install dependencies
You will need [`jq`](https://jqlang.github.io/jq/download/) and [`foundry`](https://book.getfoundry.sh/getting-started/installation) installed, as well as Go.

### 1. Set env vars

To contribute a standard OP-Stack chain configuration, the following data is required: contracts deployment, rollup config, L2 genesis. We provide a tool to scrape this information from your local [monorepo](https://github.com/ethereum-optimism/optimism) folder.

First, make a copy of `.env.example` named `.env`, and alter the variables to appropriate values.

### 2. Run script

#### Frontier chains

Frontier chains are chains with customizations beyond the standard OP
Stack configuration. To contribute a frontier OP-Stack chain
configuration, you can run:


```shell
sh scripts/add-chain.sh frontier
```

#### Standard chains
A chain may meet the definition of a **standard** chain. Adding a standard chain is a two-step process.

First, the chain should be added as a frontier chain but with `--standard-chain-candidate` flag set:

```shell
sh scripts/add-chain.sh standard --standard-chain-candidate
```

The remaining steps should then be followed to merge the config data into the registry -- a prerequisite for [promoting the chain](#promote-chain) to a standard chain.

### 3. Understand output
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

### 4. Run tests locally

Run the following command from the `validation` folder to run the Go validation checks, for only the chain you added (replace the chain name or ID accordingly):
```
go test -run=/OP-Sepolia
```
or
```
go test -run=/11155420
```
You can even focus on a particular test and chain combination:
```
go test -run=TestGasPriceOracleParams/11155420
```
Omit the `-run=` flag to run checks for all chains.

> [!NOTE]
> If you set `--standard-chain-candidate`, your chain will be checked against the majority of the standard configuration requirements. These are defined in the [specs](https://specs.optimism.io/protocol/configurability.html). However, these requirements are currently a draft, pending governance approval.
>
> The final requirement to a standard chain concerns the `ProxyAdminOwner`. That check will not be run until the chain is promoted to standard.

### 5. Run codegen and check output
This is a tool which will rewrite certain summary files of all the chains in the registry, including the one you are adding. The output will be checked in a continuous integration checks (it is required to pass):

```
sh scripts/codegen.sh
```

> [!NOTE]
> Please double check the diff to this file. This data may be consumed by external services, e.g. wallets. If anything looks incorrect, please get in touch.

### 6. Open Your Pull Request
When opening a PR:
- Open it from a non-protected branch in your fork (e.g. avoid the `main` branch). This allows maintainers to push to your branch if needed, which streamlines the review and merge process.
- Open one PR per chain you would like to add. This ensures the merge of one chain is not blocked by unexpected issues.

Once the PR is opened, the same automated checks from Step 4 will then run on your PR, and your PR will be reviewed in due course. Once these checks pass the PR will be merged.

<a name="promote-chain"></a>
## Promote a chain to standard
This process is only possible for chains already in the registry.

Run this command:
```
sh scripts/promote-to-standard.sh <chain-id>
```

This command will:
* activate the full suite of validation checks for standard chains, including checks on the `ProxyAdminOwner`
* declare the chain as a standard chain
* set the `superchain_time`, so that the chain receives future hardforks with the rest of the superchain (baked into downstream OPStack software).

## License

MIT License, see [`LICENSE` file](./LICENSE).
