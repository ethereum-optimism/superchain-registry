package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ethereum-optimism/optimism/op-deployer/pkg/deployer/state"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/gh"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/go-github/v68/github"
	"github.com/urfave/cli/v2"
)

var (
	L1RPCURLFlag = &cli.StringFlag{
		Name:     "l1-rpc-url",
		Usage:    "The URL of the L1 RPC endpoint.",
		EnvVars:  []string{"L1_RPC_URL"},
		Required: true,
	}
	PRURLFlag = &cli.StringFlag{
		Name:     "pr-url",
		Usage:    "URL to the pull request.",
		EnvVars:  []string{"PR_NUMBER", "CIRCLE_PULL_REQUEST"},
		Required: true,
	}
	GitSHAFlag = &cli.StringFlag{
		Name:     "git-sha",
		Usage:    "The git SHA of the commit being reported on.",
		EnvVars:  []string{"GIT_SHA", "CIRCLE_SHA1"},
		Required: true,
	}
	GithubTokenFlag = &cli.StringFlag{
		Name:     "github-token",
		Usage:    "The GitHub token to use for API requests.",
		EnvVars:  []string{"GITHUB_TOKEN"},
		Required: true,
	}
	GithubRepoFlag = &cli.StringFlag{
		Name:     "github-repo",
		Usage:    "The GitHub repository to comment on.",
		EnvVars:  []string{"GITHUB_REPO"},
		Value:    "ethereum-optimism/superchain-registry",
		Required: true,
	}
)

func main() {
	app := &cli.App{
		Name:  "print-staging-report",
		Usage: "Prints a standards compliance report for the Standard Blockspace Charter.",
		Flags: []cli.Flag{
			L1RPCURLFlag,
			PRURLFlag,
			GitSHAFlag,
			GithubTokenFlag,
			GithubRepoFlag,
		},
		Action: PrintStagingReport,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func PrintStagingReport(cliCtx *cli.Context) error {
	l1RPCURL := cliCtx.String(L1RPCURLFlag.Name)
	prURL := cliCtx.String(PRURLFlag.Name)
	gitSHA := cliCtx.String(GitSHAFlag.Name)
	githubToken := cliCtx.String(GithubTokenFlag.Name)
	githubRepo := cliCtx.String(GithubRepoFlag.Name)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := paths.RequireRoot(wd); err != nil {
		return fmt.Errorf("root directory error: %w", err)
	}

	rpcClient, err := rpc.Dial(l1RPCURL)
	if err != nil {
		return fmt.Errorf("failed to dial RPC client: %w", err)
	}

	var meta config.StagingMetadata
	if err := paths.ReadTOMLFile(path.Join(paths.StagingDir(wd), "meta.toml"), &meta); err != nil {
		return fmt.Errorf("failed to read meta.toml: %w", err)
	}

	if meta.DeploymentTxHash == nil {
		return fmt.Errorf("deployment tx hash is required")
	}

	var st state.State
	if err := paths.ReadJSONFile(path.Join(paths.StagingDir(wd), "state.json"), &st); err != nil {
		return fmt.Errorf("failed to read state.json: %w", err)
	}

	intent := st.AppliedIntent
	if intent == nil {
		return fmt.Errorf("no intent found in state.json")
	}

	var params validation.ConfigParams
	if err := paths.ReadTOMLFile(paths.ValidationsFile(wd, meta.Superchain), &params); err != nil {
		return fmt.Errorf("failed to read standard params: %w", err)
	}

	ctx, cancel := context.WithTimeout(cliCtx.Context, 5*time.Minute)
	defer cancel()

	allReport := report.ScanAll(ctx, rpcClient, *meta.DeploymentTxHash, intent.L1ContractsLocator, &st)
	output.WriteOK("scanned L1 and L2")

	ghClient := github.NewClient(nil).WithAuthToken(githubToken)

	currUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get authenticated GitHub user: %w", err)
	}

	contractsVersion := validation.Semver(intent.L1ContractsLocator.Tag)
	stdPrestate := validation.StandardPrestates[contractsVersion]
	stdVersions := validation.StandardVersionsMainnet[contractsVersion]
	var stdConfigs validation.ConfigParams
	var stdRoles validation.RolesConfig
	switch allReport.L1.DeploymentChainID {
	case 1:
		stdConfigs = validation.StandardConfigParamsMainnet
		stdRoles = validation.StandardConfigRolesMainnet
	case 11155111:
		stdConfigs = validation.StandardConfigParamsSepolia
		stdRoles = validation.StandardConfigRolesSepolia
	default:
		return fmt.Errorf("unsupported chain ID: %d", allReport.L1.DeploymentChainID)
	}

	comment, err := report.RenderComment(
		&allReport,
		stdConfigs,
		stdRoles,
		stdPrestate,
		stdVersions,
		gitSHA,
	)
	if err != nil {
		return fmt.Errorf("failed to render comment: %w", err)
	}

	prNum, err := gh.GetPRNumberFromURL(prURL)
	if err != nil {
		return fmt.Errorf("failed to get PR number from URL: %w", err)
	}

	commenter := gh.NewGithubCommenter(githubRepo, ghClient)

	existingComment, err := commenter.FindComment(
		ctx,
		prNum,
		gh.CommentContaining(report.CommentMagic),
		gh.CommentFrom(currUser.GetLogin()),
	)
	if err != nil {
		return fmt.Errorf("failed to find existing comment: %w", err)
	}

	if existingComment == nil {
		if err := commenter.PostComment(ctx, prNum, comment); err != nil {
			return fmt.Errorf("failed to post comment: %w", err)
		}
		output.WriteOK("posted comment on PR %d", prNum)
	} else {
		if err := commenter.EditComment(ctx, existingComment.GetID(), comment); err != nil {
			return fmt.Errorf("failed to edit comment: %w", err)
		}
		output.WriteOK("edited comment on PR %d", prNum)
	}

	return nil
}
