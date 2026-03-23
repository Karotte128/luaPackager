package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/karotte128/luapackager/bundler"
	"github.com/karotte128/luapackager/parser"
)

func main() {
	// Command-line flags
	entry := flag.String("entry", "", "entrypoint Lua module (e.g., main)")
	dir := flag.String("dir", ".", "base directory for Lua modules")
	output := flag.String("out", "bundle.lua", "output file name for the bundled Lua script")
	flag.Parse()

	if *entry == "" {
		fmt.Fprintln(os.Stderr, "Error: entry module must be specified with -entry")
		os.Exit(1)
	}

	// File-based LookupFunc
	lookup := fileLookup(*dir)

	// Parse the module graph
	bundleInput, err := parser.Parse(*entry, lookup)
	if err != nil {
		log.Fatal("Parser error:", err)
	}

	// Generate bundled Lua output
	luaCode, err := bundler.Bundle(bundleInput)
	if err != nil {
		log.Fatal("Bundler error:", err)
	}

	// Save to output file
	if err := os.WriteFile(*output, []byte(luaCode), 0644); err != nil {
		log.Fatal("Failed to write output file:", err)
	}

	fmt.Printf("Bundled Lua script saved to %s\n", *output)
}

// fileLookup returns a LookupFunc that reads modules from a base directory.
func fileLookup(baseDir string) parser.LookupFunc {
	return func(module string) (string, error) {
		// Convert module name to path: foo.bar -> foo/bar.lua
		path := filepath.Join(baseDir, strings.ReplaceAll(module, ".", string(filepath.Separator))+".lua")

		// Read file contents
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("module %q not found at %q: %w", module, path, err)
		}

		return string(data), nil
	}
}
