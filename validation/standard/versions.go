package standard

import (
	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type Tag string

type (
	BytecodeHashTags       = map[Tag]L1ContractBytecodeHashes
	BytecodeImmutablesTags = map[Tag]ContractBytecodeImmutables
)

type VersionTags struct {
	Releases        map[Tag]superchain.ContractVersions `toml:"releases"`
	StandardRelease Tag                                 `toml:"standard_release,omitempty"`
}

var (
	Versions VersionTags = VersionTags{
		Releases: make(map[Tag]superchain.ContractVersions, 0),
	}
	BytecodeHashes     BytecodeHashTags       = make(BytecodeHashTags, 0)
	BytecodeImmutables BytecodeImmutablesTags = make(BytecodeImmutablesTags, 0)
)
