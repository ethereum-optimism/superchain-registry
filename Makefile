lint:
	golangci-lint run superchain/... validation/... add-chain/... --fix

codegen:
	go run superchain/internal/codegen/main.go
	pnpm codegen

test-all: test-add-chain test-superchain test-validation

clean-add-chain:
	rm -f superchain/configs/sepolia/testchain_*.yaml
	rm -f superchain/extra/addresses/sepolia/testchain_*.json
	rm -f superchain/extra/genesis-system-configs/sepolia/testchain_*.json

test-add-chain:
	go test ./add-chain/... -run TestAddChain_Main -v
	go test ./add-chain/... -run '[^TestAddChain_Main]' -v
	make clean-add-chain

test-superchain:
	go test ./superchain/... -v

test-validation:
	go test ./validation/... -v


