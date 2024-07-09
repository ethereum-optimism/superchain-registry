set positional-arguments

# Adding a chain
add-chain:
	go run ./add-chain
	go run ./add-chain check-rollup-config
	go run ./add-chain compress-genesis
	go run ./add-chain check-genesis
	just codegen

# Promote a standard candidate chain to a standard chain, example: just promote-to-standard 10
promote-to-standard CHAIN:
	go run ./add-chain promote-to-standard --chain-id={{CHAIN}}
	just codegen

# Generate auto-generated files
codegen: clean-add-chain
	go run superchain/internal/codegen/main.go
	node ./scripts/codegen.js

# Lint all Go code
lint-all:
	golangci-lint run superchain/... validation/... add-chain/... --fix

# Test all Go code
test-all: test-add-chain test-superchain test-validation

# Test all Go code in the add-chain module
test-add-chain:
	# We separate the first test from the rest because it generates artifacts
	# Which need to exist before the remaining tests run.
	go test ./add-chain/... -run TestAddChain_Main -v
	go test ./add-chain/... -run '[^TestAddChain_Main]' -v

# Test all Go code in the superchain module
test-superchain:
	go test ./superchain/... -v

# Test all Go code in the validation module
test-validation:
	go test ./validation/... -v

# Run validation checks for chains with a name or chain ID matching the supplied regex, example: just validate 10
validate CHAIN:
	go test ./validation/... -v -run=TestValidation/{{CHAIN}}

# Clean test files generated by the add-chain tooling
clean-add-chain:
	rm -f superchain/configs/sepolia/testchain_*.yaml
	rm -f superchain/extra/addresses/sepolia/testchain_*.json
	rm -f superchain/extra/genesis-system-configs/sepolia/testchain_*.json

# Tidying all go.mod files
tidy-all: tidy-add-chain tidy-superchain tidy-validation

# Tidy the add-chain go.mod file
tidy-add-chain:
	cd add-chain && go mod tidy

# Tidy the superchain go.mod file
tidy-superchain:
	cd superchain && go mod tidy

# Tidy the validation go.mod file
tidy-validation:
	cd validation && go mod tidy

# Removing a chain, example: just remove-chain sepolia op
remove-chain SUPERCHAIN_TARGET CHAIN:
	rm superchain/configs/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.yaml
	rm superchain/extra/addresses/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.json
	rm superchain/extra/genesis-system-configs/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.json
	rm superchain/extra/genesis/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.json.gz
	just codegen
