package standard

import "github.com/ethereum-optimism/superchain-registry/superchain"

type Tag string

type VersionTags = map[Tag]superchain.ContractVersions
type BytecodeHashTags = map[Tag]superchain.L1ContractBytecodeHashes

var Versions VersionTags = make(VersionTags, 0)
var BytecodeHashes BytecodeHashTags = make(BytecodeHashTags, 0)
