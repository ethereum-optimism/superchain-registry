package validation

import (
	"flag"
	"os"
	"testing"
)

var (
	focussedChainId uint64
	focus           bool
)

func init() {
	flag.Uint64Var(&focussedChainId, "chain-id", 1, "ChainId to focus tests on")
	flag.BoolVar(&focus, "focus", false, "Whether to focus tests or not")
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
