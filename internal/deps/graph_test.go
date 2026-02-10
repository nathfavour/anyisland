package deps

import (
	"errors"
	"testing"
)

func TestDependencyGraph_CheckCircularity(t *testing.T) {
	tests := []struct {
		name    string
		deps    map[string][]string
		wantErr error
	}{
		{
			name: "No circularity",
			deps: map[string][]string{
				"A": {"B", "C"},
				"B": {"D"},
				"C": {"D"},
			},
			wantErr: nil,
		},
		{
			name: "Simple circularity",
			deps: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			wantErr: ErrCircularDependency,
		},
		{
			name: "Indirect circularity",
			deps: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			wantErr: ErrCircularDependency,
		},
		{
			name: "Self circularity",
			deps: map[string][]string{
				"A": {"A"},
			},
			wantErr: ErrCircularDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraph()
			for pkg, deps := range tt.deps {
				for _, dep := range deps {
					g.AddDependency(pkg, dep)
				}
			}
			if err := g.CheckCircularity(); !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckCircularity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDependencyGraph_Resolve(t *testing.T) {
	depsMap := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
		"D": {},
	}

	getDeps := func(pkg string) ([]string, error) {
		return depsMap[pkg], nil
	}

	g := NewGraph()
	resolved, err := g.Resolve("A", getDeps)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Order should be D, then B and C (in any order), then A
	// B and C could be swapped depending on map iteration order
	
	if len(resolved) != 4 {
		t.Errorf("Expected 4 resolved packages, got %d: %v", len(resolved), resolved)
	}
	
	if resolved[3] != "A" {
		t.Errorf("Expected A to be last, got %s", resolved[3])
	}
	if resolved[0] != "D" {
		t.Errorf("Expected D to be first, got %s", resolved[0])
	}
}
