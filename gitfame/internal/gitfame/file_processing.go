package gitfame

import (
	"path/filepath"
	"strings"
)

func shouldProcessFile(
	filePath string,
	allowedExtensions map[string]bool,
	allowedLanguageExtensions map[string]bool,
	useLanguageFilter bool,
	restrictPatterns []string,
	excludePatterns []string,
) bool {
	if len(restrictPatterns) > 0 {
		if !matchesAnyPattern(filePath, restrictPatterns) {
			return false
		}
	}

	if len(excludePatterns) > 0 {
		if matchesAnyPattern(filePath, excludePatterns) {
			return false
		}
	}

	extension := strings.ToLower(filepath.Ext(filePath))

	if len(allowedExtensions) > 0 {
		if extension == "" || !allowedExtensions[extension] {
			return false
		}
	}

	if useLanguageFilter {
		if extension == "" || !allowedLanguageExtensions[extension] {
			return false
		}
	}

	return true
}

func matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		matched, _ := filepath.Match(pattern, filePath)
		if matched {
			return true
		}
	}

	return false
}
