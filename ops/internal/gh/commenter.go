package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

// CommentSelector identifies a previously posted comment by a marker string
// contained in its body. Author optionally names the login expected to have
// written it. Comments written by other non-bot accounts never match.
type CommentSelector struct {
	Marker string
	Author string
}

type Commenter interface {
	PostComment(ctx context.Context, pr int, message string) error
	EditComment(ctx context.Context, commentID int64, message string) error
	FindComment(ctx context.Context, pr int, selector CommentSelector) (*github.IssueComment, error)
}

type GithubCommenter struct {
	owner    string
	repo     string
	ghClient *github.Client
}

func NewGithubCommenter(repoFull string, ghClient *github.Client) *GithubCommenter {
	owner, repoName := SplitOrgRepo(repoFull)
	return &GithubCommenter{
		owner:    owner,
		repo:     repoName,
		ghClient: ghClient,
	}
}

func (gc *GithubCommenter) PostComment(ctx context.Context, pr int, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	_, _, err := gc.ghClient.Issues.CreateComment(
		ctx,
		gc.owner,
		gc.repo,
		pr,
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}

	return nil
}

func (gc *GithubCommenter) EditComment(ctx context.Context, commentID int64, message string) error {
	comment := &github.IssueComment{
		Body: &message,
	}

	_, _, err := gc.ghClient.Issues.EditComment(
		ctx,
		gc.owner,
		gc.repo,
		commentID,
		comment,
	)
	if err != nil {
		return fmt.Errorf("failed to edit comment: %w", err)
	}

	return nil
}

func (gc *GithubCommenter) FindComment(ctx context.Context, pr int, selector CommentSelector) (*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var all []*github.IssueComment
	for {
		comments, resp, err := gc.ghClient.Issues.ListComments(
			ctx,
			gc.owner,
			gc.repo,
			pr,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list comments: %w", err)
		}
		all = append(all, comments...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return SelectComment(all, selector), nil
}

// SelectComment returns the comment that best matches the selector, or nil if
// there is none. A comment is eligible only if it contains the marker and was
// written by the selector's author or by a bot account. Comments written by
// the selector's author rank first, then comments written by bot accounts.
// Within the same rank the most recent comment wins.
func SelectComment(comments []*github.IssueComment, selector CommentSelector) *github.IssueComment {
	var best *github.IssueComment
	bestRank := -1
	for _, comment := range comments {
		if comment == nil || !strings.Contains(comment.GetBody(), selector.Marker) {
			continue
		}
		rank := selector.rank(comment)
		if rank == 0 {
			continue
		}
		if best == nil || rank > bestRank || (rank == bestRank && isMoreRecent(comment, best)) {
			best = comment
			bestRank = rank
		}
	}
	return best
}

// UpsertComment edits the existing comment matching the selector, or posts a
// new comment if there is none. It reports whether an existing comment was
// edited.
func UpsertComment(ctx context.Context, commenter Commenter, pr int, selector CommentSelector, message string) (bool, error) {
	existing, err := commenter.FindComment(ctx, pr, selector)
	if err != nil {
		return false, fmt.Errorf("failed to find existing comment: %w", err)
	}

	if existing == nil {
		if err := commenter.PostComment(ctx, pr, message); err != nil {
			return false, err
		}
		return false, nil
	}

	if err := commenter.EditComment(ctx, existing.GetID(), message); err != nil {
		return true, err
	}
	return true, nil
}

func (s CommentSelector) rank(comment *github.IssueComment) int {
	user := comment.GetUser()
	switch {
	case s.Author != "" && strings.EqualFold(user.GetLogin(), s.Author):
		return 2
	case user.GetType() == "Bot":
		return 1
	default:
		return 0
	}
}

func isMoreRecent(a, b *github.IssueComment) bool {
	at, bt := a.GetCreatedAt().Time, b.GetCreatedAt().Time
	if !at.Equal(bt) {
		return at.After(bt)
	}
	return a.GetID() > b.GetID()
}
