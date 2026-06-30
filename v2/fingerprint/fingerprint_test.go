package fingerprint

import "testing"

func TestShortIsStableHex10(t *testing.T) {
	a := Short()
	t.Logf("machine fingerprint = %s  (full: %s)", a, Full())
	if len(a) != 10 {
		t.Fatalf("expected 10 chars, got %d (%q)", len(a), a)
	}
	for _, c := range a {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex char in fingerprint: %q", a)
		}
	}
	if b := Short(); a != b {
		t.Fatalf("fingerprint not deterministic: %s != %s", a, b)
	}
}
