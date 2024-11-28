#!/bin/bash

# Script to trigger a CircleCI pipeline via API
# Usage: ./circleci-upload-chain-artifacts.sh <chain> [branch]

# Configuration
CONFIG_FILE="$HOME/.circleci/cli.yml"
CIRCLECI_CLI_INSTALL_LINK="https://circleci.com/docs/local-cli/#installation"
CIRCLECI_SETUP_LINK="https://app.circleci.com/settings/user/tokens"

# Check if the CircleCI CLI configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo "CircleCI CLI configuration file not found at $CONFIG_FILE."
    echo "Please install the CircleCI CLI: $CIRCLECI_CLI_INSTALL_LINK"
    echo "Then run 'circleci setup' to configure your token: $CIRCLECI_SETUP_LINK"
    exit 1
fi

# Extract the token from the configuration file
TOKEN=$(awk '/^token:/ {print $2}' "$CONFIG_FILE")

if [ -z "$TOKEN" ]; then
    echo "Token not found in $CONFIG_FILE."
    echo "Please run 'circleci setup' to configure your token: $CIRCLECI_SETUP_LINK"
    exit 1
fi

# Check if the 'chain' parameter is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <chain> [branch]"
  echo "Error: The 'chain' parameter is required."
  read -p "Enter the chain name: " CHAIN
else
  CHAIN=$1
fi

if [ -z "$CHAIN" ]; then
    echo "Error: Chain name is required."
    exit 1
fi

BRANCH=${2:-main}  # Default to 'main' branch if not provided

# Define the API endpoint
API_ENDPOINT="https://circleci.com/api/v2/project/gh/ethereum-optimism/superchain-registry/pipeline"

# Trigger the pipeline via API
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST $API_ENDPOINT \
    -H "Circle-Token: $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
          "branch": "'"$BRANCH"'",
          "parameters": {
            "chain": "'"$CHAIN"'",
            "interactive": true
          }
        }')

if [ "$RESPONSE" -eq 200 ] || [ "$RESPONSE" -eq 201 ]; then
  echo "Successfully triggered the pipeline for chain '$CHAIN' on branch '$BRANCH'."
else
  echo "Failed to trigger the pipeline. HTTP response code: $RESPONSE"
  echo "Please check your API token and project details."
fi
