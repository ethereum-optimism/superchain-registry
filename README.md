# superchain-registry

> :warning: This repository is a **work in progress**.  At a later date, it will be proposed to, and must be approved by, Optimism Governance.  Until that time, the configuration described here is subject to change.
> 
> :no_entry: We're making some changes to this repository and we've paused adding new chains for now. We'll reopen this process once the repository is ready. The Superchain itself, of course, remains open for business.

The Superchain Registry repository hosts Superchain-configuration data in a minimal human-readable form.
This includes mainnet and testnet Superchain targets, and their respective member chains.

Other configuration, such as contract-permissions and `SystemConfig` parameters are hosted and governed onchain.

The superchain configs are made available in minimal form, to embed in OP-Stack software.
Full deployment artifacts and genesis-states can be derived from the minimal form
using the reference [`op-chain-ops`] tooling.

The `semver.yaml` files each represent the semantic versioning lockfile for the all of the smart contracts in that superchain.
It is meant to be used when building transactions that upgrade the implementations set in the proxies.

## Adding a Chain

The following are the steps you need to take to add a chain to the 

### 0. Install dependencies
You will need [`jq`](https://jqlang.github.io/jq/download/) and [`foundry`](https://book.getfoundry.sh/getting-started/installation) installed, as well as Go.

### 1. Set env vars

To contribute a standard OP-Stack chain configuration, the following data is required: contracts deployment, rollup config, L2 genesis. We provide a tool to scrape this information from your local [monorepo](https://github.com/ethereum-optimism/optimism) folder.

> [!NOTE]
> The standard configuration requirements are defined in the [specs](https://specs.optimism.io/protocol/configurability.html). However, these requirements are currently a draft, pending governance approval.

First, make a copy of `.env.example` named `.env`, and alter the variables to appropriate values.

### 2. Run script
#### Standard chains
If your chain meets the definition of a **standard** chain, you can run:


```shell
sh scripts/add-chain.sh standard
```

#### Frontier chains

Frontier chains are chains with customizations beyond the standard OP
Stack configuration. To contribute a frontier OP-Stack chain
configuration, you can run:


```shell
sh scripts/add-chain.sh frontier
```

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
There are currently two sets of validation checks:

#### Go validation checks
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

#### Solidity validation checks
Run
```shell
sh ./scripts/check-security-configs.sh
```
from the repository root to run checks for all chains.

### 5. Open Your Pull Request
When opening a PR:
- Open it from a non-protected branch in your fork (e.g. avoid the `main` branch). This allows maintainers to push to your branch if needed, which streamlines the review and merge process.
- Open one PR per chain you would like to add. This ensures the merge of one chain is not blocked by unexpected issues.

Once the PR is opened, the same automated checks from Step 4 will then run on your PR, and your PR will be reviewed in due course. Once these checks pass the PR will be merged.

## Adding a superchain target
A superchain target defines a set of layer 2 chains which share a `SuperchainConfig` and `ProtocolVersions` contract deployment on layer 1. It is usually named after the layer 1 chain, possibly with an extra identifier to distinguish devnets.


> **Note**
> Example: `sepolia` and `sepolia-dev-0` are distinct superchain targets, although they are on the same layer 1 chain.


A new Superchain Target can be added by creating a new superchain config directory,
with a `superchain.yaml` config file.

> **Note**
> This is an infrequent operation and unecessary if you are just looking to add a chain to an existing superchain.

Here's an example:

```bash
cd superchain-registry

export SUPERCHAIN_TARGET=goerli-dev-0
mkdir superchain/configs/$SUPERCHAIN_TARGET

cat > superchain/configs/$SUPERCHAIN_TARGET/superchain.yaml << EOF
name: Goerli Dev 0
l1:
  chain_id: 5
  public_rpc: https://ethereum-goerli-rpc.allthatnode.com
  explorer: https://goerli.etherscan.io

protocol_versions_addr: null # todo
superchain_config_addr: null # todo
EOF
```
Superchain-wide configuration, like the `ProtocolVersions` contract address, should be configured here when available.

### Approved contract versions
Each superchain target should have a `semver.yaml` file in the same directory declaring the approved contract semantic versions for that superchain, e.g:
```yaml
l1_cross_domain_messenger: 1.4.0
l1_erc721_bridge: 1.0.0
l1_standard_bridge: 1.1.0
l2_output_oracle: 1.3.0
optimism_mintable_erc20_factory: 1.1.0
optimism_portal: 1.6.0
system_config: 1.3.0

# superchain-wide contracts
protocol_versions: 1.0.0
superchain_config:
```

### `implementations`
Per superchain a set of canonical implementation deployments, per semver version, is tracked.
As default, an empty collection of deployments can be set:
```bash
cat > superchain/implementations/networks/$SUPERCHAIN_TARGET.yaml << EOF
l1_cross_domain_messenger:
l1_erc721_bridge:
l1_standard_bridge:
l2_output_oracle:
optimism_mintable_erc20_factory:
optimism_portal:
system_config:
EOF
```

## Setting up your editor for formatting and linting
If you use VSCode, you can place the following in a `settings.json` file in the gitignored `.vscode` directory:

```json
{
    "go.formatTool": "gofumpt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "gopls": {
        "formatting.gofumpt": true,
    },
}
```

## License

MIT License, see [`LICENSE` file](./LICENSE).
