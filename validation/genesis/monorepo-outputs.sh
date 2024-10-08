export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm

echo "Installing and selecting correct Node version..."
nvm install $1
nvm use $1

echo "Running install command..."
eval $2

echo "Installing and selecting correct go version..."
go_version=$(go mod edit -print | grep -m 1 '^go ' | awk '{print $2}')

# Source the gvm script to load gvm functions into the shell
# NOTE gvm is necessary because older versions of op-node used a
# library which is not compatible with newer versions of Go
# Running with go1.21 results in
# The version of quic-go you're using can't be built on Go 1.21 yet. For more details, please see https://github.com/quic-go/quic-go/wiki/quic-go-and-Go-versions."
. ~/.gvm/scripts/gvm
gvm install go${go_version}
gvm use go${go_version}

echo "Running l2 genesis creation command..."
eval "$3"
