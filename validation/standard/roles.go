package standard

type (
	Resolutions map[string]map[string]string // e.g. "AddressManager": "owner()":  "ProxyAdmin"
	L1          struct {
		Universal      Resolutions `toml:"universal"`
		NonFaultProofs Resolutions `toml:"nonFaultProofs"`
		FaultProofs    Resolutions `toml:"FaultProofs"`
	}
	L2 struct {
		Universal Resolutions `toml:"universal"`
	}
)

func (r L1) GetResolutions(isFaultProofs bool) Resolutions {
	combinedResolutions := make(Resolutions)
	for k, v := range r.Universal {
		combinedResolutions[k] = v
	}
	if isFaultProofs {
		for k, v := range r.FaultProofs {
			combinedResolutions[k] = v
		}
	} else {
		for k, v := range r.NonFaultProofs {
			combinedResolutions[k] = v
		}
	}
	return combinedResolutions
}

type Roles struct {
	L1 L1 `toml:"l1"`
	L2 L2 `toml:"L2"`
}

type (
	Multisigs     struct{}
	MultisigRoles struct {
		L1          L1 `toml:"l1"`
		L2          L2 `toml:"l2"`
		KeyHandover struct {
			L1 L1
			L2 L2
		} `toml:"key-handover"`
	}
)
