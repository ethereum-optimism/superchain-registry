package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/klauspost/compress/zstd"
)

func main() {
	if err := mainErr(); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

type GenesisAccount struct {
	CodeHash common.Hash                                          `json:"codeHash,omitempty"`
	Storage  jsonutil.LazySortedJsonMap[common.Hash, common.Hash] `json:"storage,omitempty"`
	Balance  *hexutil.Big                                         `json:"balance,omitempty"`
	Nonce    uint64                                               `json:"nonce,omitempty"`
}

type CompressibleGenesis struct {
	Config     *params.ChainConfig                                        `json:"config"`
	Nonce      uint64                                                     `json:"nonce"`
	Timestamp  uint64                                                     `json:"timestamp"`
	ExtraData  []byte                                                     `json:"extraData"`
	GasLimit   uint64                                                     `json:"gasLimit"   gencodec:"required"`
	Difficulty *hexutil.Big                                               `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash                                                `json:"mixHash"`
	Coinbase   common.Address                                             `json:"coinbase"`
	Alloc      jsonutil.LazySortedJsonMap[common.Address, GenesisAccount] `json:"alloc"      gencodec:"required"`

	// These fields are used for consensus tests. Please don't use them
	// in actual genesis blocks.
	Number        uint64       `json:"number"`
	GasUsed       uint64       `json:"gasUsed"`
	ParentHash    common.Hash  `json:"parentHash"`
	BaseFee       *hexutil.Big `json:"baseFeePerGas"` // EIP-1559
	ExcessBlobGas *uint64      `json:"excessBlobGas"` // EIP-4844
	BlobGasUsed   *uint64      `json:"blobGasUsed"`   // EIP-4844

	// StateHash represents the genesis state, to allow instantiation of a chain with missing initial state.
	// Chains with history pruning, or extraordinarily large genesis allocation (e.g. after a regenesis event)
	// may utilize this to get started, and then state-sync the latest state, while still verifying the header chain.
	StateHash *common.Hash `json:"stateHash,omitempty"`
}

func mainErr() error {
	wd := "/Users/matthewslipper/dev/superchain-registry"

	superchains := []string{"mainnet", "sepolia", "sepolia-dev-0"}

	dictPath := path.Join(paths.ExtraDir(wd), "dictionary")
	dict, err := os.ReadFile(dictPath)
	if err != nil {
		return fmt.Errorf("failed to read dictionary: %w", err)
	}

	toRemove := map[string]struct{}{}

	for _, superchain := range superchains {
		genesisP := path.Join(wd, "superchain", "extra", "genesis", superchain)

		fmt.Println("walking genesis directory", genesisP)

		err := filepath.Walk(genesisP, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() || path.Ext(p) != ".gz" {
				fmt.Println("skipping", p)
				return nil
			}

			fmt.Println("processing genesis file", p)

			shortName := strings.TrimSuffix(path.Base(p), ".json.gz")
			genPath := p

			var compGen CompressibleGenesis
			genF, err := os.Open(genPath)
			if err != nil {
				return fmt.Errorf("error opening genesis file: %w", err)
			}
			gzr, err := gzip.NewReader(genF)
			if err != nil {
				return fmt.Errorf("error creating gzip reader: %w", err)
			}
			if err := json.NewDecoder(gzr).Decode(&compGen); err != nil {
				return fmt.Errorf("error decoding genesis file: %w", err)
			}
			gzr.Close()
			genF.Close()

			genesis := &core.Genesis{
				Nonce:         compGen.Nonce,
				Timestamp:     compGen.Timestamp,
				ExtraData:     compGen.ExtraData,
				GasLimit:      compGen.GasLimit,
				Difficulty:    compGen.Difficulty.ToInt(),
				Mixhash:       compGen.Mixhash,
				Coinbase:      compGen.Coinbase,
				Number:        compGen.Number,
				GasUsed:       compGen.GasUsed,
				ParentHash:    compGen.ParentHash,
				BaseFee:       compGen.BaseFee.ToInt(),
				ExcessBlobGas: compGen.ExcessBlobGas,
				BlobGasUsed:   compGen.BlobGasUsed,
				StateHash:     compGen.StateHash,
				Alloc:         make(types.GenesisAlloc),
			}

			for addr, compAlloc := range compGen.Alloc {
				bal := compAlloc.Balance
				if bal == nil {
					bal = new(hexutil.Big)
				}

				alloc := types.Account{
					Storage: make(map[common.Hash]common.Hash),
					Balance: bal.ToInt(),
					Nonce:   compAlloc.Nonce,
				}
				for k, v := range compAlloc.Storage {
					alloc.Storage[k] = v
				}
				if compAlloc.CodeHash != (common.Hash{}) {
					bcodeFile := path.Join(wd, "superchain", "extra", "bytecodes", compAlloc.CodeHash.String()+".bin.gz")
					bcf, err := os.Open(bcodeFile)
					if err != nil {
						return fmt.Errorf("error opening bytecode file: %w", err)
					}
					bcgzr, err := gzip.NewReader(bcf)
					if err != nil {
						return fmt.Errorf("error creating gzip reader: %w", err)
					}
					bcd, err := io.ReadAll(bcgzr)
					if err != nil {
						return fmt.Errorf("error reading bytecode file: %w", err)
					}
					bcgzr.Close()
					bcf.Close()
					alloc.Code = bcd
					toRemove[bcodeFile] = struct{}{}
				}
				genesis.Alloc[addr] = alloc
			}

			fmt.Println("dumping genesis")

			outF, err := os.OpenFile(
				paths.GenesisFile(wd, superchain, shortName),
				os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
				0644,
			)
			if err != nil {
				return fmt.Errorf("error opening output file: %w", err)
			}

			zstdW, err := zstd.NewWriter(outF, zstd.WithEncoderDict(dict))
			if err != nil {
				return fmt.Errorf("error creating zstd writer: %w", err)
			}
			if err := json.NewEncoder(zstdW).Encode(genesis); err != nil {
				return fmt.Errorf("error encoding genesis: %w", err)
			}
			if err := zstdW.Close(); err != nil {
				return fmt.Errorf("error closing zstd writer: %w", err)
			}
			outF.Close()
			toRemove[genPath] = struct{}{}

			return nil
		})
		if err != nil {
			return fmt.Errorf("error walking genesis directory: %w", err)
		}
	}

	for p := range toRemove {
		fmt.Println("removing", p)
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("error removing file: %w", err)
		}
	}

	return nil
}
