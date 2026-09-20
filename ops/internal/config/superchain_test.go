package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestFindValidL1URLStopsOnForbidden(t *testing.T) {
	forbidden := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	t.Cleanup(forbidden.Close)

	var fallbackCalled atomic.Bool
	fallback := newChainIDServer(t, 1, func() { fallbackCalled.Store(true) })

	_, err := FindValidL1URL(
		context.Background(),
		log.NewLogger(log.DiscardHandler()),
		[]string{forbidden.URL, fallback.URL},
		1,
	)

	require.ErrorContains(t, err, "l1-rpc-url request forbidden at index 0")
	require.ErrorContains(t, err, "403 Forbidden")
	require.False(t, fallbackCalled.Load(), "a forbidden response must stop URL discovery")
}

func TestFindValidL1URLContinuesAfterChainIDMismatch(t *testing.T) {
	mismatch := newChainIDServer(t, 1, nil)
	match := newChainIDServer(t, 2, nil)

	url, err := FindValidL1URL(
		context.Background(),
		log.NewLogger(log.DiscardHandler()),
		[]string{mismatch.URL, match.URL},
		2,
	)

	require.NoError(t, err)
	require.Equal(t, match.URL, url)
}

func newChainIDServer(t *testing.T, chainID uint64, onRequest func()) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest()
		}

		var request struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode JSON-RPC request: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Result  string          `json:"result"`
		}{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result:  fmt.Sprintf("0x%x", chainID),
		}); err != nil {
			t.Errorf("encode JSON-RPC response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
