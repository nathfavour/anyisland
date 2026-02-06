package agent

import (
	"strings"
)

func AnalyzeDiscretion(files []string, readme string) DiscretionResult {
	// 1. Check for "Empty" or "Skeleton" repositories
	if len(files) < 3 {
		return DiscretionResult{
			Allowed: false,
			Reason:  "Repository is too sparse to be a functional tool.",
		}
	}

	// 2. Identify "Documentation Only" repositories
	isDocOnly := true
	docExtensions := []string{".md", ".txt", ".pdf", ".html", ".png", ".jpg"}
	
	for _, f := range files {
		ext := ""
		if idx := strings.LastIndex(f, "."); idx != -1 {
			ext = f[idx:]
		}
		
		isDoc := false
		for _, docExt := range docExtensions {
			if strings.EqualFold(ext, docExt) {
				isDoc = true
				break
			}
		}
		
		if !isDoc && !strings.HasPrefix(f, ".") && !strings.Contains(f, "LICENSE") {
			isDocOnly = false
			break
		}
	}

	if isDocOnly {
		return DiscretionResult{
			Allowed: false,
			Reason:  "Repository appears to be documentation or assets only.",
		}
	}

	// 3. Check for "Vague" keyword signals in README
	vagueKeywords := []string{"awesome-list", "collection of", "curated list", "bookmarks"}
	lowerReadme := strings.ToLower(readme)
	for _, kw := range vagueKeywords {
		if strings.Contains(lowerReadme, kw) {
			return DiscretionResult{
				Allowed: false,
				Reason:  "Repository appears to be a curated list or collection rather than a tool.",
			}
		}
	}

	return DiscretionResult{Allowed: true}
}
