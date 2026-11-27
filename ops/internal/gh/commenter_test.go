package gh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/require"
)

const testMarker = "<!--- marker -->"

func testComment(id int64, login, userType, body string, createdAt time.Time) *github.IssueComment {
	return &github.IssueComment{
		ID:        github.Ptr(id),
		Body:      github.Ptr(body),
		CreatedAt: &github.Timestamp{Time: createdAt},
		User: &github.User{
			Login: github.Ptr(login),
			Type:  github.Ptr(userType),
		},
	}
}

func TestSelectComment(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	withMarker := "report " + testMarker

	tests := []struct {
		name     string
		comments []*github.IssueComment
		author   string
		wantID   int64
	}{
		{
			name: "no comments",
		},
		{
			name: "no comment contains the marker",
			comments: []*github.IssueComment{
				testComment(1, "ci[bot]", "Bot", "unrelated", t0),
				testComment(2, "alice", "User", "also unrelated", t0),
			},
		},
		{
			name: "single match from a user",
			comments: []*github.IssueComment{
				testComment(1, "alice", "User", "unrelated", t0),
				testComment(2, "bob", "User", withMarker, t0),
			},
			wantID: 2,
		},
		{
			name: "bot preferred over a more recent user comment",
			comments: []*github.IssueComment{
				testComment(1, "ci[bot]", "Bot", withMarker, t0),
				testComment(2, "alice", "User", withMarker, t0.Add(time.Hour)),
			},
			wantID: 1,
		},
		{
			name: "most recent bot comment wins",
			comments: []*github.IssueComment{
				testComment(1, "ci[bot]", "Bot", withMarker, t0),
				testComment(2, "other[bot]", "Bot", withMarker, t0.Add(2*time.Hour)),
				testComment(3, "ci[bot]", "Bot", withMarker, t0.Add(time.Hour)),
			},
			wantID: 2,
		},
		{
			name: "author preferred over a more recent bot comment",
			comments: []*github.IssueComment{
				testComment(1, "reporter", "User", withMarker, t0),
				testComment(2, "ci[bot]", "Bot", withMarker, t0.Add(time.Hour)),
			},
			author: "reporter",
			wantID: 1,
		},
		{
			name: "author match ignores case",
			comments: []*github.IssueComment{
				testComment(1, "Reporter[bot]", "Bot", withMarker, t0),
				testComment(2, "ci[bot]", "Bot", withMarker, t0.Add(time.Hour)),
			},
			author: "reporter[bot]",
			wantID: 1,
		},
		{
			name: "most recent author comment wins",
			comments: []*github.IssueComment{
				testComment(1, "reporter", "User", withMarker, t0.Add(time.Hour)),
				testComment(2, "reporter", "User", withMarker, t0),
			},
			author: "reporter",
			wantID: 1,
		},
		{
			name: "falls back to a bot when the author has no comment",
			comments: []*github.IssueComment{
				testComment(1, "ci[bot]", "Bot", withMarker, t0),
				testComment(2, "alice", "User", withMarker, t0.Add(time.Hour)),
			},
			author: "reporter",
			wantID: 1,
		},
		{
			name: "equal timestamps resolved by highest id",
			comments: []*github.IssueComment{
				testComment(5, "ci[bot]", "Bot", withMarker, t0),
				testComment(7, "ci[bot]", "Bot", withMarker, t0),
				testComment(6, "ci[bot]", "Bot", withMarker, t0),
			},
			wantID: 7,
		},
		{
			name: "comment without a user",
			comments: []*github.IssueComment{
				{ID: github.Ptr(int64(1)), Body: github.Ptr(withMarker)},
			},
			wantID: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectComment(tt.comments, CommentSelector{Marker: testMarker, Author: tt.author})
			if tt.wantID == 0 {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tt.wantID, got.GetID())
		})
	}
}

type fakeGithub struct {
	t        *testing.T
	mu       sync.Mutex
	pages    [][]*github.IssueComment
	requests []string
	posted   []string
	edited   map[int64]string
}

func newFakeGithub(t *testing.T, pages ...[]*github.IssueComment) (*fakeGithub, *GithubCommenter) {
	fake := &fakeGithub{t: t, pages: pages, edited: make(map[int64]string)}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	client := github.NewClient(nil).WithAuthToken("test-token")
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client.BaseURL = baseURL

	return fake, NewGithubCommenter("test-org/test-repo", client)
}

