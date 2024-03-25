set -e

if [ "$CI" = "true" ]
then
  MAINNET_RPC_URL=https://ci-mainnet-l1.optimism.io
  SEPOLIA_RPC_URL=https://ci-sepolia-l1.optimism.io
  COMPUTE_UNITS_PER_SECOND=320
else
  MAINNET_RPC_URL=https://ethereum-mainnet-rpc.allthatnode.com
  SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.allthatnode.com
  COMPUTE_UNITS_PER_SECOND=320
fi


forge build
# Note: If RPC is being rate-limited, consider reducing
# --compute-units-per-second or using --fork-retries and
# --fork-retry-backoff to stay under the limit.
if [ "$SECURITY_CONFIGS" = "true" ]; then
 EXEC_SCRIPT="CheckSecurityConfigs"
 forge script $EXEC_SCRIPT --fork-url=$SEPOLIA_RPC_URL --compute-units-per-second=$COMPUTE_UNITS_PER_SECOND
fi

if [ "$MULTISIG_CONFIGS" = "true" ]; then
  EXEC_SCRIPT="CheckNetworkMultiSigConfigs"
fi

forge script $EXEC_SCRIPT --fork-url=$MAINNET_RPC_URL --compute-units-per-second=$COMPUTE_UNITS_PER_SECOND