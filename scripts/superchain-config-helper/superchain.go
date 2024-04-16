package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
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

	// address := "0x229047fed2591dbec1eF1118d64F7aF3dB9EB290"
	// functionSignature := "resourceConfig()((uint32,uint8,uint8,uint32,uint32,uint128))"

	// cmd := exec.Command("cast", "call", address, functionSignature, "--rpc-url", *rpcURL)
	// output, err := cmd.Output()

	// if err != nil {
	// 	fmt.Printf("Error executing command: %v\n", err)
	// 	return
	// }

	// fmt.Println(string(output))

	// var superchainConfig SuperchainConfig
	// superchainConfig.getConf()
	// fmt.Println(superchainConfig.Name)

	// // url := "https://eth-mainnet.g.alchemy.com/v2/6eCsAcYZmnnoS-FV6PCfqmObgK0prq-u"
	// to := "0x6481ff79597fe4f77e1063f615ec5bdaddeffd4b"
	// calldata := "0xcc731b02" // cast calldata "resourceConfig()" -> 0xcc731b02

	// result, err := makeEthCallRequest(*rpcURL, to, calldata)
	// if err != nil {
	// 	log.Fatalf("Error making RPC request: %v", err)
	// }

	// fmt.Printf("Result: %+v\n", result)
}

func entrypoint(ctx *cli.Context) error {
	rpcURL := ctx.String("l1-rpc-url")
	network := ctx.String("network")

	chains := []string{"base", "mode", "op", "zora"}
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

		challengePeriod, _ := getEitherOnChainInteger(rpcURL, l1AddressesConfig.L2OutputOracleProxy, "0xce5db8d6", "0xf4daa291")
		challengePeriodRow = append(challengePeriodRow, challengePeriod)

		feeScalar, _ := getFeeScalar(rpcURL, l1AddressesConfig.SystemConfigProxy)
		feeScalarRow = append(feeScalarRow, "Fixed: "+feeScalar.FixedData+" Dynamic: "+feeScalar.DynamicData)

		gasLimitRow = append(gasLimitRow, "TODO")

		genesisStateRow = append(genesisStateRow, genesisConfig.Genesis.L2.Hash)

		l2BlockTime, _ := getEitherOnChainInteger(rpcURL, l1AddressesConfig.L2OutputOracleProxy, "0x93991af3", "0x002134cc")
		l2BlockTimeRow = append(l2BlockTimeRow, l2BlockTime+" seconds")
	}
	rows = append(rows, chainIdRow)
	rows = append(rows, batchInboxAddressRow)
	rows = append(rows, explorerUrlRow)
	rows = append(rows, challengePeriodRow)
	rows = append(rows, feeScalarRow)
	rows = append(rows, gasLimitRow)
	rows = append(rows, genesisStateRow)
	rows = append(rows, l2BlockTimeRow)
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

func getEitherOnChainInteger(rpcURL string, l2OutputOracleProxy string, functionSigOne string, functionSigTwo string) (string, error) {
	var value string
	var err error

	value, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, functionSigOne)
	if value == "0" || err != nil {
		value, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, functionSigTwo)
		if err != nil {
			return "", fmt.Errorf("there was an error fetching the on-chain integer: %w", err)
		}
	}
	return value, nil
}

// func getChallengePeriod(rpcURL string, l2OutputOracleProxy string) (string, error) {
// 	var challengePeriod string
// 	var err error

// 	challengePeriod, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, "0xce5db8d6") // finalizationPeriodSeconds()
// 	if challengePeriod == "0" || err != nil {
// 		challengePeriod, err = getOnChainInteger(rpcURL, l2OutputOracleProxy, "0xf4daa291") // FINALIZATION_PERIOD_SECONDS()
// 		if err != nil {
// 			return "", fmt.Errorf("there was an error fetching the challenge period: %w", err)
// 		}
// 	}
// 	return challengePeriod, nil
// }

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
	bigIntValue.SetString(value, 16) // 16 is the base for hexadecimal
	if err != nil {
		return "0", fmt.Errorf("failed to parse value as an integer: %w", err)
	}

	return bigIntValue.String(), nil
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
		//log.Fatalf("HTTP GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//log.Fatalf("Reading response body failed: %v", err)
	}

	err = unmarshalFunc(body, config)
	if err != nil {
		//log.Fatalf("Error parsing file (skipping): %s", url)
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
