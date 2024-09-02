export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm

echo "Installing and selecting correct Node version"
nvm install $1
nvm use $1

echo "Running install command"
eval $2

echo "Running op-node genesis l2 command"
eval "$3"
