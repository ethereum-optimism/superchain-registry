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
cat > $SUPERCHAIN_REPO/superchain/configs/$SUPERCHAIN_TARGET/$CHAIN_NAME.yaml << EOF
name: $CHAIN_NAME
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

canyon_time: $(jq -j .canyon_time $ROLLUP_CONFIG)
delta_time: $(jq -j .delta_time $ROLLUP_CONFIG)
ecotone_time: $(jq -j .ecotone_time $ROLLUP_CONFIG)

EOF


# infer appropriate L1_RPC_URL
case $SUPERCHAIN_TARGET in
    "mainnet")
        L1_RPC_URL="https://ethereum-mainnet-rpc.allthatnode.com"
        ;;
    "sepolia")
        L1_RPC_URL="https://ethereum-sepolia-rpc.allthatnode.com"
        ;;
    *)
        echo "Unsupported Superchain Target $SUPERCHAIN_TARGET"
        exit 1
        ;;
esac

# scrape addresses from static deployment artifacts
AddressManager=$(jq -j .address $DEPLOYMENTS_DIR/AddressManager.json)
L1CrossDomainMessengerProxy=$(jq -j .address $DEPLOYMENTS_DIR/L1CrossDomainMessengerProxy.json)
L1ERC721BridgeProxy=$(jq -j .address $DEPLOYMENTS_DIR/L1ERC721BridgeProxy.json)
L1StandardBridgeProxy=$(jq -j .address $DEPLOYMENTS_DIR/L1StandardBridgeProxy.json)
L2OutputOracleProxy=$(jq -j .address $DEPLOYMENTS_DIR/L2OutputOracleProxy.json)
OptimismMintableERC20FactoryProxy=$(jq -j .address $DEPLOYMENTS_DIR/OptimismMintableERC20FactoryProxy.json)
SystemConfigProxy=$(jq -j .address $DEPLOYMENTS_DIR/SystemConfigProxy.json)
OptimismPortalProxy=$(jq -j .address $DEPLOYMENTS_DIR/OptimismPortalProxy.json)
SystemConfigProxy=$(jq -j .address $DEPLOYMENTS_DIR/SystemConfigProxy.json)

# scrape remaining address live from the chain
SuperchainConfig=$(cast call $OptimismPortalProxy "superchainConfig()(address)" -r $L1_RPC_URL) || "" # first command could fail if bedrock
Guardian=$(cast call $SuperchainConfig "guardian()(address)" -r $L1_RPC_URL) || $(cast call $OptimismPortalProxy "GUARDIAN()(address)" -r $L1_RPC_URL) # first command could fail if bedrock
Challenger=$(cast call $L2OutputOracleProxy "challenger()(address)" -r $L1_RPC_URL) # first command could fail if FPAC: TODO figure out how to get this role if FPAC
ProxyAdmin=$(jq -j .address $DEPLOYMENTS_DIR/ProxyAdmin.json)
ProxyAdminOwner=$(cast call $ProxyAdmin "owner()(address)" -r $L1_RPC_URL)
SystemConfigOwner=$(cast call $SystemConfigProxy "owner()(address)" -r $L1_RPC_URL)

# add extra addresses data to registry
mkdir -p $SUPERCHAIN_REPO/superchain/extra/addresses/$SUPERCHAIN_TARGET
cat > $SUPERCHAIN_REPO/superchain/extra/addresses/$SUPERCHAIN_TARGET/$CHAIN_NAME.json << EOF
{
  "AddressManager": "$AddressManager",
  "L1CrossDomainMessengerProxy": "$L1CrossDomainMessengerProxy",
  "L1ERC721BridgeProxy": "$L1ERC721BridgeProxy",
  "L1StandardBridgeProxy": "$L1StandardBridgeProxy",
  "L2OutputOracleProxy": "$L2OutputOracleProxy",
  "OptimismMintableERC20FactoryProxy": "$OptimismMintableERC20FactoryProxy",
  "OptimismPortalProxy": "$OptimismPortalProxy",
  "SystemConfigProxy": "$SystemConfigProxy",
  "ProxyAdmin": "$ProxyAdmin",
  "ProxyAdminOwner": "$ProxyAdminOwner",
  "SystemConfigOwner": "$SystemConfigOwner",
  "Guardian": "$Guardian",
  "Challenger": "$Challenger"
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
