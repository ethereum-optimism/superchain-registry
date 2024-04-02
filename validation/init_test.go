package validation

import (
	"flag"
	"os"
	"strconv"
	"testing"
)

var (
	focussedChainId    *uint64
	rawFocussedChainId string
)

func init() {
	flag.StringVar(&rawFocussedChainId, "focus-chain-id", "", "ChainId to focus tests on")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if rawFocussedChainId != "" {
		fcid, err := strconv.ParseUint(rawFocussedChainId, 10, 64)
		if err != nil {
			panic(err)
		}
		focussedChainId = &fcid
	}
	os.Exit(m.Run())
}
