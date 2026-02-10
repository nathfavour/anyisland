package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeDiscretion(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		readme   string
		expected bool
	}{
		{
			name:     "Functional tool",
			files:    []string{"main.go", "go.mod", "utils.go", "README.md"},
			readme:   "This is a useful tool.",
			expected: true,
		},
		{
			name:     "Too sparse",
			files:    []string{"README.md", "LICENSE"},
			readme:   "Hello world",
			expected: false,
		},
		{
			name:     "Documentation only",
			files:    []string{"README.md", "guide.pdf", "image.png", "LICENSE"},
			readme:   "Just docs",
			expected: false,
		},
		{
			name:     "Curated list",
			files:    []string{"README.md", "links.txt", "contributing.md", "other.md"},
			readme:   "An awesome-list of things",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeDiscretion(tt.files, tt.readme)
			assert.Equal(t, tt.expected, result.Allowed, result.Reason)
		})
	}
}
