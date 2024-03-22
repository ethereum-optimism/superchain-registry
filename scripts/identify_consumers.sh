#!/bin/bash

# As per issue: https://github.com/ethereum-optimism/security-pod/issues/73 - We want to identify all consumers of the superchain-registry and what types they're relying on.
# Inspired by the following Github search: https://github.com/search?q=org%3Aethereum-optimism+%22github.com%2Fethereum-optimism%2Fsuperchain-registry%2Fsuperchain%22+NOT+repo%3Aethereum-optimism%2Fsuperchain-registry+&type=code
# Note: You should have all the relevant repositories cloned locally.

echo -e "\nSearching current directory for Superchain dependencies...\n"

# Search top-level directory where you store all your cloned repositories. Change if needed.
directory="../../"
file_count=0

find "$directory" -type d -name "superchain-registry" -prune -o -type f -name "*.go" -exec grep -l "github.com/ethereum-optimism/superchain-registry/superchain" {} \; | while IFS= read -r file; do
    echo "File: $file"
    file_count=$(expr $file_count + 1)
    output=$(grep -n "superchain\." "$file")
    echo -e "\n$output"
    echo "File Count $file_count"
done

echo "Complete."