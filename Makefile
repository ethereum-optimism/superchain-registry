#!make
ROOT_DIR:=$(CURDIR)
include .env

### Adding a chain
.PHONY: add-chain
add-chain:
	go run ./add-chain
	# go run ./add-chain check-rollup-config
	mkdir -p ./superchain/extra/genesis/$(SUPERCHAIN_TARGET)
	cd $(MONOREPO_DIR) && go run ./op-chain-ops/cmd/registry-data \
		--l2-genesis=$(GENESIS_CONFIG) \
		--bytecodes-dir=$(ROOT_DIR)/superchain/extra/bytecodes \
		--output=$(ROOT_DIR)/superchain/extra/genesis/$(SUPERCHAIN_TARGET)/$(CHAIN_SHORT_NAME).json.gz
	make codegen

promote-to-standard:
	go run ./add-chain promote-to-standard --chain-id=${chain-id}

### Auto-generated files
codegen:
	go run superchain/internal/codegen/main.go
	node ./scripts/codegen.js

### Linting
lint-all:
	golangci-lint run superchain/... validation/... add-chain/... --fix

### Testing
test-all: test-add-chain test-superchain test-validation

test-add-chain:
# We separate the first test from the rest because it generates artifacts
# Which need to exist before the remaining tests run.
	go test ./add-chain/... -run TestAddChain_Main -v
	go test ./add-chain/... -run '[^TestAddChain_Main]' -v
	make clean-add-chain

test-superchain:
	go test ./superchain/... -v

test-validation:
	go test ./validation/... -v

### Cleaning
clean-add-chain:
	rm -f superchain/configs/sepolia/testchain_*.yaml
	rm -f superchain/extra/addresses/sepolia/testchain_*.json
	rm -f superchain/extra/genesis-system-configs/sepolia/testchain_*.json

### Tidying
tidy-all: tidy-add-chain tidy-superchain tidy-validation

tidy-add-chain:
	cd add-chain && go mod tidy

tidy-superchain:
	cd superchain && go mod tidy

tidy-validation:
	cd validation && go mod tidy

### Removing a chain, example: make remove-chain superchain_target=sepolia chain=mychain
remove-chain:
	rm superchain/configs/$(superchain_target)/$(chain).yaml
	rm superchain/extra/addresses/$(superchain_target)/$(chain).json
	rm superchain/extra/genesis-system-configs/$(superchain_target)/$(chain).json
	rm superchain/extra/genesis/$(superchain_target)/$(chain).json.gz
