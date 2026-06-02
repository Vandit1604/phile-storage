package p2p

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Vandit1604/phile-storage/internal/content"
)

func chdirTemp(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// TestFetchVerifies proves the trustless property: a node accepts bytes that
// hash to the requested CID and rejects bytes that do not.
func TestFetchVerifies(t *testing.T) {
	chdirTemp(t)
	ctx := context.Background()

	block := []byte("decentralized payload")
	good, err := content.Compute(block)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	wrong, err := content.Compute([]byte("different payload"))
	if err != nil {
		t.Fatalf("compute wrong: %v", err)
	}

	// The server returns `block` for ANY requested CID — honest for `good`,
	// dishonest for `wrong`.
	server, err := NewNode(ctx, "server", 0, func(string) ([]byte, error) {
		return block, nil
	})
	if err != nil {
		t.Fatalf("server node: %v", err)
	}
	defer server.Close()

	client, err := NewNode(ctx, "client", 0, func(string) ([]byte, error) {
		return nil, fmt.Errorf("client holds nothing")
	})
	if err != nil {
		t.Fatalf("client node: %v", err)
	}
	defer client.Close()

	got, err := client.Fetch(ctx, good, server.AddrInfo())
	if err != nil {
		t.Fatalf("honest fetch failed: %v", err)
	}
	if string(got) != string(block) {
		t.Fatalf("fetched wrong bytes: %q", got)
	}

	if _, err := client.Fetch(ctx, wrong, server.AddrInfo()); err == nil {
		t.Fatal("expected integrity check to reject mismatched bytes")
	}
}
