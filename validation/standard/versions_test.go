package standard

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestL1BytecodeHashes(t *testing.T) {
	var testBytecodeHashes L1ContractBytecodeHashes

	_, err := testBytecodeHashes.GetBytecodeHashFor("FakeContractName")
	require.ErrorIs(t, err, ErrNoSuchContractName)

	testBytecodeHashes.FaultDisputeGame = ""
	_, err = testBytecodeHashes.GetBytecodeHashFor("FaultDisputeGame")
	require.ErrorIs(t, err, ErrHashNotSpecified)

	testBytecodeHashes.DisputeGameFactory = "0x123"
	_, err = testBytecodeHashes.GetBytecodeHashFor("DisputeGameFactory")
	require.NoError(t, err)

	contractNames := testBytecodeHashes.GetNonEmpty()
	require.Equal(t, 1, len(contractNames))
}
