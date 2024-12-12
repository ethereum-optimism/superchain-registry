package standard_test

import (
	"context"
	"testing"

	. "github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

// This test will check:
// For each each superchain
// For each contract release
// For each contract version declaration which specifies an address
// There is a contract at that address with the matching bytecode and semver.
// This is a consistency check on the standard package itself, not any particular chain
func TestStandardVersionConsistency(t *testing.T) {
	for _, superchain := range []string{"sepolia", "mainnet"} {
		rpcEndpoint := Superchains[superchain].Config.L1.PublicRPC
		require.NotEmpty(t, rpcEndpoint)

		client, err := ethclient.Dial(rpcEndpoint)
		require.NoErrorf(t, err, "could not dial rpc endpoint %s", rpcEndpoint)

		for tag, release := range standard.ContractVersions[superchain] {
			for _, contract := range release.GetNonEmpty() {
				vc, err := release.VersionedContractFor(contract)
				require.NoError(t, err)
				if vc.Address != nil && vc.ImplementationAddress != nil {
					require.FailNow(t, "should not specify both address and implementation address")
				}
				addressToCheck := vc.Address
				if vc.ImplementationAddress != nil {
					addressToCheck = vc.ImplementationAddress
				}
				if addressToCheck != nil {
					version, err := validation.GetContractVersion(context.Background(), common.Address(*addressToCheck), client)
					require.NoErrorf(t, err, "could not get version for address %s, contract %s, superchain %s, release %s", addressToCheck, contract, superchain, tag)
					require.Equal(t, vc.Version, version, "version mismatch for %s", addressToCheck)

					dummyChainId := uint64(12345678) // we don't actually need a chain ID for this since we wouldn't ever pass a proxy here
					bytecodeHash, err := validation.GetBytecodeHash(context.Background(), dummyChainId, contract, common.Address(*addressToCheck), client, tag)
					require.NoError(t, err)

					h, err := standard.BytecodeHashes[tag].GetBytecodeHashFor(contract)
					require.NoErrorf(t, err, "could not get hash for contract %s, release %s", contract, tag)
					require.Equal(t, h, bytecodeHash, "bytecode hash mismatch for ", contract, "release", tag, "superchain", superchain, "should be", bytecodeHash)
				}
			}
		}
	}
}
