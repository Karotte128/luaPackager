package bundler

import (
	"fmt"
	"sort"
)

type sortedModule struct {
	Name   string
	Source string
}

type visitState int

const (
	unvisited visitState = iota
	visiting
	visited
)

func sortModules(mods map[string]Module) ([]sortedModule, error) {
	state := make(map[string]visitState, len(mods))
	result := make([]sortedModule, 0, len(mods))

	// deterministic module order
	keys := make([]string, 0, len(mods))
	for k := range mods {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var visit func(string) error

	visit = func(name string) error {
		switch state[name] {
		case visited:
			return nil
		case visiting:
			// cycle detected → ignore, do not error
			return nil
		}

		mod, ok := mods[name]
		if !ok {
			return fmt.Errorf("unknown module %q", name)
		}

		state[name] = visiting

		// optional: sort dependencies for deterministic traversal
		deps := append([]string(nil), mod.Requires...)
		sort.Strings(deps)

		for _, dep := range deps {
			if _, ok := mods[dep]; !ok {
				return fmt.Errorf("module %q depends on missing %q", name, dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}

		state[name] = visited
		result = append(result, sortedModule{Source: mod.Source, Name: name})

		return nil
	}

	for _, name := range keys {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}
