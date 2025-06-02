package report

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestRenderComment(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		l1ReportJSON, err := os.ReadFile("testdata/l1-report.json")
		require.NoError(t, err)

		expComment, err := os.ReadFile("testdata/expected-comment.md")
		require.NoError(t, err)

		var l1Report L1Report
		require.NoError(t, json.Unmarshal(l1ReportJSON, &l1Report))

		l2Report := &L2Report{
			Release: string(validation.Semver170),
			GenesisDiffs: []string{
				"genesis.alloc.0x0000000000000000000000000000000000000123: exists in second map but not in first (value: map[balance:0x64 code:0x010203 nonce:0x1 storage:map[0x456:0x789]])",
				"genesis.config.chainId: 1 => 11155111",
				"genesis.timestamp: 0x0 => 0x64",
				"genesis.difficulty: 0x1 => 0x0",
			},
		}

		comment, err := RenderComment(
			&Report{
				L1:          &l1Report,
				L2:          l2Report,
				GeneratedAt: time.Unix(1234, 0),
			},
			validation.StandardConfigParamsSepolia,
			validation.StandardConfigRolesSepolia,
			validation.Prestate{
				Hash: validation.Hash(common.HexToHash("0x038512e02c4c3f7bdaec27d00edf55b7155e0905301e1a88083e4e0a6764d54c")),
			},
			validation.StandardVersionsSepolia[validation.Semver160],
			"1234567890abcdef",
			"testChainShortName",
		)
		require.NoError(t, err)

		require.Equal(t, string(expComment), comment)
	})

	t.Run("error states", func(t *testing.T) {
		comment, err := RenderComment(
			&Report{
				L1Err:       errors.New("l1 report failed"),
				L2Err:       errors.New("l2 report failed"),
				GeneratedAt: time.Unix(1234, 0),
			},
			validation.StandardConfigParamsSepolia,
			validation.StandardConfigRolesSepolia,
			validation.StandardPrestates.StablePrestate(),
			validation.StandardVersionsSepolia[validation.Semver160],
			"1234567890abcdef",
			"testChainShortName",
		)
		require.NoError(t, err)

		expComment, err := os.ReadFile("testdata/expected-error-comment.md")
		require.NoError(t, err)

		require.Equal(t, string(expComment), comment)
	})
}
