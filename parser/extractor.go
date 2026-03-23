package parser

import (
	"fmt"
	"regexp"
	"sort"
)

var requireStrictRe = regexp.MustCompile(
	`require\s*\(\s*["']([^"']+)["']\s*\)`,
)

var requireCallRe = regexp.MustCompile(
	`require\s*\(`,
)

func extractRequires(src string) ([]string, error) {
	// 1. extract valid requires
	matches := requireStrictRe.FindAllStringSubmatch(src, -1)

	found := make(map[string]struct{})
	for _, m := range matches {
		name := normalizeModuleName(m[1])
		found[name] = struct{}{}
	}

	// 2. detect dynamic requires
	calls := requireCallRe.FindAllStringIndex(src, -1)
	if len(calls) != len(matches) {
		return nil, fmt.Errorf("dynamic require detected")
	}

	// 3. stable ordering
	result := make([]string, 0, len(found))
	for k := range found {
		result = append(result, k)
	}
	sort.Strings(result)

	return result, nil
}
