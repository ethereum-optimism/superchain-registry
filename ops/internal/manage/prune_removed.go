package manage

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-fetcher/chaininfo"
	"github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

var requiredSuperchains = []config.Superchain{
	config.MainnetSuperchain,
	config.SepoliaSuperchain,
}

func ValidateRequiredSuperchains(wd string) error {
	for _, superchain := range requiredSuperchains {
		configPath := paths.SuperchainConfig(wd, superchain)
		if _, err := os.Stat(configPath); err != nil {
			return fmt.Errorf("required superchain %s is missing its superchain config: %w", superchain, err)
		}
	}
	return nil
}

func PruneRemovedChains(lgr log.Logger, wd string) error {
	if err := ValidateRequiredSuperchains(wd); err != nil {
		return err
	}

	syncer, err := NewCodegenSyncer(lgr, wd, make(map[uint64]chaininfo.Config))
	if err != nil {
		return fmt.Errorf("error creating codegen syncer: %w", err)
	}
	if err := syncer.SyncAll(); err != nil {
		return fmt.Errorf("error syncing codegen: %w", err)
	}
	return nil
}
