#!/bin/bash
set -e

current_dir=$(pwd)
monorepo_dir="${current_dir}/../../../optimism-temporary/"
contract_dir="${monorepo_dir}/packages/contracts-bedrock/"


go_version=$(grep -m 1 '^go ' ${monorepo_dir}/go.mod | awk '{print $2}')

# Source the gvm script to load gvm functions into the shell
set +e
gvm install go${go_version} || exit 1
gvm use go${go_version} || exit 1
cd ${monorepo_dir} || exit 1
set -e

echo "Running op-node genesis l2 command"

go run op-node/cmd/main.go genesis l2 \
	--deploy-config=./packages/contracts-bedrock/deploy-config/mainnet.json \
	--outfile.l2=expected-genesis.json \
	--outfile.rollup=rollup.json \
	--deployment-dir=./packages/contracts-bedrock/deployments/mainnet \
	--l1-rpc=https://ethereum-rpc.publicnode.com
