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

		for tag, release := range standard.NetworkVersions[superchain].Releases {
			for _, contract := range release.GetNonEmpty() {
				releaseVersion := release
				vc, err := releaseVersion.VersionedContractFor(contract)
				if err != nil {
					t.Fatal(err)
				}
				if vc.Address != nil && vc.ImplementationAddress != nil {
					t.Fatal("should not specify both address and implementation address")
				}
				var addressToCheck *Address
				if vc.Address != nil {
					addressToCheck = vc.Address
				}
				if vc.ImplementationAddress != nil {
					addressToCheck = vc.ImplementationAddress
				}
				if addressToCheck != nil {
					version, err := validation.GetVersion(context.Background(), common.Address(*addressToCheck), client)
					if err != nil {
						t.Fatal(err, "could not get version for %s", addressToCheck, "contract", contract, "superchain", superchain, "release", tag)
					}
					require.Equal(t, vc.Version, version, "version mismatch for %s", addressToCheck)

					if version == vc.Version {
						t.Log("version match for", addressToCheck, "contract", contract, "superchain", superchain, "release", tag)
					}

					dummyChainId := uint64(12345678) // we don't actually need a chain ID for this since we wouldn't ever pass a proxy here
					bytedcodeHash, err := validation.GetBytecodeHash(context.Background(), dummyChainId, contract, common.Address(*addressToCheck), client)
					if err != nil {
						t.Fatal(err)
					}

					h, err := standard.BytecodeHashes[tag].GetBytecodeHashFor(contract)
					if err != nil {
						t.Fatal(err, "could not get hash for ", contract, "release", tag, "should be", bytedcodeHash)
					}

					require.Equal(t, h, bytedcodeHash, "bytecode hash mismatch for ", contract, "release", tag, "should be", bytedcodeHash)

				}
			}
		}
	}
}
