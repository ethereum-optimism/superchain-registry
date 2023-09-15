package superchain

import (
	"path"
	"testing"
)

func TestConfigs(t *testing.T) {
	n := 0
	for name, sch := range Superchains {
		if name != sch.Superchain {
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

func TestGenesis(t *testing.T) {
	for id := range OPChains {
		_, err := LoadGenesis(id)
		if err != nil {
			t.Fatalf("failed to load genesis of chain %d: %v", id, err)
		}
	}
}

// TestImplementations ensures that the global Implementations
// map is populated.
func TestImplementations(t *testing.T) {
	if len(Implementations) == 0 {
		t.Fatal("no implementations found")
	}
}

// TestContractImplementations tests specific contracts implementations are set
// correctly.
func TestContractImplementations(t *testing.T) {
	impls, err := NewContractImplementations(path.Join("implementations", "implementations.yaml"))
	if err != nil {
		t.Fatalf("failed to load contract implementations: %v", err)
	}
	if impls.L1CrossDomainMessenger["1.6.0"] != HexToAddress("0xf4d5682dA3ad1820ea83E1cEE5Fd92a3A7BabC30") {
		t.Fatal("wrong L1CrossDomainMessenger address")
	}
	if impls.L1ERC721Bridge["1.3.0"] != HexToAddress("0x8ADd7FB53A242e827373519d260EE3B8F7612Ba1") {
		t.Fatal("wrong L1ERC721Bridge address")
	}
	if impls.L1StandardBridge["1.3.0"] != HexToAddress("0x9c540e769B9453d174EdB683a90D9170e6559F16") {
		t.Fatal("wrong L1StandardBridge address")
	}
	if impls.L2OutputOracle["1.5.0"] != HexToAddress("0x7a811C9862ab54E677EEdA7e6F075aC86a1f551e") {
		t.Fatal("wrong L2OutputOracle address")
	}
	if impls.OptimismMintableERC20Factory["1.4.0"] != HexToAddress("0x135B9097A0e1e56190251c62f111B676Fb4Ec494") {
		t.Fatal("wrong OptimismMintableERC20 address")
	}
	if impls.OptimismPortal["1.9.0"] != HexToAddress("0x8Cfa294bD0c6F63cD65d492bdB754eAcf684D871") {
		t.Fatal("wrong OptimismPortal address")
	}
	if impls.SystemConfig["1.7.0"] != HexToAddress("0x09323D05868393c7EBa8190BAc173f843b82030a") {
		t.Fatal("wrong SystemConfig address")
	}
}
