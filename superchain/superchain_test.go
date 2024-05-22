package superchain

import (
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestAddressFor(t *testing.T) {
	al := AddressList{
		ProxyAdmin:     HexToAddress("0xD98bD7a1F2384D890d0D6153CbCFcCF6F813ab6c"),
		AddressManager: Address{},
	}
	want := HexToAddress("0xD98bD7a1F2384D890d0D6153CbCFcCF6F813ab6c")
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
		L1CrossDomainMessenger: "1.9.9",
		OptimismPortal:         "",
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

				require.NoError(t, yaml.Unmarshal(configBytes, &chainConfig))

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
	impls, err := newContractImplementations("")
	if err != nil {
		t.Fatalf("failed to load contract implementations: %v", err)
	}
	if impls.L1CrossDomainMessenger.Get("1.6.0") != HexToAddress("0xf4d5682dA3ad1820ea83E1cEE5Fd92a3A7BabC30") {
		t.Fatal("wrong L1CrossDomainMessenger address")
	}
	if impls.L1ERC721Bridge.Get("1.3.0") != HexToAddress("0x8ADd7FB53A242e827373519d260EE3B8F7612Ba1") {
		t.Fatal("wrong L1ERC721Bridge address")
	}
	if impls.L1StandardBridge.Get("1.3.0") != HexToAddress("0x9c540e769B9453d174EdB683a90D9170e6559F16") {
		t.Fatal("wrong L1StandardBridge address")
	}
	if impls.L2OutputOracle.Get("1.5.0") != HexToAddress("0x7a811C9862ab54E677EEdA7e6F075aC86a1f551e") {
		t.Fatal("wrong L2OutputOracle address")
	}
	if impls.OptimismMintableERC20Factory.Get("1.4.0") != HexToAddress("0x135B9097A0e1e56190251c62f111B676Fb4Ec494") {
		t.Fatal("wrong OptimismMintableERC20 address")
	}
	if impls.OptimismPortal.Get("1.9.0") != HexToAddress("0x8Cfa294bD0c6F63cD65d492bdB754eAcf684D871") {
		t.Fatal("wrong OptimismPortal address")
	}
	if impls.SystemConfig.Get("1.7.0") != HexToAddress("0x09323D05868393c7EBa8190BAc173f843b82030a") {
		t.Fatal("wrong SystemConfig address")
	}
}

// TestContractVersionsCheck will fail if the superchain semver file
// is not read correctly.
func TestContractVersionsCheck(t *testing.T) {
	for _, versions := range SuperchainSemver {
		if err := versions.Check(true); err != nil {
			t.Fatal(err)
		}
	}
}

// TestContractVersionsResolve will test that the high level interface used works.
func TestContractVersionsResolve(t *testing.T) {
	impls, err := newContractImplementations("sepolia")
	if err != nil {
		t.Fatalf("failed to load contract implementations: %v", err)
	}

	if impls.L1CrossDomainMessenger.Get("1.6.0") == (Address{}) {
		t.Fatal("wrong L1CrossDomainMessenger address")
	}
	if impls.L1ERC721Bridge.Get("1.3.0") == (Address{}) {
		t.Fatal("wrong L1ERC721Bridge address")
	}
	if impls.L1StandardBridge.Get("1.3.0") == (Address{}) {
		t.Fatal("wrong L1StandardBridge address")
	}
	if impls.L2OutputOracle.Get("1.5.0") == (Address{}) {
		t.Fatal("wrong L2OutputOracle address")
	}
	if impls.OptimismMintableERC20Factory.Get("1.4.0") == (Address{}) {
		t.Fatal("wrong OptimismMintableERC20 address")
	}
	if impls.OptimismPortal.Get("1.9.0") == (Address{}) {
		t.Fatal("wrong OptimismPortal address")
	}
	if impls.SystemConfig.Get("1.7.0") == (Address{}) {
		t.Fatal("wrong SystemConfig address")
	}

	versions := ContractVersions{
		L1CrossDomainMessenger:       "1.6.0",
		L1ERC721Bridge:               "1.3.0",
		L1StandardBridge:             "1.3.0",
		L2OutputOracle:               "1.5.0",
		OptimismMintableERC20Factory: "1.4.0",
		OptimismPortal:               "1.9.0",
		SystemConfig:                 "1.7.0",
	}

	list, err := impls.Resolve(versions)
	if err != nil {
		t.Fatalf("unable to resolve: %s", err)
	}

	if list.L1CrossDomainMessenger.Version != "v1.6.0" {
		t.Fatalf("wrong L1CrossDomainMessenger version: %s", list.L1CrossDomainMessenger.Version)
	}
	if list.L1ERC721Bridge.Version != "v1.3.0" {
		t.Fatalf("wrong L1ERC721Bridge version: %s", list.L1ERC721Bridge.Version)
	}
	if list.L1StandardBridge.Version != "v1.3.0" {
		t.Fatalf("wrong L1StandardBridge version: %s", list.L1StandardBridge.Version)
	}
	if list.L2OutputOracle.Version != "v1.5.0" {
		t.Fatalf("wrong L2OutputOracle version: %s", list.L2OutputOracle.Version)
	}
	if list.OptimismMintableERC20Factory.Version != "v1.4.0" {
		t.Fatalf("wrong OptimismMintableERC20Factory version: %s", list.OptimismMintableERC20Factory.Version)
	}
	if list.OptimismPortal.Version != "v1.9.0" {
		t.Fatalf("wrong OptimismPortal version: %s", list.OptimismPortal.Version)
	}
	if list.SystemConfig.Version != "v1.7.0" {
		t.Fatalf("wrong SystemConfig version: %s", list.SystemConfig.Version)
	}

	// Check that fault proof contracts were not populated
	// The fault proof contracts should be skipped when L2OutputOracle is configured
	require.Equalf(t, VersionedContract{}, list.AnchorStateRegistry, "AnchorStateRegistry erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.DelayedWETH, "DelayedWETH erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.DisputeGameFactory, "DisputeGameFactory erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.FaultDisputeGame, "FaultDisputeGame erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.MIPS, "MIPS erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.PermissionedDisputeGame, "PermissionedDisputeGame erroneously configured with L2OutputOracle")
	require.Equalf(t, VersionedContract{}, list.PreimageOracle, "PreimageOracle erroneously configured with L2OutputOracle")
}

// TestContractVersionsResolveFaultProofContracts will test that resolve works with fault proof contracts.
func TestContractVersionsResolveFaultProofContracts(t *testing.T) {
	impls, err := newContractImplementations("sepolia")
	if err != nil {
		t.Fatalf("failed to load contract implementations: %v", err)
	}

	if impls.L1CrossDomainMessenger.Get("1.6.0") == (Address{}) {
		t.Fatal("wrong L1CrossDomainMessenger address")
	}
	if impls.L1ERC721Bridge.Get("1.3.0") == (Address{}) {
		t.Fatal("wrong L1ERC721Bridge address")
	}
	if impls.L1StandardBridge.Get("1.3.0") == (Address{}) {
		t.Fatal("wrong L1StandardBridge address")
	}
	if impls.L2OutputOracle.Get("1.5.0") == (Address{}) {
		t.Fatal("wrong L2OutputOracle address")
	}
	if impls.OptimismMintableERC20Factory.Get("1.4.0") == (Address{}) {
		t.Fatal("wrong OptimismMintableERC20 address")
	}
	if impls.OptimismPortal.Get("1.9.0") == (Address{}) {
		t.Fatal("wrong OptimismPortal address")
	}
	if impls.SystemConfig.Get("1.7.0") == (Address{}) {
		t.Fatal("wrong SystemConfig address")
	}

	// Configure fault proofs contracts, with no L2OutputOracle
	versions := ContractVersions{
		L1CrossDomainMessenger:       "1.6.0",
		L1ERC721Bridge:               "1.3.0",
		L1StandardBridge:             "1.3.0",
		OptimismMintableERC20Factory: "1.4.0",
		OptimismPortal:               "1.9.0",
		SystemConfig:                 "1.7.0",
		AnchorStateRegistry:          "1.0.0",
		DelayedWETH:                  "1.0.0",
		DisputeGameFactory:           "1.0.0",
		FaultDisputeGame:             "1.2.0",
		MIPS:                         "1.0.1",
		PermissionedDisputeGame:      "1.2.0",
		PreimageOracle:               "1.0.0",
	}

	list, err := impls.Resolve(versions)
	if err != nil {
		t.Fatalf("unable to resolve: %s", err)
	}

	if list.L1CrossDomainMessenger.Version != "v1.6.0" {
		t.Fatalf("wrong L1CrossDomainMessenger version: %s", list.L1CrossDomainMessenger.Version)
	}
	if list.L1ERC721Bridge.Version != "v1.3.0" {
		t.Fatalf("wrong L1ERC721Bridge version: %s", list.L1ERC721Bridge.Version)
	}
	if list.L1StandardBridge.Version != "v1.3.0" {
		t.Fatalf("wrong L1StandardBridge version: %s", list.L1StandardBridge.Version)
	}
	if list.OptimismMintableERC20Factory.Version != "v1.4.0" {
		t.Fatalf("wrong OptimismMintableERC20Factory version: %s", list.OptimismMintableERC20Factory.Version)
	}
	if list.OptimismPortal.Version != "v1.9.0" {
		t.Fatalf("wrong OptimismPortal version: %s", list.OptimismPortal.Version)
	}
	if list.SystemConfig.Version != "v1.7.0" {
		t.Fatalf("wrong SystemConfig version: %s", list.SystemConfig.Version)
	}
	// Check fault proof contracts
	if list.AnchorStateRegistry.Version != "v1.0.0" {
		t.Fatalf("wrong AnchorStateRegistry version: %s", list.AnchorStateRegistry.Version)
	}
	if list.DelayedWETH.Version != "v1.0.0" {
		t.Fatalf("wrong DelayedWETH version: %s", list.DelayedWETH.Version)
	}
	if list.DisputeGameFactory.Version != "v1.0.0" {
		t.Fatalf("wrong DisputeGameFactory version: %s", list.DisputeGameFactory.Version)
	}
	if list.FaultDisputeGame.Version != "v1.2.0" {
		t.Fatalf("wrong FaultDisputeGame version: %s", list.FaultDisputeGame.Version)
	}
	if list.MIPS.Version != "v1.0.1" {
		t.Fatalf("wrong MIPS version: %s", list.MIPS.Version)
	}
	if list.PermissionedDisputeGame.Version != "v1.2.0" {
		t.Fatalf("wrong PermissionedDisputeGame version: %s", list.PermissionedDisputeGame.Version)
	}
	if list.PreimageOracle.Version != "v1.0.0" {
		t.Fatalf("wrong PreimageOracle version: %s", list.PreimageOracle.Version)
	}

	// Check output oracle was not populated
	require.Equalf(t, VersionedContract{}, list.L2OutputOracle, "L2OutputOracle erroneously configured with fault proof contracts")
}

// TestResolve ensures that resolve finds implementations that exactly match the requested version
func TestResolve(t *testing.T) {
	cases := []struct {
		name    string
		set     AddressSet
		version string
		expect  VersionedContract
	}{
		{
			name: "match singleton option",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
			},
			version: "v0.0.1",
			expect: VersionedContract{
				Version: "v0.0.1",
				Address: HexToAddress("0x01"),
			},
		},
		{
			name: "match first address",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v1.0.1": HexToAddress("0x03"),
			},
			version: "v0.0.1",
			expect: VersionedContract{
				Version: "v0.0.1",
				Address: HexToAddress("0x01"),
			},
		},
		{
			name: "match last address",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v1.0.1": HexToAddress("0x03"),
			},
			version: "v1.0.1",
			expect: VersionedContract{
				Version: "v1.0.1",
				Address: HexToAddress("0x02"),
			},
		},
		{
			name: "match middle address",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v1.0.1": HexToAddress("0x03"),
			},
			version: "v1.0.0",
			expect: VersionedContract{
				Version: "v1.0.0",
				Address: HexToAddress("0x02"),
			},
		},
		{
			name: "match first address (missing prefix)",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v1.0.1": HexToAddress("0x03"),
			},
			version: "0.0.1",
			expect: VersionedContract{
				Version: "v0.0.1",
				Address: HexToAddress("0x01"),
			},
		},
		{
			name: "match last address (missing prefix)",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v2.0.1": HexToAddress("0x03"),
			},
			version: "2.0.1",
			expect: VersionedContract{
				Version: "v2.0.1",
				Address: HexToAddress("0x03"),
			},
		},
		{
			name: "match middle address (missing prefix)",
			set: AddressSet{
				"v0.0.1": HexToAddress("0x01"),
				"v1.0.0": HexToAddress("0x02"),
				"v2.0.1": HexToAddress("0x03"),
			},
			version: "1.0.0",
			expect: VersionedContract{
				Version: "v1.0.0",
				Address: HexToAddress("0x02"),
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := resolve(test.set, test.version)
			if err != nil {
				t.Fatal(err)
			}
			if resolved != test.expect {
				t.Fatalf("wrong version: %s", resolved.Version)
			}
		})
	}
}

