package parser

import (
	"fmt"

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

	requires, err := extractRequires(src)
	if err != nil {
		return fmt.Errorf("module %q: %w", name, err)
	}

	p.modules[name] = bundler.Module{
		Source:   src,
		Requires: requires,
	}

	for _, dep := range requires {
		if err := p.parseModule(dep); err != nil {
			return err
		}
	}

	return nil
}
