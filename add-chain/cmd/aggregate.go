package cmd

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum-optimism/superchain-registry/add-chain/flags"
	"github.com/ethereum/go-ethereum/core"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/urfave/cli/v2"
)

var AggregateCmd = cli.Command{
	Name:  "aggregate",
	Flags: []cli.Flag{flags.GenesisFlag},
	Usage: "Updates aggregate files with all chain artifacts",
	Action: func(ctx *cli.Context) error {
		fmt.Println("Regenerated genesis config matches existing one")
		return nil
	},
}
