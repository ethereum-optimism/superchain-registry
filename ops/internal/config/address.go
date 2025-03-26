package config

import (
	"bytes"
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

	if len(dataStr) != 42 {
		return fmt.Errorf("invalid address: %s", dataStr)
	}

	if !common.IsHexAddress(dataStr) {
		return fmt.Errorf("invalid address: %s", dataStr)
	}

	addr := common.HexToAddress(dataStr)
	if addr.Hex() != dataStr {
		return fmt.Errorf("invalid checksummed address: %s", dataStr)
	}
	*a = ChecksummedAddress(addr)
	return nil
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

	if len(dataStr) != 42 {
		return fmt.Errorf("invalid address length: %s", dataStr)
	}

	if !common.IsHexAddress(dataStr) {
		return fmt.Errorf("invalid hex address: %s", dataStr)
	}

	// Convert to checksummed address
	addr := common.HexToAddress(dataStr)
	*a = ChecksummedAddress(addr)
	return nil
}

func (a *ChecksummedAddress) MarshalJSON() ([]byte, error) {
	if bytes.Equal(common.Address(*a).Bytes(), common.Address{}.Bytes()) {
		// Return null for zero addresses so it doesn't pollute the json output
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, common.Address(*a).Hex())), nil
}
