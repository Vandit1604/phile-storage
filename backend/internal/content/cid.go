// Package content turns raw bytes into content identifiers (CIDs).
//
// A CID is the fingerprint of some bytes: the same bytes always produce the
// same CID, and any change produces a different one. Addressing files by CID
// instead of by name gives us deduplication, immutability, and the ability to
// verify that bytes we received are the bytes we asked for.
package content

import (
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// rawCodec marks a CID as pointing at opaque bytes (not a DAG node).
const rawCodec = 0x55

var prefix = cid.Prefix{
	Version:  1,
	Codec:    rawCodec,
	MhType:   mh.SHA2_256,
	MhLength: -1,
}

// Compute returns the CID for a block of bytes.
func Compute(data []byte) (cid.Cid, error) {
	c, err := prefix.Sum(data)
	if err != nil {
		return cid.Undef, fmt.Errorf("compute cid: %w", err)
	}
	return c, nil
}

// ComputeReader drains r and returns both the bytes and their CID, so callers
// fetching from an untrusted source can verify before persisting.
func ComputeReader(r io.Reader) (cid.Cid, []byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("read block: %w", err)
	}
	c, err := Compute(data)
	if err != nil {
		return cid.Undef, nil, err
	}
	return c, data, nil
}

// Verify reports whether data hashes to want.
func Verify(data []byte, want cid.Cid) bool {
	got, err := Compute(data)
	if err != nil {
		return false
	}
	return got.Equals(want)
}

// Parse decodes a CID string, rejecting anything malformed before it is used
// as a lookup key or path segment.
func Parse(s string) (cid.Cid, error) {
	c, err := cid.Decode(s)
	if err != nil {
		return cid.Undef, fmt.Errorf("parse cid %q: %w", s, err)
	}
	return c, nil
}
