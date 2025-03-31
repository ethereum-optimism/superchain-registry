package manage

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
	"time"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/fs"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
)

//go:embed chains.md.tmpl
var chainsReadmeTemplateData string

var funcMap = template.FuncMap{
	"checkmark": func(in bool) string {
		if in {
			return "✅"
		}

		return "❌"
	},
	"optedInSuperchain": func(in *uint64) string {
		if in != nil && *in < uint64(time.Now().Unix()) {
			return "✅"
		}

		return "❌"
	},
}

var tmpl = template.Must(template.New("chains-readme").Funcs(funcMap).Parse(chainsReadmeTemplateData))

type ChainsReadmeData struct {
	Superchains []config.Superchain
	ChainData   [][]*config.Chain
}

func GenChainsReadme(rootP string, outP string) error {
	superchains, err := paths.Superchains(rootP)
	if err != nil {
		return fmt.Errorf("error getting superchains: %w", err)
	}

	var data ChainsReadmeData

	for _, superchain := range superchains {
		cfgs, err := CollectChainConfigs(paths.SuperchainDir(rootP, superchain))
		if err != nil {
			return fmt.Errorf("error collecting chain configs: %w", err)
		}

		chainData := make([]*config.Chain, len(cfgs))
		for i, cfg := range cfgs {
			chainData[i] = cfg.Config
		}

		data.Superchains = append(data.Superchains, superchain)
		data.ChainData = append(data.ChainData, chainData)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := fs.AtomicWrite(outP, 0o755, buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write chains readme: %w", err)
	}

	return nil
}
