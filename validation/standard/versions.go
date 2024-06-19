package standard

import "github.com/ethereum-optimism/superchain-registry/superchain"

type Tag string

type VersionsType = map[Tag]superchain.ContractVersions

var Versions VersionsType = make(VersionsType, 0)