// TestResolveWithError ensures that resolve errors appropriately
func TestResolveWithError(t *testing.T) {
	cases := []struct {
		name        string
		set         AddressSet
		version     string
		expectError string
	}{
		{
			name: "Empty version supplied",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "",
			expectError: ErrEmptyVersion.Error(),
		},
		{
			name: "Semver with operator prefix",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "^0.0.1",
			expectError: "invalid semver",
		},
		{
			name: "Semver with operator and 'v' prefix",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "~v0.0.1",
			expectError: "invalid semver",
		},
		{
			name: "Semver with wildcard",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "2.0.x",
			expectError: "invalid semver",
		},
		{
			name: "Semver with wildcard and version",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "v2.0.x",
			expectError: "invalid semver",
		},
		{
			name: "No exact match",
			set: AddressSet{
				"v0.0.1":  HexToAddress("0x01"),
				"v2.0.1":  HexToAddress("0x02"),
				"v99.0.1": HexToAddress("0x03"),
			},
			version:     "2.0.0",
			expectError: "cannot resolve semver",
		},
		{
			name:        "No implementations available",
			set:         AddressSet{},
			version:     "2.0.0",
			expectError: "no implementations found",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			resolved, err := resolve(test.set, test.version)
			require.ErrorContains(t, err, test.expectError)
			require.Equal(t, VersionedContract{}, resolved)
		})
	}
}

