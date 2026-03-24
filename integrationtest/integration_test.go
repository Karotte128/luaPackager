package integrationtest

import (
	"errors"
	"strings"
	"testing"

	"github.com/karotte128/luapackager/bundler"
	"github.com/karotte128/luapackager/parser"
)

func mustNotContain(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Fatalf("did not expect output to contain %q\n\n%s", substr, output)
	}
}

func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Fatalf("expected output to contain %q\n\n%s", substr, output)
	}
}

func TestIntegration_SimpleModule(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `print("hello")`, nil
		}
		return "", errors.New("not found")
	}

	// Use only exported API
	bundleInput, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(bundleInput)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, `print("hello")`) {
		t.Fatal("output missing module content")
	}
	if !strings.Contains(output, `require("main")`) {
		t.Fatal("output missing entry require")
	}
}

func TestIntegration_TransitiveDependencies(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("a")`,
			"a":    `require("b")`,
			"b":    `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	bundleInput, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(bundleInput)
	if err != nil {
		t.Fatal(err)
	}

	// Ensure ordering: b before a before main
	iB := strings.Index(output, `package.preload["b"]`)
	iA := strings.Index(output, `package.preload["a"]`)
	iMain := strings.Index(output, `package.preload["main"]`)

	if !(iB < iA && iA < iMain) {
		t.Fatalf("incorrect module order: b=%d a=%d main=%d", iB, iA, iMain)
	}
}

func TestIntegration_DiamondDependencies(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("b"); require("c")`,
			"b":    `require("d")`,
			"c":    `require("d")`,
			"d":    `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	bundleInput, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(bundleInput)
	if err != nil {
		t.Fatal(err)
	}

	// All modules should appear
	for _, mod := range []string{"main", "b", "c", "d"} {
		if !strings.Contains(output, `package.preload["`+mod+`"]`) {
			t.Fatalf("module %q missing in bundle", mod)
		}
	}
}

func TestIntegration_Cycle(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"a": `require("b")`,
			"b": `require("a")`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	bundleInput, err := parser.Parse("a", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(bundleInput)
	if err != nil {
		t.Fatal(err)
	}

	for _, mod := range []string{"a", "b"} {
		if !strings.Contains(output, `package.preload["`+mod+`"]`) {
			t.Fatalf("module %q missing in bundle", mod)
		}
	}
}

func TestIntegration_DynamicRequireError(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require(x)`, nil
		}
		return "", errors.New("not found")
	}

	_, err := parser.Parse("main", lookup)
	if err == nil {
		t.Fatal("expected parser to fail on dynamic require")
	}
}

func TestIntegration_MissingModuleError(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require("missing")`, nil
		}
		return "", errors.New("not found")
	}

	_, err := parser.Parse("main", lookup)
	if err == nil {
		t.Fatal("expected parser to fail on missing module")
	}
}

func TestIntegration_ModuleLevelExternal(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `--!external socket
require("socket")
require("a")`,
			"a": `return { value = 42 }`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustContain(t, output, `package.preload["a"]`)
	mustNotContain(t, output, `package.preload["socket"]`)
}

func TestIntegration_InlineExternal(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("socket") --!external
require("a")`,
			"a": `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustNotContain(t, output, `package.preload["socket"]`)
	mustContain(t, output, `package.preload["a"]`)
}

func TestIntegration_MixedExternalForms(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `--!external json
require("socket") --!external
require("json")
require("a")`,
			"a": `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	mustNotContain(t, output, `package.preload["socket"]`)
	mustNotContain(t, output, `package.preload["json"]`)
	mustContain(t, output, `package.preload["a"]`)
}

func TestIntegration_ExternalNotIncludedButStillRequired(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `--!external socket
local s = require("socket")`, nil
		}
		return "", errors.New("not found")
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	// ensure require is still present in source
	mustContain(t, output, `require("socket")`)
	mustNotContain(t, output, `package.preload["socket"]`)
}

func TestIntegration_PerModuleExternalIsolation(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("a")
require("b")`,
			"a": `--!external socket
require("socket")`,
			"b":      `require("socket")`,
			"socket": `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	// socket should be included because b requires it normally
	mustContain(t, output, `package.preload["socket"]`)
}

func TestIntegration_ExternalChain(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `--!external a
require("a")`,
			"a": `require("b")`,
			"b": `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	input, err := parser.Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	output, err := bundler.Bundle(input)
	if err != nil {
		t.Fatal(err)
	}

	// since "a" is external, neither a nor b should be bundled
	mustNotContain(t, output, `package.preload["a"]`)
	mustNotContain(t, output, `package.preload["b"]`)
}

func TestIntegration_DynamicRequireStillFails(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require(x) --!external`, nil
		}
		return "", errors.New("not found")
	}

	_, err := parser.Parse("main", lookup)
	if err == nil {
		t.Fatal("expected error for dynamic require even with external annotation")
	}
}
