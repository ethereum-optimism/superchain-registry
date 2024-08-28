export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm

echo "Installing and selecting correct Node version"
nvm install $1
nvm use $1

echo "Running install command"
eval $2

echo "Installing and selecting correct go version"
go_version=$(grep -m 1 '^go ' go.mod | awk '{print $2}')

# Source the gvm script to load gvm functions into the shell
. ~/.gvm/scripts/gvm || exit 1
gvm install go${go_version} || exit 1
gvm use go${go_version} || exit 1

echo "Running op-node genesis l2 command"
eval "$3"
