package superchain

import (
	"strings"
	"testing"
)

func TestAddressChecksum(t *testing.T) {
	testCases := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
	}
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			t.Run("lower", func(t *testing.T) {
				var a Address
				if err := a.UnmarshalText([]byte(strings.ToLower(tc))); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("upper", func(t *testing.T) {
				var a Address
				if err := a.UnmarshalText([]byte(strings.ToUpper(tc))); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("tweak", func(t *testing.T) {
				tweaked := strings.Replace(tc, "a", "A", 1)
				var a Address
				if err := a.UnmarshalText([]byte(tweaked)); err == nil {
					t.Fatal("expected error")
				}
			})
			t.Run("roundtrip", func(t *testing.T) {
				var a Address
				if err := a.UnmarshalText([]byte(tc)); err != nil {
					t.Fatal(err)
				}
				dat, err := a.MarshalText()
				if err != nil {
					t.Fatal(err)
				}
				if string(dat) != tc {
					t.Fatalf("encoded %q but expected %q", dat, tc)
				}
			})
		})
	}
}
