package paths

import (
	"fmt"
	"os"
	"path"
)

func StagingDir(wd string) string {
	return path.Join(wd, ".staging")
}

func SuperchainDir(wd string, name string) string {
	return path.Join(wd, "superchain", "configs", name)
}

func ChainConfig(wd string, name string, shortName string) string {
	return path.Join(SuperchainDir(wd, name), shortName+".toml")
}

func SuperchainConfig(wd string, name string) string {
	return path.Join(SuperchainDir(wd, name), "superchain.toml")
}

func Superchains(wd string) ([]string, error) {
	configsDir := path.Join(wd, "superchain", "configs")

	dir, err := os.ReadDir(configsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", configsDir, err)
	}

	var superchains []string
	for _, entry := range dir {
		if entry.IsDir() {
			superchains = append(superchains, entry.Name())
		}
	}
	return superchains, nil
}

func ExtraDir(wd string) string {
	return path.Join(wd, "superchain", "extra")
}

func GenesisFile(wd string, name string, shortName string) string {
	return path.Join(ExtraDir(wd), "genesis", name, shortName+".json.zst")
}

func AddressesFile(wd string) string {
	return path.Join(ExtraDir(wd), "addresses", "addresses.json")
}

func ValidationsDir(wd string) string {
	return path.Join(wd, "validation", "standard")
}

func ValidationsFile(wd string, superchain string) string {
	return path.Join(ValidationsDir(wd), fmt.Sprintf("standard-config-params-%s.toml", superchain))
}

func RequireDir(p string) error {
	stat, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", p, err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", p)
	}

	return nil
}

func RequireRoot(wd string) error {
	p := StagingDir(wd)
	if err := RequireDir(p); err != nil {
		return fmt.Errorf("not at repo root or IO error: %w", err)
	}
	return nil
}
