package standard

type (
	Resolutions map[string]map[string]string // e.g. "AddressManager": "owner()":  "ProxyAdmin"
	L1          struct {
		Universal Resolutions `toml:"universal"`
		NonFPAC   Resolutions `toml:"nonFPAC"`
		FPAC      Resolutions `toml:"FPAC"`
	}
	L2 struct {
		Universal Resolutions `toml:"universal"`
	}
)

func (r L1) GetResolutions(isFPAC bool) Resolutions {
	combinedResolutions := make(Resolutions)
	for k, v := range r.Universal {
		combinedResolutions[k] = v
	}
	if isFPAC {
		for k, v := range r.FPAC {
			combinedResolutions[k] = v
		}
	} else {
		for k, v := range r.NonFPAC {
			combinedResolutions[k] = v
		}
	}
	return combinedResolutions
}

type Roles struct {
	L1 L1 `toml:"l1"`
	L2 L2 `toml:"L2"`
}

type Multisigs struct {
}
type MultisigRoles struct {
	L1          L1 `toml:"l1"`
	L2          L2 `toml:"l2"`
	KeyHandover struct {
		L1 L1
		L2 L2
	} `toml:"key-handover"`
}
