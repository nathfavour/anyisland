package cli

import (
	"testing"
)

func TestNeedsShell(t *testing.T) {
	cases := map[string]bool{
		"CGO_ENABLED=0 make build":                                      true,
		"CGO_ENABLED=0 go build -ldflags '-s -w' -o vibeaura ./cmd/vibeaura": true,
		"go build -o tool .":                                            false,
		"git submodule update --init --recursive":                       false,
	}
	for step, want := range cases {
		if got := needsShell(step); got != want {
			t.Fatalf("needsShell(%q) = %v, want %v", step, got, want)
		}
	}
}
