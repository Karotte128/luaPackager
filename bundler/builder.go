package bundler

import "fmt"

func buildBundle(modules []sortedModule) string {
	var wrappedSource string

	for _, module := range modules {
		wrappedSource += wrapModule(module.Name, module.Source)
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
