package validation

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetBytes_Success(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	expected := []byte{0xde, 0xad, 0xbe, 0xef}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return(expected, nil).
		Once()

	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	result, err := getBytes("myMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.Equal(t, expected, result)

	mockClient.AssertExpectations(t)
}

func TestGetBytes_Error(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return(nil, errors.New("some call error")).
		Once()

	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	_, err := getBytes("failingMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "some call error")

	mockClient.AssertExpectations(t)
}

func TestGetHexString_Success(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	expected := []byte{0x00, 0x11, 0x22, 0x33}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return(expected, nil).
		Once()

	contractAddress := superchain.MustHexToAddress("0xabcdef7890abcdef1234567890abcdef12345678")
	hexVal, err := getHexString("hexMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.Equal(t, "00112233", hexVal)

	mockClient.AssertExpectations(t)
}

func TestGetHexString_Error(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return(nil, errors.New("getHexString error")).
		Once()

	contractAddress := superchain.MustHexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	_, err := getHexString("hexMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getHexString error")

	mockClient.AssertExpectations(t)
}

func TestGetBool_True(t *testing.T) {
	t.Parallel()

	// Returning bytes that correspond to `true` value
	mockClient := &testutils.MockEthClient{}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return([]byte("0x0100000000000000000000000000000000000000000000000000000000000000"), nil).
		Once()

	contractAddress := superchain.MustHexToAddress("0x1111117890abcdef1234567890abcdef12345678")
	val, err := getBool("boolMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.True(t, val)

	mockClient.AssertExpectations(t)
}

func TestGetBool_False(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return([]byte(""), nil).
		Once()

	contractAddress := superchain.MustHexToAddress("0x2222227890abcdef1234567890abcdef12345678")
	val, err := getBool("boolMethod()", contractAddress, mockClient)
	require.NoError(t, err)
	require.False(t, val)

	mockClient.AssertExpectations(t)
}

func TestGetBool_ErrorUnexpectedValue(t *testing.T) {
	t.Parallel()

	mockClient := &testutils.MockEthClient{}
	mockClient.On("CallContract", mock.Anything, mock.Anything, (*big.Int)(nil)).
		Return([]byte("0xabcdef"), nil).
		Once()

	contractAddress := superchain.MustHexToAddress("0x2222227890abcdef1234567890abcdef12345678")
	_, err := getBool("boolMethod()", contractAddress, mockClient)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected non-bool return value")

	mockClient.AssertExpectations(t)
}
