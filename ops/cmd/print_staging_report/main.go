package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/ethereum-optimism/superchain-registry/ops/internal/config"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/gh"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/manage"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/output"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/paths"
	"github.com/ethereum-optimism/superchain-registry/ops/internal/report"
	"github.com/ethereum-optimism/superchain-registry/validation"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/go-github/v68/github"
	"github.com/urfave/cli/v2"
)

var (
	SepoliaRPCURLFlag = &cli.StringFlag{
		Name:     "sepolia-rpc-url",
		Usage:    "The URL of the Sepolia RPC endpoint.",
		EnvVars:  []string{"SEPOLIA_RPC_URL"},
		Required: true,
	}
	MainnetRPCURLFlag = &cli.StringFlag{
		Name:     "mainnet-rpc-url",
		Usage:    "The URL of the mainnet RPC endpoint.",
		EnvVars:  []string{"MAINNET_RPC_URL"},
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
		Name:    "github-token",
		Usage:   "The GitHub token to use for API requests.",
		EnvVars: []string{"GITHUB_TOKEN"},
	}
	GithubRepoFlag = &cli.StringFlag{
		Name:    "github-repo",
		Usage:   "The GitHub repository to comment on.",
		EnvVars: []string{"GITHUB_REPO"},
		Value:   "ethereum-optimism/superchain-registry",
	}
	GithubCommentAuthorFlag = &cli.StringFlag{
		Name:    "comment-author",
		Usage:   "The login of the account that posts the report comment. Only comments from this account or from bot accounts are updated.",
		EnvVars: []string{"GITHUB_COMMENT_AUTHOR"},
	}
	DeployerCacheDirFlag = &cli.StringFlag{
		Name:    "deployer-cache-dir",
		Usage:   "The path to the op-deployer binaries cache directory.",
		EnvVars: []string{"DEPLOYER_CACHE_DIR"},
	}
)

func main() {
	app := &cli.App{
		Name:  "print-staging-report",
		Usage: "Prints a standards compliance report for the Standard Blockspace Charter.",
		Flags: []cli.Flag{
			SepoliaRPCURLFlag,
			MainnetRPCURLFlag,
			PRURLFlag,
			GitSHAFlag,
			GithubTokenFlag,
			GithubRepoFlag,
			GithubCommentAuthorFlag,
			DeployerCacheDirFlag,
		},
		Action: PrintStagingReport,
	}
	if err := app.Run(os.Args); err != nil {
		output.WriteStderr("%v", err)
		os.Exit(1)
	}
}

func PrintStagingReport(cliCtx *cli.Context) error {
	prURL := cliCtx.String(PRURLFlag.Name)
	gitSHA := cliCtx.String(GitSHAFlag.Name)
	githubToken := cliCtx.String(GithubTokenFlag.Name)
	githubRepo := cliCtx.String(GithubRepoFlag.Name)
	commentAuthor := cliCtx.String(GithubCommentAuthorFlag.Name)
	deployerCacheDir := cliCtx.String(DeployerCacheDirFlag.Name)

	wd, err := paths.FindRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := paths.RequireRoot(wd); err != nil {
		return fmt.Errorf("root directory error: %w", err)
	}

	stagedChainCfgs, err := manage.StagedChainConfigs(wd)
	if errors.Is(err, manage.ErrNoStagedConfig) {
		output.WriteOK("no staged chain config found, exiting")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get staged chain config: %w", err)
	}
	if len(stagedChainCfgs) > 1 {
		output.WriteWarn("multiple staged chain configs found, only the first one will be used")
	}

	chainCfg := stagedChainCfgs[0]
	var stdConfigs validation.ConfigParams
	var stdRoles validation.RolesConfig
	var l1RPCURL string
	switch chainCfg.Superchain {
	case config.MainnetSuperchain:
		stdConfigs = validation.StandardConfigParamsMainnet
		stdRoles = validation.StandardConfigRolesMainnet
		l1RPCURL = cliCtx.String(MainnetRPCURLFlag.Name)
	case config.SepoliaSuperchain:
		stdConfigs = validation.StandardConfigParamsSepolia
		stdRoles = validation.StandardConfigRolesSepolia
		l1RPCURL = cliCtx.String(SepoliaRPCURLFlag.Name)
	default:
		output.WriteWarn("skipping staging report for chain %s in unsupported superchain: %s",
			chainCfg.ShortName, chainCfg.Superchain)
		return nil
	}

	if chainCfg.DeploymentTxHash == nil {
		return fmt.Errorf("deployment tx hash is required")
	}

	contractsVersion := validation.Semver(chainCfg.DeploymentL1ContractsVersion)
	stdPrestate := validation.StandardPrestates.StablePrestate()
	stdVersions := validation.StandardVersionsMainnet[contractsVersion]

	rpcClient, err := rpc.Dial(l1RPCURL)
	if err != nil {
		return fmt.Errorf("failed to dial RPC client: %w", err)
	}

	var params validation.ConfigParams
	if err := paths.ReadTOMLFile(paths.ValidationsFile(wd, chainCfg.Superchain), &params); err != nil {
		return fmt.Errorf("failed to read standard params: %w", err)
	}

	ctx, cancel := context.WithTimeout(cliCtx.Context, 5*time.Minute)
	defer cancel()

	statePath := path.Join(paths.StagingDir(wd), "state.json")
	allReport := report.ScanAll(ctx, wd, l1RPCURL, rpcClient, statePath, chainCfg, deployerCacheDir)
	output.WriteOK("scanned L1 and L2")

	comment, err := report.RenderComment(
		&allReport,
		stdConfigs,
		stdRoles,
		stdPrestate,
		stdVersions,
		gitSHA,
		chainCfg.ShortName,
	)
	if err != nil {
		return fmt.Errorf("failed to render comment: %w", err)
	}

	if githubToken == "" {
		output.WriteOK("skipping GitHub comment, no token provided")
	} else {
		if err := postGithubComment(ctx, prURL, githubRepo, githubToken, commentAuthor, comment); err != nil {
			output.WriteNotOK("failed to post comment: %v", err)
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "%s\n", comment)
	return nil
}

func postGithubComment(ctx context.Context, prURL string, githubRepo string, ghToken string, commentAuthor string, comment string) error {
	ghClient := github.NewClient(nil).WithAuthToken(ghToken)

	prNum, err := gh.GetPRNumberFromURL(prURL)
	if err != nil {
		return fmt.Errorf("failed to get PR number from URL: %w", err)
	}

	commenter := gh.NewGithubCommenter(githubRepo, ghClient)
	selector := gh.CommentSelector{
		Marker: report.CommentMagic,
		Author: commentAuthor,
	}

	edited, err := gh.UpsertComment(ctx, commenter, prNum, selector, comment)
	if err != nil {
		return err
	}

	if edited {
		output.WriteOK("edited comment on PR %d", prNum)
	} else {
		output.WriteOK("posted comment on PR %d", prNum)
	}
	return nil
}
