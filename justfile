set positional-arguments
alias t := test-all
alias l := lint-all

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
	CODEGEN=true go run superchain/internal/codegen/main.go

# Lint all Go code
lint-all:
	golangci-lint run superchain/... validation/... add-chain/... --fix

# Test all Go code
test-all: test-add-chain test-superchain test-validation

# Test all Go code in the add-chain module
test-add-chain:
	# We separate the first test from the rest because it generates artifacts
	# Which need to exist before the remaining tests run.
	TEST_DIRECTORY=./add-chain go run gotest.tools/gotestsum@latest --format testname -- -count=1 -run TestAddChain_Main
	TEST_DIRECTORY=./add-chain/... go run gotest.tools/gotestsum@latest --format testname -- -count=1 -run '[^TestAddChain_Main]'

# Test all Go code in the superchain module
test-superchain: clean-add-chain
	TEST_DIRECTORY=./superchain go run gotest.tools/gotestsum@latest --format testname

# Unit test all Go code in the validation module, and do not run validation checks themselves
test-validation: clean-add-chain
	TEST_DIRECTORY=./validation go run gotest.tools/gotestsum@latest --format testname -- -run='[^TestValidation|^TestPromotion]'

# Runs validation checks for any chain whose config changed
validate-modified-chains REF:
  # Running validation checks only for chains whose config has changed:
  git diff --merge-base {{REF}} --name-only 'superchain/configs/*.toml' ':(exclude)superchain/**/superchain.toml' | xargs -r awk '/chain_id/ {print $3}' | xargs -I {} just validate {}
# Run validation checks for chains with a name or chain ID matching the supplied regex, example: just validate 10
validate CHAIN_ID:
	TEST_DIRECTORY=./validation go run gotest.tools/gotestsum@latest --format testname -- -run='TestValidation/.+\({{CHAIN_ID}}\)$' -count=1

# Run genesis validation (this is separated from other validation checks, because it is not a part of drift detection)
validate-genesis-allocs CHAIN_ID:
	TEST_DIRECTORY=./validation/genesis go run gotest.tools/gotestsum@latest --format standard-verbose -- -run='TestGenesisAllocs/.+\({{CHAIN_ID}}\)$' -timeout 0

promotion-test:
  TEST_DIRECTORY=./validation go run gotest.tools/gotestsum@latest --format dots -- -run Promotion

# Clean test files generated by the add-chain tooling
clean-add-chain:
	rm -f superchain/configs/sepolia/testchain_*.toml
	rm -f superchain/extra/sepolia/testchain_*.json.gz
	rm -rf -- validation/genesis/validation-inputs/*-test/

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
	rm superchain/configs/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.toml
	rm superchain/extra/genesis/{{SUPERCHAIN_TARGET}}/{{CHAIN}}.json.gz
	just codegen
