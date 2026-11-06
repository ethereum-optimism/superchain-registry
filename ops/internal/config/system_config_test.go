package config

import (
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

// Chains that launched without the post-Isthmus/Jovian genesis params must keep encoding exactly as
// before: the optional fields are omitted, not written as zeros.
func TestSystemConfig_OptionalGenesisFieldsOmittedWhenUnset(t *testing.T) {
	in := `batcherAddress = "0x6887246668a3b87F54DeB3b94Ba47a6f63F32985"
overhead = "0x0000000000000000000000000000000000000000000000000000000000000000"
scalar = "0x010000000000000000000000000000000000000000000000000c5fc500000558"
gasLimit = 60000000
`
	var sc SystemConfig
	require.NoError(t, toml.Unmarshal([]byte(in), &sc))
	require.Nil(t, sc.MinBaseFee)
	require.Nil(t, sc.OperatorFeeScalar)
	require.Nil(t, sc.OperatorFeeConstant)
	require.Nil(t, sc.DAFootprintGasScalar)

	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(sc))
	require.Equal(t, in, buf.String())
}

func TestSystemConfig_OptionalGenesisFieldsRoundTrip(t *testing.T) {
	in := `batcherAddress = "0x6887246668a3b87F54DeB3b94Ba47a6f63F32985"
overhead = "0x0000000000000000000000000000000000000000000000000000000000000000"
scalar = "0x010000000000000000000000000000000000000000000000000c5fc500000558"
gasLimit = 40000000
operatorFeeScalar = 7
operatorFeeConstant = 8
minBaseFee = 5000000000
daFootprintGasScalar = 400
`
	var sc SystemConfig
	require.NoError(t, toml.Unmarshal([]byte(in), &sc))
	require.EqualValues(t, 5000000000, *sc.MinBaseFee)
	require.EqualValues(t, 7, *sc.OperatorFeeScalar)
	require.EqualValues(t, 8, *sc.OperatorFeeConstant)
	require.EqualValues(t, 400, *sc.DAFootprintGasScalar)

	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(sc))
	require.Equal(t, in, buf.String())
}
