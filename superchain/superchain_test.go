package superchain

import "testing"

func TestConfigs(t *testing.T) {
	n := 0
	for name, sch := range Superchains {
		if name != sch.Config.Name {
			t.Errorf("superchain %q has bad key", name)
		}
		n += len(sch.ChainIDs)
	}
	for id, ch := range OPChains {
		if id != ch.ChainID {
			t.Errorf("chain %d has bad id", id)
		}
	}
	if len(OPChains) != n {
		t.Errorf("number of chains %d does not match chains in superchains %d", len(OPChains), n)
	}
	if len(OPChains) < 5 {
		t.Errorf("only got %d op chains, has everything loaded?", len(OPChains))
	}
	if len(Superchains) < 3 {
		t.Errorf("only got %d superchains, has everything loaded?", len(Superchains))
	}
	// All chains require extra addresses data until the L1 SystemConfig can support address mappings.
	if len(OPChains) != len(Addresses) {
		t.Errorf("got %d chains and %d address lists", len(OPChains), len(Addresses))
	}
	// All chains require extra genesis system config data until the
	// initial SystemConfig values can be read from the latest L1 chain state.
	if len(OPChains) != len(GenesisSystemConfigs) {
		t.Errorf("got %d chains and %d genesis system configs", len(OPChains), len(GenesisSystemConfigs))
	}
}
