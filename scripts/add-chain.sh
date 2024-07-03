set -e

# Get the absolute path of the current script, including following symlinks
SCRIPT_PATH=$(readlink -f "$0" || realpath "$0")
# Get the directory of the current script
SCRIPT_DIR=$(dirname "$SCRIPT_PATH")
# Get the parent directory of the script's directory
SUPERCHAIN_REPO=$(dirname "$SCRIPT_DIR")
source ${SUPERCHAIN_REPO}/.env

go run ./add-chain --chain-type=$1 $2
go run ./add-chain check-rollup-config
go run ./add-chain compress-genesis
go run ./add-chain check-genesis
