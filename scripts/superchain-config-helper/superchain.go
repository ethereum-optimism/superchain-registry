package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/tabwriter"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"

	"gopkg.in/yaml.v2"
)

func main() {
	color := isatty.IsTerminal(os.Stderr.Fd())
	oplog.SetGlobalLogHandler(log.NewTerminalHandler(os.Stderr, color))
	app := &cli.App{
		Name:  "superchain",
		Usage: "Summarize the configuration of each chain in the Superchain.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "l1-rpc-url",
				Value:   "https://eth-mainnet.g.alchemy.com/v2/6eCsAcYZmnnoS-FV6PCfqmObgK0prq-u",
				Usage:   "L1 RPC URL",
				EnvVars: []string{"L1_RPC_URL"},
			},
			&cli.StringFlag{
				Name:    "network",
				Value:   "mainnet",
				Usage:   "L1 RPC URL (mainnet or sepolia)",
				EnvVars: []string{"SUPERCHAIN_TARGET"},
			},
		},
		Action: entrypoint,
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error op-upgrade", "err", err)
	}
}

func entrypoint(ctx *cli.Context) error {
	rpcURL := ctx.String("l1-rpc-url")
	network := ctx.String("network")

	chains := []string{"base", "metal", "mode", "op", "zora"}
	genesisPath := "https://raw.githubusercontent.com/ethereum-optimism/superchain-registry/main/superchain/configs"             // Can't I just read locally?
	l1AddressesPath := "https://raw.githubusercontent.com/ethereum-optimism/superchain-registry/main/superchain/extra/addresses" // Can't I just read locally?

	tw := tabwriter.NewWriter(os.Stdout, 20, 0, 2, ' ', 0)
	caption := "\033[1mSuperchain Configuration\033[0m"
	rows := [][]string{}
	chainIdRow := []string{"\033[1mChain ID\033[0m"}
	batchInboxAddressRow := []string{"\033[1mBatch Inbox Address\033[0m"}
	explorerUrlRow := []string{"\033[1mExplorer URL\033[0m"}
	challengePeriodRow := []string{"\033[1mChallenge Period\033[0m"}
	feeScalarRow := []string{"\033[1mFee Scalar\033[0m"}
	gasLimitRow := []string{"\033[1mGas Limit\033[0m"}
	genesisStateRow := []string{"\033[1mGenesis State\033[0m"}
	l2BlockTimeRow := []string{"\033[1mL2 Block Time\033[0m"}

	maxResourceLimitRow := []string{"\033[1mMax Resource Limit\033[0m"}
	elasticityMultiplierRow := []string{"\033[1mElasticity Multiplier\033[0m"}
	baseFeeMaxChangeDenominatorRow := []string{"\033[1mBase Fee Max Change Denominator\033[0m"}
	minimumBaseFeeRow := []string{"\033[1mMinimum Base Fee\033[0m"}
	systemTxMaxGasRow := []string{"\033[1mSystem Tx Max Gas Row\033[0m"}
	maximumBaseFeeRow := []string{"\033[1mMaximum Base Fee Row\033[0m"}

	for _, chain := range chains {
		fmt.Printf("\nFetching for %s...\n", chain)
		genesisUrl := fmt.Sprintf("%s/%s/%s.yaml", genesisPath, network, chain)
		var genesisConfig GenesisConfig
		genesisConfig.loadGenesisConfig(genesisUrl)

		l1AddressesUrl := fmt.Sprintf("%s/%s/%s.json", l1AddressesPath, network, chain)
		var l1AddressesConfig L1AddressesConfig
		l1AddressesConfig.loadL1AddressesConfig(l1AddressesUrl)

		chainIdRow = append(chainIdRow, fmt.Sprint(genesisConfig.ChainID))
		batchInboxAddressRow = append(batchInboxAddressRow, genesisConfig.BatchInboxAddr)
		explorerUrlRow = append(explorerUrlRow, fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", genesisConfig.Explorer, genesisConfig.Explorer))

		challengePeriod, _ := getOnChainIntegerFallback(rpcURL, l1AddressesConfig.L2OutputOracleProxy, "0xce5db8d6", "0xf4daa291")
		challengePeriodRow = append(challengePeriodRow, challengePeriod)

		feeScalar, _ := getFeeScalar(rpcURL, l1AddressesConfig.SystemConfigProxy)
		feeScalarRow = append(feeScalarRow, "Fixed: "+feeScalar.FixedData+" Dynamic: "+feeScalar.DynamicData)

		gasLimit, _ := getGasLimit(rpcURL, l1AddressesConfig.SystemConfigProxy)
		gasLimitRow = append(gasLimitRow, gasLimit)

		genesisStateRow = append(genesisStateRow, genesisConfig.Genesis.L2.Hash)

		l2BlockTime, _ := getOnChainIntegerFallback(rpcURL, l1AddressesConfig.L2OutputOracleProxy, "0x93991af3", "0x002134cc")
		l2BlockTimeRow = append(l2BlockTimeRow, l2BlockTime+" seconds")

		resourceConfig, _ := getResourceConfig(rpcURL, l1AddressesConfig.SystemConfigProxy)
		maxResourceLimitRow = append(maxResourceLimitRow, resourceConfig.maxResourceLimit)
		elasticityMultiplierRow = append(elasticityMultiplierRow, resourceConfig.elasticityMultiplier)
		baseFeeMaxChangeDenominatorRow = append(baseFeeMaxChangeDenominatorRow, resourceConfig.baseFeeMaxChangeDenominator)
		minimumBaseFeeRow = append(minimumBaseFeeRow, resourceConfig.minimumBaseFee)
		systemTxMaxGasRow = append(systemTxMaxGasRow, resourceConfig.systemTxMaxGas)
		maximumBaseFeeRow = append(maximumBaseFeeRow, resourceConfig.maximumBaseFee)
	}
	rows = append(rows, chainIdRow)
	rows = append(rows, batchInboxAddressRow)
	rows = append(rows, explorerUrlRow)
	rows = append(rows, challengePeriodRow)
	rows = append(rows, feeScalarRow)
	rows = append(rows, gasLimitRow)
	rows = append(rows, genesisStateRow)
	rows = append(rows, l2BlockTimeRow)

	rows = append(rows, elasticityMultiplierRow)
	rows = append(rows, baseFeeMaxChangeDenominatorRow)
	rows = append(rows, maxResourceLimitRow)
	rows = append(rows, minimumBaseFeeRow)
	rows = append(rows, systemTxMaxGasRow)
	rows = append(rows, maximumBaseFeeRow)
	if caption != "" {
		fmt.Fprintf(tw, "\n\n%s:\n\n", caption)
	}
	fmt.Fprintln(tw, strings.Join(append([]string{"\033[1mParameter\033[0m"}, chains...), "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
	return nil
}

func getResourceConfig(rpcUrl string, systemConfigProxy string) (ResourceConfig, error) {
	functionSignature := "resourceConfig()((uint32,uint8,uint8,uint32,uint32,uint128))"

	cmd := exec.Command("cast", "call", systemConfigProxy, functionSignature, "--rpc-url", rpcUrl)
	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}

	input := string(output)
	input = strings.ReplaceAll(input, "(", "")
	input = strings.ReplaceAll(input, ")", "")
	values := strings.Split(input, ",")
	re := regexp.MustCompile(`\[.*?\]`)
	nonDigits := regexp.MustCompile(`\D`)

	return ResourceConfig{
		maxResourceLimit:            trimAllSpace(nonDigits.ReplaceAllString(re.ReplaceAllString((values[0]), ""), "")),
		elasticityMultiplier:        trimAllSpace(values[1]),
		baseFeeMaxChangeDenominator: trimAllSpace(values[2]),
		minimumBaseFee:              trimAllSpace(nonDigits.ReplaceAllString(re.ReplaceAllString(values[3], ""), "")),
		systemTxMaxGas:              trimAllSpace(nonDigits.ReplaceAllString(re.ReplaceAllString(values[4], ""), "")),
		maximumBaseFee:              trimAllSpace(nonDigits.ReplaceAllString(re.ReplaceAllString(values[5], ""), ""))}, err
}

func trimAllSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func getGasLimit(rpcURL string, systemConfigProxy string) (string, error) {
	gasLimit, err := getOnChainInteger(rpcURL, systemConfigProxy, "0xf68016b7") // gasLimit()
	return gasLimit, err
}

func getFeeScalar(rpcURL string, systemConfigProxy string) (FeeScalar, error) {
	fixedData, _ := getOnChainInteger(rpcURL, systemConfigProxy, "0x0c18c162") // overhead()
	fixedDataScientificNotation, _ := new(big.Float).SetString(fixedData)
	dynamicData, _ := getOnChainInteger(rpcURL, systemConfigProxy, "0xf45e65d8") // scalar()
	dynamicDataScientificNotation, _ := new(big.Float).SetString(dynamicData)

	return FeeScalar{
		FixedData:   fmt.Sprintf("%.2e", fixedDataScientificNotation),
		DynamicData: fmt.Sprintf("%.2e", dynamicDataScientificNotation),
	}, nil
}

func getOnChainIntegerFallback(rpcURL string, l2OutputOracleProxy string, funcSig string, fallbackFuncSig string) (string, error) {
	var value string
	var err error

	value, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, funcSig)
	if value == "0" || err != nil {
		value, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, fallbackFuncSig)
		if err != nil {
			return "", fmt.Errorf("there was an error fetching the on-chain integer: %w", err)
		}
	}
	return value, nil
}

