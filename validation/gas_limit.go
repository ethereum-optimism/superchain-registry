package validation

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/superchain-registry/superchain"
	"github.com/ethereum-optimism/superchain-registry/validation/internal/bindings"
	"github.com/ethereum-optimism/superchain-registry/validation/standard"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type getGasLimitFunc func(context.Context, common.Address, *ethclient.Client) (uint64, error)

func CheckGasLimit(chain *superchain.ChainConfig, l1Client *ethclient.Client, getGasLimitOverride *getGasLimitFunc) error {
	getGasLimit := getGasLimitDefault
	if getGasLimitOverride != nil {
		getGasLimit = *getGasLimitOverride
	}

	contractAddress, err := superchain.Addresses[chain.ChainID].AddressFor("SystemConfigProxy")
	if err != nil {
		return err
	}

	desiredParam := standard.Config.Params[chain.Superchain].SystemConfig.GasLimit
	actualParam, err := getGasLimit(context.Background(), common.Address(contractAddress), l1Client)
	if err != nil {
		return err
	}

	if !isIntWithinBounds(actualParam, desiredParam) {
		return fmt.Errorf("%d is not within bounds %d", actualParam, desiredParam)
	}
	return nil
}

func getGasLimitDefault(ctx context.Context, systemConfigAddr common.Address, client *ethclient.Client) (uint64, error) {
	callOpts := &bind.CallOpts{Context: ctx}
	systemConfig, err := bindings.NewSystemConfig(systemConfigAddr, client)
	if err != nil {
		return 0, err
	}
	return Retry(systemConfig.GasLimit)(callOpts)
}
