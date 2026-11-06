package manage

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestSystemConfigFromRollup_CarriesGenesisParams(t *testing.T) {
	in := eth.SystemConfig{
		BatcherAddr:          common.HexToAddress("0x7d451299ab21bfa7132c690b91d749756ec1dbb7"),
		Scalar:               eth.Bytes32(common.HexToHash("0x010000000000000000000000000000000000000000000000000c3c9d00000558")),
		GasLimit:             40_000_000,
		OperatorFeeParams:    eth.EncodeOperatorFeeParams(eth.OperatorFeeParams{Scalar: 7, Constant: 8}),
		MinBaseFee:           5_000_000_000,
		DAFootprintGasScalar: 400,
	}
	out := SystemConfigFromRollup(in)

	require.EqualValues(t, 40_000_000, out.GasLimit)
	require.NotNil(t, out.MinBaseFee)
	require.EqualValues(t, 5_000_000_000, *out.MinBaseFee)
	require.NotNil(t, out.OperatorFeeScalar)
	require.EqualValues(t, 7, *out.OperatorFeeScalar)
	require.NotNil(t, out.OperatorFeeConstant)
	require.EqualValues(t, 8, *out.OperatorFeeConstant)
	require.NotNil(t, out.DAFootprintGasScalar)
	require.EqualValues(t, 400, *out.DAFootprintGasScalar)
}

func TestSystemConfigFromRollup_OmitsZeroValues(t *testing.T) {
	out := SystemConfigFromRollup(eth.SystemConfig{GasLimit: 30_000_000})
	require.Nil(t, out.MinBaseFee)
	require.Nil(t, out.OperatorFeeScalar)
	require.Nil(t, out.OperatorFeeConstant)
	require.Nil(t, out.DAFootprintGasScalar)
}
