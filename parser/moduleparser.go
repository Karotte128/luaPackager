package parser

import (
	"fmt"
	"slices"

	"github.com/karotte128/luapackager/bundler"
)

type parser struct {
	lookup  LookupFunc
	visited map[string]bool
	modules map[string]bundler.Module
}

func (p *parser) parseModule(name string) error {
	if p.visited[name] {
		return nil
	}
	p.visited[name] = true

	src, err := p.lookup(name)
	if err != nil {
		return fmt.Errorf("module %q: lookup failed: %w", name, err)
	}

	externals := extractExternals(src)

	requires, extractErr := extractRequires(src)
	if extractErr != nil {
		return fmt.Errorf("module %q: %w", name, extractErr)
	}

	internalDeps := subtractDeps(requires, externals)

	p.modules[name] = bundler.Module{
		Source:   src,
		Requires: internalDeps,
	}

	for _, dep := range internalDeps {
		if err := p.parseModule(dep); err != nil {
			return err
		}
	}

	return nil
}

func subtractDeps(all []string, externals []string) []string {
	result := make([]string, 0, len(all))
	for _, dep := range all {
		if !slices.Contains(externals, dep) {
			result = append(result, dep)
		}
	}
	return result
}
