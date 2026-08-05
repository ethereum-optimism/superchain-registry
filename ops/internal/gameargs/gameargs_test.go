package gameargs

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestParseChallenger(t *testing.T) {
	expected := common.HexToAddress("0xfd1d2e729ae8eee2e146c033bf4400fe75284301")
	args := make([]byte, PermissionedArgsLength)
	copy(args[len(args)-common.AddressLength:], expected.Bytes())

	actual, err := ParseChallenger(args)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestParseChallengerRejectsPermissionlessArgs(t *testing.T) {
	_, err := ParseChallenger(make([]byte, PermissionlessArgsLength))
	require.ErrorIs(t, err, ErrInvalidGameArgs)
}
