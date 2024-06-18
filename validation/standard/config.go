package standard

// Config.Params is keyed by superchain target, e.g. "mainnet" or "sepolia" or "sepolia-dev-0"
type ConfigType struct {
	Params map[string]*Params
	Roles  *Roles
}

var Config = ConfigType{}
