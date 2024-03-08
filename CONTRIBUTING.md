# Superchain registry contributing guide

See [Superchain Upgrades] OP-Stack specifications.

[Superchain Upgrades]: https://github.com/ethereum-optimism/optimism/blob/develop/specs/superchain-upgrades.md

## Adding a superchain target

A new Superchain Target can be added by creating a new superchain config directory,
with a `superchain.yaml` config file. Here's an example:

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

## Adding a chain

### Set env vars

To contribute a full OP-Stack chain configuration, the following data is required: contracts deployment, rollup config, L2 genesis. We provide a tool to scrape this information from your local monorepo folder.

First, export the following variables (please replace with appropriate values):
```shell
export MONOREPO_DIR=../optimism
export SUPERCHAIN_TARGET=mainnet
export CHAIN_NAME=awesomechain
export DEPLOYMENTS_DIR=${MONOREPO_DIR}/packages/contracts-bedrock/deployments/getting-started
export ROLLUP_CONFIG=${MONOREPO_DIR}/op-node/rollup.json
export GENESIS_CONFIG=${MONOREPO_DIR}/op-node/genesis.json
export PUBLIC_RPC="http://awe.some.rpc"
export SEQUENCER_RPC="http://awe.some.seq.rpc"
export EXPLORER="https://awesomescan.org"
```

### Run script
Then, run

```shell
sh scripts/add-chain.sh
```

### Understand output
The tool will write the following data:
- The main configuration source, with genesis data, and address of onchain system configuration.
- Addresses of L1 contracts. (Note that all L2 addresses are statically known addresses defined in the OP-Stack specification, and thus not configured per chain.)
- Genesis system config data
- Compressed `genesis.json` definitions (in the extra/genesis directory) which pull in the bytecode by hash

The genesis largely consists of contracts common with other chains:
all contract bytecode is deduplicated and hosted in the `extra/bytecodes` directory.

The format is a gzipped JSON `genesis.json` file, with either:
- a `alloc` attribute, structured like a standard `genesis.json`,
  but with `codeHash` (bytes32, `keccak256` hash of contract code) attribute per account,
  instead of the `code` attribute seen in standard Ethereum genesis definitions.
- a `stateHash` attribute: to omit a large state (e.g. for networks with a re-genesis or migration history).
  Nodes can load the genesis block header, and state-sync to complete the node initialization.


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
