# Superchain registry contributing guide

See [Superchain Upgrades] OP-Stack specifications.

[Superchain Upgrades]: https://github.com/ethereum-optimism/optimism/blob/develop/specs/superchain-upgrades.md

## Adding a superchain target

A new Superchain Target can be added by creating a new superchain config directory,
with a `superchain.yaml` config file.

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

### Prerequisites

To contribute a full OP-Stack chain configuration, the following data is required:
contracts deployment, rollup config, L2 genesis.

For example:
```bash
cd optimism
export SUPERCHAIN_REPO=../superchain-registry
export SUPERCHAIN_TARGET=goerli-dev-0
export CHAIN_NAME=op-labs-devnet-0
export DEPLOYMENTS_DIR=./packages/contracts-bedrock/deployments/internal-devnet
export ROLLUP_CONFIG=./internal_devnet_rollup.json
export GENESIS_CONFIG=./internal_devnet_genesis.json
```

### `configs`

The config is the main configuration source, with genesis data, and address of onchain system configuration:

```bash
cat > $SUPERCHAIN_REPO/superchain/configs/$SUPERCHAIN_TARGET/$CHAIN_NAME.yaml << EOF
name: $CHAIN_NAME
chain_id: $(jq -j .l2_chain_id $ROLLUP_CONFIG)
public_rpc: ""
sequencer_rpc: ""
explorer: ""

batch_inbox_addr: "$(jq -j .batch_inbox_address $ROLLUP_CONFIG)"

genesis:
  l1:
    hash: "$(jq -j .genesis.l1.hash $ROLLUP_CONFIG)"
    number: $(jq -j .genesis.l1.number $ROLLUP_CONFIG)
  l2:
    hash: "$(jq -j .genesis.l2.hash $ROLLUP_CONFIG)"
    number: $(jq -j .genesis.l2.number $ROLLUP_CONFIG)
  l2_time: $(jq -j .genesis.l2_time $ROLLUP_CONFIG)
EOF
```

### Extras

Extra configuration is made available for node-operator UX, but not a hard requirement.

#### `extra/addresses`

Addresses of L1 contracts.

Note that all L2 addresses are statically known addresses defined in the OP-Stack specification,
and thus not configured per chain.

```bash
# create extra addresses data
mkdir -p $SUPERCHAIN_REPO/superchain/extra/addresses/$SUPERCHAIN_TARGET
cat > $SUPERCHAIN_REPO/superchain/extra/addresses/$SUPERCHAIN_TARGET/$CHAIN_NAME.json << EOF
{
  "AddressManager": "$(jq -j .address $DEPLOYMENTS_DIR/AddressManager.json)",
  "L1CrossDomainMessengerProxy": "$(jq -j .address $DEPLOYMENTS_DIR/L1CrossDomainMessengerProxy.json)",
  "L1ERC721BridgeProxy": "$(jq -j .address $DEPLOYMENTS_DIR/L1ERC721BridgeProxy.json)",
  "L1StandardBridgeProxy": "$(jq -j .address $DEPLOYMENTS_DIR/L1StandardBridgeProxy.json)",
  "L2OutputOracleProxy": "$(jq -j .address $DEPLOYMENTS_DIR/L2OutputOracleProxy.json)",
  "OptimismMintableERC20FactoryProxy": "$(jq -j .address $DEPLOYMENTS_DIR/OptimismMintableERC20FactoryProxy.json)",
  "OptimismPortalProxy": "$(jq -j .address $DEPLOYMENTS_DIR/OptimismPortalProxy.json)",
  "SystemConfigProxy": "$(jq -j .address $DEPLOYMENTS_DIR/SystemConfigProxy.json)",
  "ProxyAdmin": "$(jq -j .address $DEPLOYMENTS_DIR/ProxyAdmin.json)"
}
EOF
```

#### `extra/genesis-system-configs`

Genesis system config data is provided but may be optional for chains deployed with future OP-Stack protocol versions.

Newer versions of the `SystemConfig` contract persist the L1 starting block in the contract storage,
and a node can load the initial system-config values from L1
by inspecting the `SystemConfig` receipts of this given L1 block.

```bash
# create genesis-system-config data
# (this is deprecated, users should load this from L1, when available via SystemConfig).
mkdir -p $SUPERCHAIN_REPO/superchain/extra/genesis-system-configs/$SUPERCHAIN_TARGET
jq -r .genesis.system_config $ROLLUP_CONFIG > $SUPERCHAIN_REPO/superchain/extra/genesis-system-configs/$SUPERCHAIN_TARGET/$CHAIN_NAME.json
```

#### `extra/genesis`

The `extra/genesis` directory hosts compressed `genesis.json` definitions that pull in the bytecode by hash

The genesis largely consists of contracts common with other chains:
all contract bytecode is deduplicates and hosted in the `extra/bytecodes` directory.

The format is a gzipped JSON `genesis.json` file, with either:
- a `alloc` attribute, structured like a standard `genesis.json`,
  but with `codeHash` (bytes32, `keccak256` hash of contract code) attribute per account,
  instead of the `code` attribute seen in standard Ethereum genesis definitions.
- a `stateHash` attribute: to omit a large state (e.g. for networks with a re-genesis or migration history).
  Nodes can load the genesis block header, and state-sync to complete the node initialization.

```bash
# create extra genesis data
mkdir -p $SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET
go run ./op-chain-ops/cmd/registry-data \
  --l2-genesis=$GENESIS_CONFIG \
  --bytecodes-dir=$SUPERCHAIN_REPO/superchain/extra/bytecodes \
  --output=$SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET/$CHAIN_NAME.json.gz
```

## Setting up your editor for formatting and linting
If you use VSCode, you can place the following in a `settings.json` file in the gitignored `.vscode` directory:

```
{
    "go.formatTool": "gofumpt",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "workspace",
    "gopls": {
        "formatting.gofumpt": true,
    },
}
```
