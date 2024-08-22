set -e

echo "Inferring and selecting correct Node version"
export NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh" # This loads nvm
nvm use $1

echo "Running install command"
eval $2

echo "Inferring and selecting correct go version"

go_version=$(grep -m 1 '^go ' go.mod | awk '{print $2}')

# Source the gvm script to load gvm functions into the shell
set +e
source ~/.gvm/scripts/gvm || exit 1
gvm install go${go_version} || exit 1
gvm use go${go_version} || exit 1
set -e


echo "Running op-node genesis l2 command"

eval "$3"