func getOnChainInteger(rpcURL string, address string, signature string) (string, error) {
	result, err := makeEthCallRequest(rpcURL, address, signature)
	if err != nil {
		return "0", fmt.Errorf("failed to make Ethereum call: %w", err)
	}

	value, ok := result["result"].(string)
	if !ok {
		return "0", fmt.Errorf("failed to extract result from Ethereum call")
	}

	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		value = value[2:]
	}

	bigIntValue := new(big.Int)
	bigIntValue.SetString(value, 16)

	return bigIntValue.String(), nil
}

type ResourceConfig struct {
	maxResourceLimit            string
	elasticityMultiplier        string
	baseFeeMaxChangeDenominator string
	minimumBaseFee              string
	systemTxMaxGas              string
	maximumBaseFee              string
}

type FeeScalar struct {
	FixedData   string
	DynamicData string
}

type GenesisConfig struct {
	Name             string  `yaml:"name"`
	ChainID          int     `yaml:"chain_id"`
	PublicRPC        string  `yaml:"public_rpc"`
	SequencerRPC     string  `yaml:"sequencer_rpc"`
	Explorer         string  `yaml:"explorer"`
	SuperchainLevel  int     `yaml:"superchain_level"`
	SystemConfigAddr string  `yaml:"system_config_addr"`
	BatchInboxAddr   string  `yaml:"batch_inbox_addr"`
	Genesis          Genesis `yaml:"genesis"`
}

