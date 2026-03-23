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

	bundle := buildBundle(input.Modules)

	bundle += fmt.Sprintf(`

	-- ENTRYPOINT
	require("%s")`, input.Entry)

	return bundle, nil
}
