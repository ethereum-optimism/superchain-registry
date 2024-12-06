package config

import (
	"fmt"
	"reflect"

	"github.com/ethereum-optimism/superchain-registry/superchain"
)

type LegacyPlasma struct {
	DAChallengeAddress         *superchain.Address `json:"da_challenge_address"`
	DAChallengeContractAddress *superchain.Address `json:"da_challenge_contract_address"`
	DACommitmentType           *string             `json:"da_commitment_type"`
	DAChallengeWindow          *uint64             `json:"da_challenge_window"`
	DAResolveWindow            *uint64             `json:"da_resolve_window"`
	UsePlasma                  *bool               `json:"use_plasma"`
	PlasmaConfig               *interface{}        `json:"plasma_config"`
}

func (lp *LegacyPlasma) CheckNonNilFields() error {
	v := reflect.ValueOf(*lp)
	typeOfV := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			fieldName := typeOfV.Field(i).Name
			return fmt.Errorf("field %s is non-nil", fieldName)
		}
	}

	return nil
}
