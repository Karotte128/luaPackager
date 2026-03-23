package bundler

import "fmt"

func validate(input BundlerInput) error {
	// 1. Entry must exist
	_, ok := input.Modules[input.Entry]
	if !ok {
		return fmt.Errorf("entry module %q not found", input.Entry)
	}

	// 2. All requires must exist
	for name, m := range input.Modules {
		for _, dep := range m.Requires {
			if _, ok := input.Modules[dep]; !ok {
				return fmt.Errorf("module %q requires missing module %q", name, dep)
			}
		}
	}

	return nil
}
