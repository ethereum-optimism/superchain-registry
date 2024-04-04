set -e

show_usage() {
  echo "Usage: $0 <chain_type> [-h|--help]"
  echo "  chain_type: The type of chain to add. Must be 'standard' or 'frontier'."
  echo "  -h, --help: Show this usage information."
}

if [[ $# -eq 0 || $1 == "-h" || $1 == "--help" ]]; then
  show_usage
  exit 0
fi

TYPE=$1

case $TYPE in
    "standard")
        echo "Adding $TYPE chain to superchain-registry..."
        SUPERCHAIN_LEVEL=2
        ;;
    "frontier")
        echo "Adding $TYPE chain to superchain-registry..."
        SUPERCHAIN_LEVEL=1
        ;;
    *)
        echo "Invalid chain type $TYPE"
        show_usage
        exit 1
        ;;
esac

# Get the absolute path of the current script, including following symlinks
SCRIPT_PATH=$(readlink -f "$0" || realpath "$0")
# Get the directory of the current script
SCRIPT_DIR=$(dirname "$SCRIPT_PATH")
# Get the parent directory of the script's directory
PARENT_DIR=$(dirname "$SCRIPT_DIR")

SUPERCHAIN_REPO=${PARENT_DIR}


# load and echo env vars
source ${SUPERCHAIN_REPO}/.env

echo "Chain Name                      ${CHAIN_NAME}"
echo "Chain Short Name                ${CHAIN_SHORT_NAME}"
echo "Superchain target:              ${SUPERCHAIN_TARGET}"
echo "Reading from monrepo directory: ${MONOREPO_DIR}"
echo "With deployments directory:     ${DEPLOYMENTS_DIR}"
echo "Rollup config:                  ${ROLLUP_CONFIG}"
echo "Genesis config:                 ${GENESIS_CONFIG}"
echo "Public RPC endpoint:            ${PUBLIC_RPC}"
echo "Sequencer RPC endpoint:         ${SEQUENCER_RPC}"
echo "Block Explorer:                 ${EXPLORER}"

[ -d "$SUPERCHAIN_REPO/superchain/configs/$SUPERCHAIN_TARGET" ] || { echo "Superchain target directory not found. Please follow instructions to "adding a superchain target" in CONTRIBUTING.md"; exit 1; }


# add chain config
cat > $SUPERCHAIN_REPO/superchain/configs/$SUPERCHAIN_TARGET/$CHAIN_SHORT_NAME.yaml << EOF
name: $CHAIN_NAME
short_name: $CHAIN_SHORT_NAME
chain_id: $(jq -j .l2_chain_id $ROLLUP_CONFIG)
public_rpc: $PUBLIC_RPC
sequencer_rpc: $SEQUENCER_RPC
explorer: $EXPLORER

superchain_level: $SUPERCHAIN_LEVEL

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


# add extra addresses data
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


# create genesis-system-config data
# (this is deprecated, users should load this from L1, when available via SystemConfig).
mkdir -p $SUPERCHAIN_REPO/superchain/extra/genesis-system-configs/$SUPERCHAIN_TARGET
jq -r .genesis.system_config $ROLLUP_CONFIG > $SUPERCHAIN_REPO/superchain/extra/genesis-system-configs/$SUPERCHAIN_TARGET/$CHAIN_NAME.json


# create extra genesis data
mkdir -p $SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET
cd $MONOREPO_DIR
go run ./op-chain-ops/cmd/registry-data \
  --l2-genesis=$GENESIS_CONFIG \
  --bytecodes-dir=$SUPERCHAIN_REPO/superchain/extra/bytecodes \
  --output=$SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET/$CHAIN_NAME.json.gz
