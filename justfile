set positional-arguments := true
set dotenv-load := true

alias t := test-all
alias l := lint-all

_lint DIRECTORY FLAGS='':
		cd {{ DIRECTORY }} && golangci-lint run {{FLAGS}}

# Lint all Go code
lint-all: (_lint 'ops' '--fix') (_lint 'validation' '--fix')

_go-test DIRECTORY:
    cd {{ DIRECTORY }} && go run gotest.tools/gotestsum@v1.12.0 --format testname ./...

_go-tidy DIRECTORY:
    cd {{ DIRECTORY }} && go mod tidy

# Test all Go code in the ops module
test-ops: (_go-test 'ops')

# Unit test all Go code in the validation module, and do not run validation checks themselves
test-validation: (_go-test 'validation')

# Test all Go code
test-all: test-ops test-validation

# Tidying all go.mod files
tidy-all: tidy-ops tidy-validation

# Tidy the ops go.mod file
tidy-ops: (_go-tidy 'ops')

# Tidy the validation go.mod file
tidy-validation: (_go-tidy 'validation')

_run_ops_bin bin flags='':
	cd ops && go run ./cmd/{{ bin }}/main.go {{ flags }}

apply-hardforks: (_run_ops_bin 'apply_hardforks')

sync-staging flags='': (_run_ops_bin 'sync_staging' flags)

check-staging: (_run_ops_bin 'sync_staging' '--check')

print-staging-report: (_run_ops_bin 'print_staging_report')

check-genesis-integrity: (_run_ops_bin 'check_genesis_integrity')

codegen: (_run_ops_bin 'codegen')

create-config SHORTNAME FILENAME:
	@just _run_ops_bin "create_config" "--shortname {{SHORTNAME}} --state-filename $(realpath {{FILENAME}})"

check-chainlist: (_run_ops_bin 'check_chainlist')

check-for-new-addresses:
    #!/usr/bin/env bash
    set -euo pipefail
    root_dir=$(git rev-parse --show-toplevel)
    addresses_path=$root_dir/superchain/extra/addresses/addresses.json
    hash_before=$(sha256sum $addresses_path)
    just codegen
    hash_after=$(sha256sum $addresses_path)
    if [ "$hash_before" != "$hash_after" ]; then
        echo -e "\033[31m\nError: $addresses_path has changed, please commit the new addresses.json file.\033[0m\n"
        echo -e "\033[31mFiles changed:\033[0m"
        git status --porcelain
        exit 1
    fi


