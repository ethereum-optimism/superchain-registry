#!/bin/bash

set -e

# Shell script input args
monorepo_commit="d80c145e0acf23a49c6a6588524f57e32e33b91c"
go_version="1.19"

current_dir=$(pwd)
monorepo_dir="${current_dir}/../../../optimism/"
contract_dir="${monorepo_dir}/packages/contracts-bedrock/"

git checkout $monorepoCommit

#rm -rf ${monorepo_dir}node_modules
#rm -rf ${contract_dir}node_modules

cp ./foundry-config.patch ${contract_dir}foundry-config.patch

cd ${contract_dir}
echo $(pwd)
pnpm install

git apply foundry-config.patch
forge build

git apply -R foundry-config.patch

# Source the gvm script to load gvm functions into the shell
set +e
source ~/.gvm/scripts/gvm || exit 1
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
