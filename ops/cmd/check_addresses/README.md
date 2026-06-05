# Address Checksum Validator

This tool helps ensure that all Ethereum addresses in TOML files are properly checksummed.

## What's an Ethereum Checksum Address?

Ethereum addresses use a mixed-case checksum format (EIP-55) to help detect input errors. The correct format has both uppercase and lowercase letters, and the specific case pattern serves as a checksum. For example:
- `0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed` (correct)
- `0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed` (incorrect, all lowercase)

Using checksummed addresses helps prevent errors when importing them from Solidity or other systems.

## Usage

```
go run ./cmd/check_addresses/main.go [options]
```

### Options

- `-all`: Check all TOML files in the validation directory (default: only checks standard-versions-sepolia.toml)
- `-fix`: Fix incorrect addresses (default: only reports incorrect addresses)
- `-file=<path>`: Check a specific file (default: checks standard-versions-sepolia.toml)

### Examples

Check the default file:
```
go run ./cmd/check_addresses/main.go
```

Check all TOML files:
```
go run ./cmd/check_addresses/main.go -all
```

Check a specific file:
```
go run ./cmd/check_addresses/main.go -file=path/to/file.toml
```

Fix addresses in all TOML files:
```
go run ./cmd/check_addresses/main.go -all -fix
```

## CI Integration

This tool is integrated into the CI pipeline and will fail builds if it detects incorrectly checksummed addresses. 
