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
