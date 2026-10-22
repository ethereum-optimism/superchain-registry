package report

import (
	"testing"

	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestLatestContractsReleaseForOpcm(t *testing.T) {
	opcm := common.HexToAddress("0x44e197058fb98fb3618453b3d90cdef6f5db8297")

	actual, err := latestContractsReleaseForOpcm(opcm, validation.StandardVersionsSepolia)
	require.NoError(t, err)
	require.Equal(t, "op-contracts/v7.0.0-rc.4", actual)
}
