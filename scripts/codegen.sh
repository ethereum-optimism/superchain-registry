set -e

go run superchain/internal/codegen/main.go
pnpm codegen
