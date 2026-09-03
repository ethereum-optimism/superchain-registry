package gameargs

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

const (
	PermissionlessArgsLength = 124
	PermissionedArgsLength   = 164
)

var ErrInvalidGameArgs = errors.New("invalid game args")

func ParseAbsoluteState(args []byte) (common.Hash, error) {
	if len(args) != PermissionlessArgsLength && len(args) != PermissionedArgsLength {
		return common.Hash{}, fmt.Errorf("%w: invalid length (%v)", ErrInvalidGameArgs, len(args))
	}
	// In both permissioned and permissionless game args, the absolute prestate is the first 32 bytes.
	return common.BytesToHash(args[0:32]), nil
}

func ParseChallenger(args []byte) (common.Address, error) {
	if len(args) != PermissionedArgsLength {
		return common.Address{}, fmt.Errorf("%w: permissioned game args must have length %v (got %v)", ErrInvalidGameArgs, PermissionedArgsLength, len(args))
	}
	// Permissioned game args end with proposer (20 bytes) and challenger (20 bytes).
	return common.BytesToAddress(args[len(args)-common.AddressLength:]), nil
}
