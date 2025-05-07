#!/bin/bash

# RPC URL for Mainnet (replace as needed)
RPC_URL="https://ethereum-rpc.publicnode.com"

# Find all mainnet SystemConfigProxy addresses and store them in a variable
addresses=$(find $(pwd)/superchain/configs/mainnet/ -name "*.toml" -exec sh -c 'yq eval ".addresses.SystemConfigProxy" {} 2>/dev/null || true' \; | grep -v '^$')

# Check if any addresses were found
if [ -z "$addresses" ]; then
  echo "No SystemConfigProxy addresses found."
  exit 1
fi

# Loop over each address
echo "\n Fetching blobbaseFeeScalar for each SystemConfigProxy address..."
echo "----------------------------------------"
for address in $addresses; do
  # Print the address
  echo "Address: $address"

  # Call the contract method using cast
  result=$(cast call "$address" "blobbasefeeScalar()(uint256)" --rpc-url "$RPC_URL")

  # Check if the call was successful
  if [ $? -ne 0 ]; then
    echo "Error: Failed to fetch blobbaseFeeScalar for $address\n"
    continue
  fi

  # Print the result
  echo "blobbaseFeeScalar: $result\n"
done

echo "Done."
