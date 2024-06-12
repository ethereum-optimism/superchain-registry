package superchain

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChecksums(t *testing.T) {
	for _, text := range globalListOfAddressText {
		require.NoError(t, mustBeCheckSummedHexAddr(text))
	}
}

func mustBeCheckSummedHexAddr(s string) error {
	var b Address
	err := decodeHex(b[:], []byte(s))
	if err != nil {
		return err
	}
	if expected := checksumAddress(b); expected != string(s) {
		return fmt.Errorf("Address not checksummed, got %s, expected %s", string(s), expected)
	}
	return nil
}

func TestMustBeCheckSummedHexAddr(t *testing.T) {
	testCases := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			t.Run("lower", func(t *testing.T) {
				if err := mustBeCheckSummedHexAddr(strings.ToLower(tc)); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("upper", func(t *testing.T) {
				if err := mustBeCheckSummedHexAddr(strings.ToUpper(tc)); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("tweak", func(t *testing.T) {
				tweaked := strings.Replace(tc, "a", "A", 1)
				if err := mustBeCheckSummedHexAddr(tweaked); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("roundtrip", func(t *testing.T) {
				var a Address
				require.NoError(t, a.UnmarshalText([]byte(tc)))
				dat, err := a.MarshalText()
				require.NoError(t, err)
				if string(dat) != tc {
					t.Fatalf("encoded %q but expected %q", dat, tc)
				}
			})
		})
	}
}
