package superchain

import (
	"path"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestAddressFor(t *testing.T) {
	al := AddressList{
		ProxyAdmin:     MustHexToAddress("0xD98bd7A1F2384D890D0d6153cBcfCcF6F813Ab6c"),
		AddressManager: Address{},
	}
	want := MustHexToAddress("0xD98bd7A1F2384D890D0d6153cBcfCcF6F813Ab6c")
	got, err := al.AddressFor("ProxyAdmin")
	require.NoError(t, err)
	require.Equal(t, want, got)
	_, err = al.AddressFor("AddressManager")
	require.Error(t, err)
	_, err = al.AddressFor("Garbage")
	require.Error(t, err)
}

func TestVersionFor(t *testing.T) {
	cl := ContractVersions{
		L1CrossDomainMessenger: VersionedContract{Version: "1.9.9"},
		OptimismPortal:         VersionedContract{Version: ""},
	}
	want := "1.9.9"
	got, err := cl.VersionFor("L1CrossDomainMessenger")
	require.NoError(t, err)
	require.Equal(t, want, got)
	_, err = cl.VersionFor("OptimismPortal")
	require.Error(t, err)
	_, err = cl.VersionFor("Garbage")
	require.Error(t, err)
}

func TestChainIds(t *testing.T) {
	chainIDs := map[uint64]bool{}

	storeIfUnique := func(chainId uint64) {
		if chainIDs[chainId] {
			t.Fatalf("duplicate chain ID %d", chainId)
		}
		chainIDs[chainId] = true
	}

	targets, err := superchainFS.ReadDir("configs")
	require.NoError(t, err)

	for _, target := range targets {
		if target.IsDir() {
			entries, err := superchainFS.ReadDir(path.Join("configs", target.Name()))
			require.NoError(t, err)
			for _, entry := range entries {
				if !isConfigFile(entry) {
					continue
				}
				configBytes, err := superchainFS.ReadFile(path.Join("configs", target.Name(), entry.Name()))
				require.NoError(t, err)
				var chainConfig ChainConfig

				require.NoError(t, toml.Unmarshal(configBytes, &chainConfig))

				storeIfUnique(chainConfig.ChainID)
			}
		}
	}
}

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

// TestContractBytecodes verifies that all bytecodes can be loaded successfully,
// and hash to the code-hash in the name.
func TestContractBytecodes(t *testing.T) {
	entries, err := extraFS.ReadDir(path.Join("extra", "bytecodes"))
	if err != nil {
		t.Fatalf("failed to open bytecodes dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".bin.gz") {
			t.Fatalf("bytecode file has missing suffix: %q", name)
		}
		name = strings.TrimSuffix(name, ".bin.gz")
		var expected Hash
		if err := expected.UnmarshalText([]byte(name)); err != nil {
			t.Fatalf("bytecode filename %q failed to parse as hash: %v", e.Name(), err)
		}
		value, err := LoadContractBytecode(expected)
		if err != nil {
			t.Fatalf("failed to load contract code of %q: %v", e.Name(), err)
		}
		computed := keccak256(value)
		if expected != computed {
			t.Fatalf("expected bytecode hash %s but computed %s", expected, computed)
		}
	}
}

// TestCanyonTimestampOnBlockBoundary asserts that Canyon will activate on a block's timestamp.
// This is critical because the create2Deployer only activates on a block's timestamp.
func TestCanyonTimestampOnBlockBoundary(t *testing.T) {
	testStandardTimestampOnBlockBoundary(t, func(c *ChainConfig) *uint64 { return c.CanyonTime })
}

// TestEcotoneTimestampOnBlockBoundary asserts that Ecotone will activate on a block's timestamp.
// This is critical because the L2 upgrade transactions only activates on a block's timestamp.
func TestEcotoneTimestampOnBlockBoundary(t *testing.T) {
	testStandardTimestampOnBlockBoundary(t, func(c *ChainConfig) *uint64 { return c.EcotoneTime })
}

// TestAevoForkTimestamps ensures that network upgades that occur on a block boundary
// also occur on Aevo which has a non-standard block time.
func TestAevoForkTimestamps(t *testing.T) {
	aevoGenesisL2Time := uint64(1679193011)
	aevoBlockTime := uint64(10)
	config := Superchains["mainnet"]
	t.Run("canyon", testNetworkUpgradeTimestampOffset(aevoGenesisL2Time, aevoBlockTime, config.Config.hardForkDefaults.CanyonTime))
	t.Run("ecotone", testNetworkUpgradeTimestampOffset(aevoGenesisL2Time, aevoBlockTime, config.Config.hardForkDefaults.EcotoneTime))
}

func testStandardTimestampOnBlockBoundary(t *testing.T, ts func(*ChainConfig) *uint64) {
	for _, superchainConfig := range Superchains {
		for _, id := range superchainConfig.ChainIDs {
			chainCfg := OPChains[id]
			t.Run(chainCfg.Name, testNetworkUpgradeTimestampOffset(chainCfg.Genesis.L2Time, 2, ts(chainCfg)))
		}
	}
}

func testNetworkUpgradeTimestampOffset(l2GenesisTime uint64, blockTime uint64, upgradeTime *uint64) func(t *testing.T) {
	return func(t *testing.T) {
		if upgradeTime == nil {
			t.Skip("No network upgrade time")
		}
		if *upgradeTime == 0 {
			t.Skip("Upgrade occurred at genesis")
		}
		offset := *upgradeTime - l2GenesisTime
		if offset%blockTime != 0 {
			t.Fatalf("HF time is not on the block time. network upgade time: %v. L2 start time: %v, block time: %v ", *upgradeTime, l2GenesisTime, blockTime)
		}
	}
}

func TestSuperchainConfigUnmarshaling(t *testing.T) {
	rawTOML := `
name = "Mickey Mouse"
protocol_versions_addr = "0x252CbE9517F731C618961D890D534183822dcC8d"
superchain_config_addr = "0x02d91Cf852423640d93920BE0CAdceC0E7A00FA7"
op_contracts_manager_proxy_addr = "0xF564eEA7960EA244bfEbCBbB17858748606147bf"

canyon_time = 1
delta_time = 2
ecotone_time = 3
fjord_time = 4
granite_time = 5
holocene_time = 6
isthmus_time = 7

[l1]
  chain_id = 314
  public_rpc = "https://disney.com"
  explorer = "https://disneyscan.io"
`

	s := SuperchainConfig{}
	err := unMarshalSuperchainConfig([]byte(rawTOML), &s)
	require.NoError(t, err)

	expectL1Info := SuperchainL1Info{
		ChainID:   314,
		PublicRPC: "https://disney.com",
		Explorer:  "https://disneyscan.io",
	}

	require.Equal(t, "Mickey Mouse", s.Name)
	require.Equal(t, expectL1Info, s.L1)

	require.Equal(t, "0x252CbE9517F731C618961D890D534183822dcC8d", s.ProtocolVersionsAddr.String())
	require.Equal(t, "0x02d91Cf852423640d93920BE0CAdceC0E7A00FA7", s.SuperchainConfigAddr.String())
	require.Equal(t, uint64Ptr(uint64(1)), s.hardForkDefaults.CanyonTime)
	require.Equal(t, uint64Ptr(uint64(2)), s.hardForkDefaults.DeltaTime)
	require.Equal(t, uint64Ptr(uint64(3)), s.hardForkDefaults.EcotoneTime)
	require.Equal(t, uint64Ptr(uint64(4)), s.hardForkDefaults.FjordTime)
	require.Equal(t, uint64Ptr(uint64(5)), s.hardForkDefaults.GraniteTime)
	require.Equal(t, uint64Ptr(uint64(6)), s.hardForkDefaults.HoloceneTime)
	require.Equal(t, uint64Ptr(uint64(7)), s.hardForkDefaults.IsthmusTime)
}

func TestHardForkOverridesAndDefaults(t *testing.T) {
	defaultCanyonTime := uint64(3)
	defaultSuperchainConfig := SuperchainConfig{
		hardForkDefaults: HardForkConfiguration{
			CanyonTime: &defaultCanyonTime,
		},
	}
	nilDefaultSuperchainConfig := SuperchainConfig{
		hardForkDefaults: HardForkConfiguration{
			CanyonTime: nil,
		},
	}

	overridenCanyonTime := uint64Ptr(uint64(8))
	override := []byte(`canyon_time = 8`)
	nilOverride1 := []byte(`superchain_time = 2`)
	nilOverride2 := []byte(`superchain_time = 0`)
	nilOverride3 := []byte(`superchain_time = 10`)
	nilOverride4 := []byte(`
superchain_time = 1
[genesis]
  l2_time = 4
`)

	type testCase struct {
		name               string
		scConfig           SuperchainConfig
		rawTOML            []byte
		expectedCanyonTime *uint64
	}

	testCases := []testCase{
		{"default + override  (nil superchain_time)= override", defaultSuperchainConfig, override, overridenCanyonTime},
		{"nil default + override = override", nilDefaultSuperchainConfig, override, overridenCanyonTime},
		{"default + nil override (default after superchain_time) = default", defaultSuperchainConfig, nilOverride1, &defaultCanyonTime},
		{"default + nil override (default after zero superchain_time) = default", defaultSuperchainConfig, nilOverride2, &defaultCanyonTime},
		{"default + nil override (default before superchain_time) = nil", defaultSuperchainConfig, nilOverride3, nil},
		{"default + nil override (default after zero superchain_time but before genesis) = 0", defaultSuperchainConfig, nilOverride4, uint64Ptr(0)},
	}

	executeTestCase := func(t *testing.T, tt testCase) {
		c := ChainConfig{}

		err := toml.Unmarshal([]byte(tt.rawTOML), &c)
		require.NoError(t, err)

		c.setNilHardforkTimestampsToDefaultOrZero(&tt.scConfig)

		require.Equal(t, tt.expectedCanyonTime, c.CanyonTime)
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) { executeTestCase(t, tt) })
	}
}

func TestHardForkOverridesAndDefaults2(t *testing.T) {
	defaultSuperchainConfig := SuperchainConfig{
		hardForkDefaults: HardForkConfiguration{
			CanyonTime: uint64Ptr(0),
			DeltaTime:  uint64Ptr(1),
		},
	}

	c := ChainConfig{}

	rawTOML := `
ecotone_time = 2
fjord_time = 3
`

	err := toml.Unmarshal([]byte(rawTOML), &c)
	require.NoError(t, err)

	c.setNilHardforkTimestampsToDefaultOrZero(&defaultSuperchainConfig)

	var nil64 *uint64

	require.Equal(t, nil64, c.CanyonTime)
	require.Equal(t, nil64, c.DeltaTime)
	require.Equal(t, uint64Ptr(uint64(2)), c.EcotoneTime)
	require.Equal(t, uint64Ptr(uint64(3)), c.FjordTime)
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}
