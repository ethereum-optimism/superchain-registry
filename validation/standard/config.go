package standard

// Config is keyed by superchain target, e.g. "mainnet" or "sepolia" or "sepolia-dev-0"
var Config map[string]ConfigType

type ConfigType struct {
	*Params
	*Roles
}
