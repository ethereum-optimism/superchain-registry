#!/bin/bash
set -e

go_version=$(grep -m 1 '^go ' go.mod | awk '{print $2}')

# Source the gvm script to load gvm functions into the shell
set +e
source ~/.gvm/scripts/gvm || exit 1
gvm install go${go_version} || exit 1
gvm use go${go_version} || exit 1
set -e

echo "Running op-node genesis l2 command"

eval "$1"
