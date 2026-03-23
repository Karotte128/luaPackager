package bundler

import "fmt"

type Module struct {
	Source   string
	Requires []string
}

type BundlerInput struct {
	Entry   string
	Modules map[string]Module
}

func Bundle(input BundlerInput) (string, error) {
	validateErr := validate(input)
	if validateErr != nil {
		return "", validateErr
	}

	sorted, sortErr := sortModules(input.Modules)
	if sortErr != nil {
		return "", sortErr
	}

	bundle := buildBundle(sorted)

	bundle += fmt.Sprintf(`
-- ENTRYPOINT
require("%s")`, input.Entry)

	return bundle, nil
}
