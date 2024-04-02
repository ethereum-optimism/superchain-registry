package validation

import (
	"flag"
	"os"
	"strconv"
	"strings"
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
	var fcid uint64
	var err error
	if rawFocussedChainId != "" {
		if strings.HasPrefix(rawFocussedChainId, "0x") {
			fcid, err = strconv.ParseUint(strings.TrimPrefix(rawFocussedChainId, "0x"), 16, 64)
		} else {
			fcid, err = strconv.ParseUint(rawFocussedChainId, 10, 64)
		}
		if err != nil {
			panic(err)
		}
		focussedChainId = &fcid
	}
	os.Exit(m.Run())
}
