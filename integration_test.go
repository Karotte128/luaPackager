package luapackager

import (
	"errors"
	"strings"
	"testing"

	"github.com/karotte128/luapackager/bundler"
	"github.com/karotte128/luapackager/parser"
)

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
