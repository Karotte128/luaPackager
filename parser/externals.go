package parser

import (
	"regexp"
	"slices"
)

var moduleRe = regexp.MustCompile(`--!\s*external\s+([^\s]+)`)
var inlineRe = regexp.MustCompile(`require\s*\(\s*["']([^"']+)["']\s*\)\s*--!\s*external`)

func extractModuleExternals(src string) []string {
	var result []string

	matches := moduleRe.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		name := normalizeModuleName(m[1])
		result = append(result, name)
	}

	return result
}

func extractInlineExternals(src string) []string {
	var result []string

	matches := inlineRe.FindAllStringSubmatch(src, -1)
	for _, m := range matches {
		name := normalizeModuleName(m[1])
		result = append(result, name)
	}

	return result
}

func extractExternals(source string) []string {
	return slices.Concat(extractModuleExternals(source), extractInlineExternals(source))
}
