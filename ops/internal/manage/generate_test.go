package manage

import (
	"path"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/deployer/opaque_map"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/stretchr/testify/require"
	"github.com/tomwright/dasel"
)

func TestOpaqueToGenesis(t *testing.T) {
	om := new(opaque_map.OpaqueMap)
	err := paths.ReadJSONFile(path.Join("testdata", "expected-genesis.json"), om)
	require.NoError(t, err)

	_, err = opaqueToGenesis(om)
	require.NoError(t, err)

	err = dasel.New(om).Put("config.zippedyDooDaTime", float64(123))
	require.NoError(t, err)

	genesis, err := opaqueToGenesis(om)
	require.Error(t, err)
	require.ErrorContains(t, err, ErrNotLossless.Error())
	require.Nil(t, genesis)
}
