package validation

import (
	"errors"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/testutils"
	"github.com/stretchr/testify/require"
)

func TestGetBytes_Success(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("myMethod()")): {0xde, 0xad, 0xbe, 0xef},
		},
	}
	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	result, err := getBytes("myMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, result)
}

func TestGetBytes_Error(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Err: errors.New("some call error"),
	}
	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	_, err := getBytes("failingMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "some call error")
}

func TestGetHexString_Success(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("hexMethod()")): {0x00, 0x11, 0x22, 0x33},
		},
	}
	contractAddress := superchain.MustHexToAddress("0xabcdef7890abcdef1234567890abcdef12345678")
	hexVal, err := getHexString("hexMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.Equal(t, "00112233", hexVal)
}

func TestGetHexString_Error(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Err: errors.New("getHexString error"),
	}
	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	_, err := getHexString("hexMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getHexString error")
}

func TestGetBool_True(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("boolMethod()")): []byte("0x0100000000000000000000000000000000000000000000000000000000000000"),
		},
	}
	contractAddress := superchain.MustHexToAddress("0x1111117890abcdef1234567890abcdef12345678")
	val, err := getBool("boolMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.True(t, val)
}

func TestGetBool_False(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("boolMethod()")): []byte(""),
		},
	}
	contractAddress := superchain.MustHexToAddress("0x2222227890abcdef1234567890abcdef12345678")
	val, err := getBool("boolMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.False(t, val)
}

func TestGetBool_ErrorUnexpectedValue(t *testing.T) {
	t.Parallel()
	mockClient := &testutils.MockEthClient{
		Responses: map[string][]byte{
			string(testutils.MethodID("boolMethod()")): []byte("0xabcdef"),
		},
	}
	contractAddress := superchain.MustHexToAddress("0x2222227890abcdef1234567890abcdef12345678")
	_, err := getBool("boolMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected non-bool return value")
}
