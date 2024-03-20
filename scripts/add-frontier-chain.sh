set -e

# Get the absolute path of the current script, including following symlinks
SCRIPT_PATH=$(readlink -f "$0" || realpath "$0")
# Get the directory of the current script
SCRIPT_DIR=$(dirname "$SCRIPT_PATH")
# Get the parent directory of the script's directory
PARENT_DIR=$(dirname "$SCRIPT_DIR")

SUPERCHAIN_REPO=${PARENT_DIR}

# load and echo env vars
source ${SUPERCHAIN_REPO}/.env
echo "Adding chain to superchain-registry..."
echo "Chain Name                      ${CHAIN_NAME}"
echo "Superchain target:              ${SUPERCHAIN_TARGET}"
echo "With deployments directory:     ${DEPLOYMENTS_DIR}"

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
