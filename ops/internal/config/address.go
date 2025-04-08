package config

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type ChecksummedAddress common.Address

func NewChecksummedAddress(addr common.Address) *ChecksummedAddress {
	a := ChecksummedAddress(addr)
	return &a
}

func (a *ChecksummedAddress) UnmarshalTOML(data any) error {
	dataStr, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected a string, got %T", data)
	}

	return a.parseAddress(dataStr)
}

func (a ChecksummedAddress) MarshalTOML() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, common.Address(a).Hex())), nil
}

func (a ChecksummedAddress) String() string {
	return common.Address(a).Hex()
}

func (a *ChecksummedAddress) UnmarshalJSON(data []byte) error {
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return fmt.Errorf("failed to unmarshal ChecksummedAddress: %w", err)
	}

	return a.parseAddress(dataStr)
}

func (a *ChecksummedAddress) MarshalJSON() ([]byte, error) {
	if common.Address(*a) == (common.Address{}) {
		// Return null for zero addresses so it doesn't pollute the json output
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, common.Address(*a).Hex())), nil
}

// Helper function for validating and parsing Ethereum addresses
func (a *ChecksummedAddress) parseAddress(addrStr string) error {
	// Validate length
	if len(addrStr) != 42 {
		return fmt.Errorf("invalid address length: %s", addrStr)
	}

	// Validate hex format
	if !common.IsHexAddress(addrStr) {
		return fmt.Errorf("invalid hex address: %s", addrStr)
	}

	// Convert to checksummed address
	addr := common.HexToAddress(addrStr)

	// Validate that the address is properly checksummed
	if addr.Hex() != addrStr {
		return fmt.Errorf("invalid checksummed address: %s", addrStr)
	}

	*a = ChecksummedAddress(addr)
	return nil
}
