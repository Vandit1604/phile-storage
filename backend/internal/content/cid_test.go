package content

import "testing"

func TestComputeIsDeterministic(t *testing.T) {
	data := []byte("the address is the fingerprint")

	a, err := Compute(data)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	b, err := Compute(data)
	if err != nil {
		t.Fatalf("compute again: %v", err)
	}
	if !a.Equals(b) {
		t.Fatalf("same bytes produced different CIDs: %s vs %s", a, b)
	}
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name  string
		bytes []byte
		want  bool
	}{
		{"matching bytes", []byte("trustless"), true},
		{"tampered bytes", []byte("tampered!"), false},
		{"empty vs original", nil, false},
	}

	original := []byte("trustless")
	c, err := Compute(original)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Verify(tt.bytes, c); got != tt.want {
				t.Fatalf("Verify(%q) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := Parse("not-a-cid"); err == nil {
		t.Fatal("expected error for malformed CID")
	}

	c, _ := Compute([]byte("round trip"))
	parsed, err := Parse(c.String())
	if err != nil {
		t.Fatalf("parse valid CID: %v", err)
	}
	if !parsed.Equals(c) {
		t.Fatalf("round trip mismatch: %s vs %s", parsed, c)
	}
}
