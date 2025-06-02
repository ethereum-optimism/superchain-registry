package report

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type L1SemversReport struct {
	SystemConfig                       string
	PermissionedDisputeGame            string
	OptimismPortal                     string
	AnchorStateRegistry                string
	DelayedWETHPermissionedDisputeGame string
	DisputeGameFactory                 string
	L1CrossDomainMessenger             string
	L1StandardBridge                   string
	L1ERC721Bridge                     string
	OptimismMintableERC20Factory       string
}

type L1OwnershipReport struct {
	Guardian        common.Address
	Challenger      common.Address
	ProxyAdminOwner common.Address
}

type L1FDGReport struct {
	GameType         uint32
	AbsolutePrestate common.Hash
	MaxGameDepth     uint64
	SplitDepth       uint64
	MaxClockDuration uint64
	ClockExtension   uint64
}

type L1ProofsReport struct {
	Permissioned   L1FDGReport
	Permissionless *L1FDGReport
}

type L1SystemConfigReport struct {
	GasLimit               uint64
	Scalar                 *big.Int
	Overhead               *big.Int
	BaseFeeScalar          uint32
	BlobBaseFeeScalar      uint32
	EIP1559Denominator     uint32
	EIP1559Elasticity      uint32
	IsGasPayingToken       bool
	GasPayingToken         common.Address
	GasPayingTokenDecimals uint8
	GasPayingTokenName     string
	GasPayingTokenSymbol   string
}

type L1Report struct {
	Release           string
	DeploymentChainID uint64
	DeploymentTxHash  common.Hash
	Semvers           L1SemversReport
	Ownership         L1OwnershipReport
	Proofs            L1ProofsReport
	SystemConfig      L1SystemConfigReport
}

type L2Report struct {
	Release      string
	ChainID      common.Hash
	GenesisDiffs []string
}

type Report struct {
	L1          *L1Report
	L1Err       error
	L2          *L2Report
	L2Err       error
	GeneratedAt time.Time
}