type Genesis struct {
	L1     Layer `yaml:"l1"`
	L2     Layer `yaml:"l2"`
	L2Time int64 `yaml:"l2_time"`
}

type Layer struct {
	Hash   string `yaml:"hash"`
	Number int    `yaml:"number"`
}

type L1AddressesConfig struct {
	AddressManager                    string `json:"AddressManager"`
	L1CrossDomainMessengerProxy       string `json:"L1CrossDomainMessengerProxy"`
	L1ERC721BridgeProxy               string `json:"L1ERC721BridgeProxy"`
	L1StandardBridgeProxy             string `json:"L1StandardBridgeProxy"`
	L2OutputOracleProxy               string `json:"L2OutputOracleProxy"`
	OptimismMintableERC20FactoryProxy string `json:"OptimismMintableERC20FactoryProxy"`
	OptimismPortalProxy               string `json:"OptimismPortalProxy"`
	ProxyAdmin                        string `json:"ProxyAdmin"`
	SystemConfigProxy                 string `json:"SystemConfigProxy"`
	ProxyAdminOwner                   string `json:"ProxyAdminOwner"`
	SystemConfigOwner                 string `json:"SystemConfigOwner"`
	Guardian                          string `json:"Guardian"`
	Challenger                        string `json:"Challenger"`
}

func loadConfig(url string, unmarshalFunc func([]byte, interface{}) error, config interface{}) interface{} {
	resp, err := http.Get(url)
	if err != nil {
		log.Crit("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Crit("Reading response body failed: %v", err)
	}

	err = unmarshalFunc(body, config)
	if err != nil {
		log.Crit("Error parsing file (skipping): %s", url)
	}
	return config
}

func (c *GenesisConfig) loadGenesisConfig(url string) *GenesisConfig {
	return loadConfig(url, yaml.Unmarshal, c).(*GenesisConfig)
}

func (c *L1AddressesConfig) loadL1AddressesConfig(url string) *L1AddressesConfig {
	return loadConfig(url, json.Unmarshal, c).(*L1AddressesConfig)
}

type RPCRequest struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type EthCallRequest struct {
	From interface{} `json:"from"`
	To   string      `json:"to"`
	Data string      `json:"data"`
}

func makeEthCallRequest(url, to, calldata string) (map[string]interface{}, error) {
	payload := RPCRequest{
		Method: "eth_call",
		Params: []interface{}{
			EthCallRequest{
				From: nil,
				To:   to,
				Data: calldata,
			},
			"latest",
		},
		ID:      1,
		Jsonrpc: "2.0",
	}

	return makeRPCRequest(url, payload)
}

func makeRPCRequest(url string, payload RPCRequest) (map[string]interface{}, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
