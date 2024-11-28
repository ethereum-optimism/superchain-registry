# CircleCI Upload Chain Artifacts Script

This script allows you to trigger a CircleCI pipeline via API to upload chain artifacts. It's designed to streamline the process of starting a pipeline with the required parameters.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation and Setup](#installation-and-setup)
  - [Install the CircleCI CLI](#install-the-circleci-cli)
  - [Configure the CircleCI CLI](#configure-the-circleci-cli)
- [Usage](#usage)
  - [Syntax](#syntax)
  - [Examples](#examples)
- [Notes](#notes)
- [Troubleshooting](#troubleshooting)
- [Additional Resources](#additional-resources)
- [License](#license)

## Prerequisites

Before using the script, ensure you have the following:

- **CircleCI CLI Installed**: The script relies on the CircleCI CLI configuration file to obtain the API token.
- **CircleCI API Token**: A personal API token from CircleCI to authenticate API requests.
- **Access to the Repository**: Ensure you have the necessary permissions to trigger pipelines in the `ethereum-optimism/superchain-registry` repository.

## Installation and Setup

### Install the CircleCI CLI

If you haven't installed the CircleCI CLI, follow these steps:

1. Install via Homebrew (macOS):
   brew install circleci

2. Install via Script (Linux):
   curl -fLSs https://circle.ci/cli | bash

3. Verify Installation:
   circleci version

   You should see the version of the CircleCI CLI installed.

Refer to the [CircleCI CLI Installation Guide](https://circleci.com/docs/local-cli/#installation) for detailed instructions.

### Configure the CircleCI CLI

1. Run CircleCI Setup:
   circleci setup

2. Enter Your CircleCI API Token:
   - Log in to CircleCI.
   - Go to [Personal API Tokens](https://app.circleci.com/settings/user/tokens).
   - Click **Create New Token**, name it (e.g., `PipelineTriggerToken`), and click **Create Token**.
   - Copy the token and enter it when prompted.

3. Verify Setup:
   The configuration file will be saved to `~/.circleci/cli.yml`.

## Usage

Place the `circleci-upload-chain-artifacts.sh` script in your desired directory or in the `.circleci` folder of your repository.

### Make the Script Executable

chmod +x circleci-upload-chain-artifacts.sh

### Syntax

./circleci-upload-chain-artifacts.sh <chain> [branch]

- `<chain>`: **(Required)** The name of the target chain (e.g., `op-mainnet`).
- `[branch]`: **(Optional)** The branch to trigger the pipeline on. Defaults to `main`.

### Examples

Trigger Pipeline on Default Branch:
./circleci-upload-chain-artifacts.sh op-mainnet

Trigger Pipeline on Specific Branch:
./circleci-upload-chain-artifacts.sh op-mainnet develop

Interactive Chain Input:
If you run the script without providing the `<chain>` parameter, it will prompt you to enter it:

./circleci-upload-chain-artifacts.sh

Output:
Usage: ./circleci-upload-chain-artifacts.sh <chain> [branch]
Error: The 'chain' parameter is required.
Enter the chain name:

### Default Branch Override

By default, the script is set to trigger the pipeline on the `migrate-github-actions-to-circleci` branch.
To change this, modify the `BRANCH` variable in the script or provide a branch name as an argument.

## Notes

- **Script Location**: Place the script in the `.circleci` folder for better organization.
- **API Endpoint**: The script triggers pipelines for the `ethereum-optimism/superchain-registry` project.
- **Token Security**: Keep your CircleCI API token secure. Do not share it or commit it to version control.

## Troubleshooting

### CircleCI CLI Configuration File Not Found

Error:
CircleCI CLI configuration file not found at ~/.circleci/cli.yml.
Solution:
Install and set up the CircleCI CLI as described in the [Installation and Setup](#installation-and-setup) section.

### Token Not Found in Configuration File

Error:
Token not found in ~/.circleci/cli.yml.
Solution:
Run `circleci setup` to configure your API token.

### Failed to Trigger the Pipeline

Error:
Failed to trigger the pipeline. HTTP response code: <response_code>
Solution:
- Verify that your API token is correct and has the necessary permissions.
- Ensure that the `API_ENDPOINT`, `ORG_NAME`, and `PROJECT_NAME` in the script are correct.
- Check your network connection.

### Script Permissions

Error:
bash: ./circleci-upload-chain-artifacts.sh: Permission denied
Solution:
Make the script executable:
chmod +x circleci-upload-chain-artifacts.sh

## Additional Resources

- [CircleCI API Documentation](https://circleci.com/docs/api/v2/#operation/triggerPipeline)
- [CircleCI CLI Documentation](https://circleci.com/docs/local-cli)
- [Ethereum Optimism GitHub Repository](https://github.com/ethereum-optimism/superchain-registry)

## License

This script is provided under the [MIT License](../LICENSE).