func (f *fakeGithub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requests = append(f.requests, r.Method+" "+r.URL.Path)

	readBody := func() string {
		raw, err := io.ReadAll(r.Body)
		require.NoError(f.t, err)
		var comment github.IssueComment
		require.NoError(f.t, json.Unmarshal(raw, &comment))
		return comment.GetBody()
	}

	const listPath = "/repos/test-org/test-repo/issues/42/comments"
	var editID int64
	switch {
	case r.Method == http.MethodGet && r.URL.Path == listPath:
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			_, err := fmt.Sscanf(p, "%d", &page)
			require.NoError(f.t, err)
		}
		if page < len(f.pages) {
			next := fmt.Sprintf("http://%s%s?page=%d", r.Host, listPath, page+1)
			w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, next))
		}
		var comments []*github.IssueComment
		if page <= len(f.pages) {
			comments = f.pages[page-1]
		}
		if comments == nil {
			comments = []*github.IssueComment{}
		}
		require.NoError(f.t, json.NewEncoder(w).Encode(comments))
	case r.Method == http.MethodPost && r.URL.Path == listPath:
		f.posted = append(f.posted, readBody())
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	case r.Method == http.MethodPatch && scanEditPath(r.URL.Path, &editID):
		f.edited[editID] = readBody()
		_, _ = w.Write([]byte(`{}`))
	default:
		http.Error(w, `{"message":"Resource not accessible by integration"}`, http.StatusForbidden)
	}
}

func scanEditPath(path string, id *int64) bool {
	n, err := fmt.Sscanf(path, "/repos/test-org/test-repo/issues/comments/%d", id)
	return err == nil && n == 1
}

func TestGithubCommenterFindComment(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	fake, commenter := newFakeGithub(t,
		[]*github.IssueComment{
			testComment(1, "alice", "User", "hello", t0),
			testComment(2, "ci[bot]", "Bot", testMarker, t0.Add(time.Hour)),
		},
		[]*github.IssueComment{
			testComment(3, "ci[bot]", "Bot", testMarker, t0.Add(2*time.Hour)),
			testComment(4, "alice", "User", testMarker, t0.Add(3*time.Hour)),
		},
	)

	got, err := commenter.FindComment(context.Background(), 42, CommentSelector{Marker: testMarker})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, int64(3), got.GetID())
	require.Equal(t, []string{
		"GET /repos/test-org/test-repo/issues/42/comments",
		"GET /repos/test-org/test-repo/issues/42/comments",
	}, fake.requests)
}

func TestUpsertComment(t *testing.T) {
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	selector := CommentSelector{Marker: testMarker}

	t.Run("posts when there is no existing comment", func(t *testing.T) {
		fake, commenter := newFakeGithub(t, []*github.IssueComment{
			testComment(1, "alice", "User", "hello", t0),
		})

		edited, err := UpsertComment(context.Background(), commenter, 42, selector, "new "+testMarker)
		require.NoError(t, err)
		require.False(t, edited)
		require.Equal(t, []string{"new " + testMarker}, fake.posted)
		require.Empty(t, fake.edited)
		require.Equal(t, []string{
			"GET /repos/test-org/test-repo/issues/42/comments",
			"POST /repos/test-org/test-repo/issues/42/comments",
		}, fake.requests)
	})

	t.Run("edits the existing comment", func(t *testing.T) {
		fake, commenter := newFakeGithub(t, []*github.IssueComment{
			testComment(1, "ci[bot]", "Bot", "old "+testMarker, t0),
		})

		edited, err := UpsertComment(context.Background(), commenter, 42, selector, "new "+testMarker)
		require.NoError(t, err)
		require.True(t, edited)
		require.Empty(t, fake.posted)
		require.Equal(t, map[int64]string{1: "new " + testMarker}, fake.edited)
		require.Equal(t, []string{
			"GET /repos/test-org/test-repo/issues/42/comments",
			"PATCH /repos/test-org/test-repo/issues/comments/1",
		}, fake.requests)
	})

	t.Run("returns list errors", func(t *testing.T) {
		_, commenter := newFakeGithub(t)

		_, err := UpsertComment(context.Background(), commenter, 43, selector, "new")
		require.ErrorContains(t, err, "failed to find existing comment")
	})
}
