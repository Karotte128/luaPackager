package parser

import "strings"

func normalizeModuleName(name string) string {
	name = strings.ReplaceAll(name, "/", ".")
	return strings.TrimSpace(name)
}
