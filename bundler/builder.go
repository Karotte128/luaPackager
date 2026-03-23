package bundler

import "fmt"

func buildBundle(modules map[string]Module) string {
	var wrappedSource string

	for name, module := range modules {
		wrappedSource += wrapModule(name, module.Source)
	}

	return wrappedSource
}

func wrapModule(name string, content string) string {
	return fmt.Sprintf(`-- BEGIN module: %s
package.preload[%q] = function()

%s

end
-- END module: %s

`, name, name, content, name)
}
