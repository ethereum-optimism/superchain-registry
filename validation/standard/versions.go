package standard

import "github.com/ethereum-optimism/superchain-registry/superchain"

type Tag string

type VersionTags = map[Tag]superchain.ContractVersions

var Versions VersionTags = make(VersionTags, 0)
