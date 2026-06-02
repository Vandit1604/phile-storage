package api

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/Vandit1604/phile-storage/internal/content"
	"github.com/ipfs/go-cid"
	"github.com/sahilm/fuzzy"
)

type searchResult struct {
	Filename string   `json:"filename"`
	CID      string   `json:"cid"`
	Holders  []string `json:"holders"`
}

// SearchHandler matches a query against this node's known files. A query that
// parses as a CID is treated as a direct content lookup; otherwise it is
// fuzzy-matched against filenames, best matches first.
func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "query 'q' is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	if c, err := content.Parse(q); err == nil {
		writeJSON(w, http.StatusOK, []searchResult{{
			Filename: s.nameForCID(ctx, c),
			CID:      c.String(),
			Holders:  s.holders(ctx, c),
		}})
		return
	}

	all, err := s.metadataStore.ListAll(ctx)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}

	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic ordering for equally-scored matches

	results := make([]searchResult, 0)
	for _, m := range fuzzy.Find(q, names) {
		entry := all[m.Str]
		results = append(results, searchResult{Filename: m.Str, CID: entry.CID, Holders: entry.Holders})
	}
	writeJSON(w, http.StatusOK, results)
}

// holders returns the known holders of c, including this node when it holds the
// block locally (DHT provider records can lag just after an upload).
func (s *Server) holders(ctx context.Context, c cid.Cid) []string {
	list, _ := s.metadataStore.GetHolders(ctx, c)
	if s.fileStore.HasBlock(c) && !contains(list, s.peerAddress) {
		list = append([]string{s.peerAddress}, list...)
	}
	return list
}

// nameForCID reverse-looks-up a filename for a CID from the local file view.
func (s *Server) nameForCID(ctx context.Context, c cid.Cid) string {
	all, _ := s.metadataStore.ListAll(ctx)
	target := c.String()
	for name, entry := range all {
		if entry.CID == target {
			return name
		}
	}
	return ""
}
