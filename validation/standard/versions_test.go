package standard

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionFor(t *testing.T) {
	cl := ContractVersions{
		L1CrossDomainMessenger: VersionedContract{Version: "1.9.9"},
		OptimismPortal:         VersionedContract{Version: ""},
	}
	want := "1.9.9"
	got, err := cl.VersionFor("L1CrossDomainMessenger")
	require.NoError(t, err)
	require.Equal(t, want, got)
	_, err = cl.VersionFor("OptimismPortal")
	require.Error(t, err)
	_, err = cl.VersionFor("Garbage")
	require.Error(t, err)
}