// TestAddressSet ensures that the AddressSet.Get method works with
// both the "v" prefix and without the "v" prefix.
func TestAddressSet(t *testing.T) {
	set := AddressSet{
		"v1.0.0": HexToAddress("0x123"),
		"1.1.0":  HexToAddress("0x234"),
	}

	if set.Get("v1.0.0") != HexToAddress("0x123") {
		t.Fatal("wrong address")
	}
	if set.Get("1.0.0") != HexToAddress("0x123") {
		t.Fatal("wrong address")
	}

	if set.Get("v1.1.0") != HexToAddress("0x234") {
		t.Fatal("wrong address")
	}
	if set.Get("1.1.0") != HexToAddress("0x234") {
		t.Fatal("wrong address")
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
	rawYAML := `
name: Mickey Mouse
l1:
  chain_id: 314
  public_rpc: https://disney.com
  explorer: https://disneyscan.io

protocol_versions_addr: "0x252CbE9517F731C618961D890D534183822dcC8d"
superchain_config_addr: "0x02d91Cf852423640d93920BE0CAdceC0E7A00FA7"

canyon_time: 1
delta_time: 2
ecotone_time: 3
fjord_time:
`

	s := SuperchainConfig{}
	err := unMarshalSuperchainConfig([]byte(rawYAML), &s)
	require.NoError(t, err)

	require.Equal(t, "Mickey Mouse", s.Name)
	require.Equal(t, SuperchainL1Info{
		ChainID:   314,
		PublicRPC: "https://disney.com",
		Explorer:  "https://disneyscan.io",
	}, s.L1)
	require.Equal(t, "0x252cbe9517f731c618961d890d534183822dcc8d", s.ProtocolVersionsAddr.String())
	require.Equal(t, "0x02d91cf852423640d93920be0cadcec0e7a00fa7", s.SuperchainConfigAddr.String())
	require.Equal(t, uint64Ptr(uint64(1)), s.hardForkDefaults.CanyonTime)
	require.Equal(t, uint64Ptr(uint64(2)), s.hardForkDefaults.DeltaTime)
	require.Equal(t, uint64Ptr(uint64(3)), s.hardForkDefaults.EcotoneTime)
	require.Nil(t, s.hardForkDefaults.FjordTime)
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
	override := []byte(`canyon_time: 8`)
	nilOverride := []byte(`canyon_time:`)
	nilOverride2 := []byte(``)
	nilOverride3 := []byte(`superchain_time: 2`)
	nilOverride4 := []byte(`superchain_time: 0`)
	nilOverride5 := []byte(`superchain_time: 10`)

	type testCase struct {
		name               string
		scConfig           SuperchainConfig
		rawYAML            []byte
		expectedCanyonTime *uint64
	}

	testCases := []testCase{
		{"default + override  (nil superchain_time)= override", defaultSuperchainConfig, override, overridenCanyonTime},
		{"default + nil override (nil superchain_time) = nil", defaultSuperchainConfig, nilOverride, nil},
		{"default + no override (nil superchain_time )= nil", defaultSuperchainConfig, nilOverride2, nil},
		{"nil default + override = override", nilDefaultSuperchainConfig, override, overridenCanyonTime},
		{"nil default + nil override = nil", nilDefaultSuperchainConfig, nilOverride, nil},
		{"nil default + no override = nil", nilDefaultSuperchainConfig, nilOverride2, nil},
		{"default + nil override (default after superchain_time) = nil", defaultSuperchainConfig, nilOverride3, &defaultCanyonTime},
		{"default + nil override (default after zero superchain_time) = nil", defaultSuperchainConfig, nilOverride4, &defaultCanyonTime},
		{"default + nil override (default before superchain_time) = nil", defaultSuperchainConfig, nilOverride5, nil},
	}

	executeTestCase := func(t *testing.T, tt testCase) {
		c := ChainConfig{}

		err := yaml.Unmarshal([]byte(tt.rawYAML), &c)
		require.NoError(t, err)

		c.setNilHardforkTimestampsToDefault(&tt.scConfig)

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

	rawYAML := `
ecotone_time: 2
fjord_time: 3
`

	err := yaml.Unmarshal([]byte(rawYAML), &c)
	require.NoError(t, err)

	c.setNilHardforkTimestampsToDefault(&defaultSuperchainConfig)

	var nil64 *uint64

	require.Equal(t, nil64, c.CanyonTime)
	require.Equal(t, nil64, c.DeltaTime)
	require.Equal(t, uint64Ptr(uint64(2)), c.EcotoneTime)
	require.Equal(t, uint64Ptr(uint64(3)), c.FjordTime)
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}
