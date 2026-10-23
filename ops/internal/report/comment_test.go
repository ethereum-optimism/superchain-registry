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

	t.Run("proof sections follow the standard game types", func(t *testing.T) {
		l1ReportJSON, err := os.ReadFile("testdata/l1-report.json")
		require.NoError(t, err)

		var l1Report L1Report
		require.NoError(t, json.Unmarshal(l1ReportJSON, &l1Report))

		stdPrestate := validation.Prestate{
			Hash: validation.Hash(common.HexToHash("0x038512e02c4c3f7bdaec27d00edf55b7155e0905301e1a88083e4e0a6764d54c")),
		}
		render := func(t *testing.T, report L1Report, stdConfig validation.ConfigParams) string {
			comment, err := RenderComment(
				&Report{L1: &report, GeneratedAt: time.Unix(1234, 0)},
				stdConfig,
				validation.StandardConfigRolesSepolia,
				stdPrestate,
				validation.StandardVersionsSepolia[validation.Semver160],
				"1234567890abcdef",
				"testChainShortName",
			)
			require.NoError(t, err)
			return comment
		}

		t.Run("super permissioned standard has no bisection rows", func(t *testing.T) {
			std := validation.StandardConfigParamsSepolia
			require.Zero(t, std.Proofs.Permissioned.MaxGameDepth, "SUPER_PERMISSIONED must not define bisection params")

			comment := render(t, l1Report, std)
			require.Contains(t, comment, "| ⚠️ | GameType | `5` | `1` |")
			require.NotContains(t, comment, "| ⚠️ | MaxGameDepth | `0` |")
			require.NotContains(t, comment, "Permissionless Proofs")
		})

		t.Run("legacy permissioned standard renders bisection rows", func(t *testing.T) {
			std := validation.StandardConfigParamsSepolia
			std.Proofs.Permissioned = validation.FDGParams{
				GameType:         1,
				MaxGameDepth:     73,
				SplitDepth:       30,
				MaxClockDuration: 302400,
				ClockExtension:   10800,
			}

			comment := render(t, l1Report, std)
			require.Contains(t, comment, "| ✅ | GameType | `1` | `1` |")
			require.Contains(t, comment, "| ✅ | MaxGameDepth | `73` | `73` |")
			require.Contains(t, comment, "| ✅ | SplitDepth | `30` | `30` |")
			require.Contains(t, comment, "| ✅ | MaxClockDuration | `302400` | `302400` |")
			require.Contains(t, comment, "| ✅ | ClockExtension | `10800` | `10800` |")
		})

		t.Run("permissionless report renders against permissionless standard", func(t *testing.T) {
			std := validation.StandardConfigParamsSepolia
			require.EqualValues(t, 9, std.Proofs.Permissionless.GameType)

			report := l1Report
			report.Proofs.Permissionless = &L1FDGReport{
				GameType:         9,
				AbsolutePrestate: common.Hash(stdPrestate.Hash),
				MaxGameDepth:     73,
				SplitDepth:       30,
				MaxClockDuration: 302400,
				ClockExtension:   10801,
			}

			comment := render(t, report, std)
			require.Contains(t, comment, "<summary>Permissionless Proofs</summary>")
			require.Contains(t, comment, "| ✅ | GameType | `9` | `9` |")
			require.Contains(t, comment, "| ✅ | MaxGameDepth | `73` | `73` |")
			require.Contains(t, comment, "| ✅ | SplitDepth | `30` | `30` |")
			require.Contains(t, comment, "| ✅ | MaxClockDuration | `302400` | `302400` |")
			require.Contains(t, comment, "| ⚠️ | ClockExtension | `10800` | `10801` |")
		})
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
