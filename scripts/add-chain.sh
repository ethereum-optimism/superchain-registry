set -e

go run ./add-chain -chain-type=$1

# Get the absolute path of the current script, including following symlinks
SCRIPT_PATH=$(readlink -f "$0" || realpath "$0")
# Get the directory of the current script
SCRIPT_DIR=$(dirname "$SCRIPT_PATH")
# Get the parent directory of the script's directory
SUPERCHAIN_REPO=$(dirname "$SCRIPT_DIR")
source ${SUPERCHAIN_REPO}/.env

# create extra genesis data
mkdir -p $SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET
cd $MONOREPO_DIR
go run ./op-chain-ops/cmd/registry-data \
  --l2-genesis=$GENESIS_CONFIG \
  --bytecodes-dir=$SUPERCHAIN_REPO/superchain/extra/bytecodes \
  --output=$SUPERCHAIN_REPO/superchain/extra/genesis/$SUPERCHAIN_TARGET/$CHAIN_NAME.json.gz
