#!/usr/bin/env bash
set -o errexit -o pipefail
set -x

# Get the list of changed files
targetList=$(git diff --name-only --merge-base main -- "validation/genesis/*.toml" "validation/genesis/*.json")

# Check if targetList is empty
if [ -z "$targetList" ]; then
  echo "No matching .toml files found in 'validation/genesis/'. Exiting."
  exit 0
fi

# Process the targetList to extract directory names and then the base names
targetList=$(echo "$targetList" | xargs dirname | xargs basename | sort -u)

# Join the array elements with commas and wrap each element in quotes
targets=$(echo "$targetList" | sed 's/.*/"&"/' | tr '\n' ',')

# Remove the trailing comma
targets=${targets%,}

# Wrap in square brackets
targets="[$targets]"

echo "Will run genesis allocs validation on chains with ids $targets"

# Now build another array, each element prepended with "golang-validate-genesis-allocs-"
prependedTargets=$(echo "$targetList" | sed 's/.*/"golang-validate-genesis-allocs-&"/' | tr '\n' ',')

# Remove the trailing comma
prependedTargets=${prependedTargets%,}

# Wrap in square brackets
prependedTargets="[$prependedTargets]"

# Install yq
brew install yq

# Use yq to replace the target-version   key
yq e ".workflows.pr-checks.jobs[0].golang-validate-genesis-allocs.matrix.parameters.chainid = $targets" -i .circleci/continue_config.yml
yq e ".workflows.pr-checks.jobs[1].genesis-allocs-all-ok.requires = $prependedTargets" -i .circleci/continue_config.yml
