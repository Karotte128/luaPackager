package parser

import (
	"github.com/karotte128/luapackager/bundler"
)

type LookupFunc func(module string) (string, error)

func Parse(entry string, lookup LookupFunc) (bundler.BundlerInput, error) {
	p := &parser{
		lookup:  lookup,
		visited: make(map[string]bool),
		modules: make(map[string]bundler.Module),
	}

	if err := p.parseModule(entry); err != nil { // start the module parsing on the entrypoint module
		return bundler.BundlerInput{}, err
	}

	return bundler.BundlerInput{
		Entry:   entry,
		Modules: p.modules,
	}, nil
}
