package bundler

import (
	"strings"
	"testing"
)

func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Fatalf("expected output to contain %q\n\nGot:\n%s", substr, output)
	}
}

func mustOrder(t *testing.T, output string, before string, after string) {
	t.Helper()

	i := strings.Index(output, before)
	j := strings.Index(output, after)

	if i == -1 || j == -1 {
		t.Fatalf("could not find substrings in output\nbefore=%q\nafter=%q\n\n%s", before, after, output)
	}

	if i > j {
		t.Fatalf("expected %q to appear before %q\n\n%s", before, after, output)
	}
}

func TestBundle_Simple(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source: `print("hello")`,
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustContain(t, out, `package.preload["main"]`)
	mustContain(t, out, `print("hello")`)
	mustContain(t, out, `require("main")`)
}

func TestBundle_DependencyOrder(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source:   `require("foo")`,
				Requires: []string{"foo"},
			},
			"foo": {
				Source: `return {}`,
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustOrder(t, out,
		`package.preload["foo"]`,
		`package.preload["main"]`,
	)
}

func TestBundle_TransitiveDependencies(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source:   `require("a")`,
				Requires: []string{"a"},
			},
			"a": {
				Source:   `require("b")`,
				Requires: []string{"b"},
			},
			"b": {
				Source: `return {}`,
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustOrder(t, out, `package.preload["b"]`, `package.preload["a"]`)
	mustOrder(t, out, `package.preload["a"]`, `package.preload["main"]`)
}

func TestBundle_DiamondDependency(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source:   `require("b"); require("c")`,
				Requires: []string{"b", "c"},
			},
			"b": {
				Source:   `require("d")`,
				Requires: []string{"d"},
			},
			"c": {
				Source:   `require("d")`,
				Requires: []string{"d"},
			},
			"d": {
				Source: `return {}`,
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustOrder(t, out, `package.preload["d"]`, `package.preload["b"]`)
	mustOrder(t, out, `package.preload["d"]`, `package.preload["c"]`)
}

func TestBundle_Cycle(t *testing.T) {
	input := BundlerInput{
		Entry: "a",
		Modules: map[string]Module{
			"a": {
				Source:   `require("b")`,
				Requires: []string{"b"},
			},
			"b": {
				Source:   `require("a")`,
				Requires: []string{"a"},
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustContain(t, out, `package.preload["a"]`)
	mustContain(t, out, `package.preload["b"]`)
}

func TestBundle_SelfCycle(t *testing.T) {
	input := BundlerInput{
		Entry: "a",
		Modules: map[string]Module{
			"a": {
				Source:   `require("a")`,
				Requires: []string{"a"},
			},
		},
	}

	_, err := Bundle(input)
	if err != nil {
		t.Fatal("self-cycle should not error")
	}
}

func TestBundle_MissingDependency(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Requires: []string{"missing"},
			},
		},
	}

	_, err := Bundle(input)
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestBundle_MissingEntry(t *testing.T) {
	input := BundlerInput{
		Entry:   "main",
		Modules: map[string]Module{},
	}

	_, err := Bundle(input)
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

func TestBundle_DeterministicOutput(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"b": {Source: `return {}`},
			"a": {Source: `return {}`},
			"main": {
				Requires: []string{"a", "b"},
			},
		},
	}

	out1, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	out2, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	if out1 != out2 {
		t.Fatal("output is not deterministic")
	}
}

func TestBundle_SourcePreserved(t *testing.T) {
	src := `
local M = {}

function M.test()
    local x = 42
    return x
		t.Fatal("ou
end

return M
`

	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source: src,
			},
		},
	}

	out, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustContain(t, out, src)
}

func TestBundle_EmptyModule(t *testing.T) {
	input := BundlerInput{
		Entry: "main",
		Modules: map[string]Module{
			"main": {
				Source: ``,
			},
		},
	}

	_, err := Bundle(input)
	if err != nil {
		t.Fatal(err)
	}
}
