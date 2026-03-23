package parser

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

func mustEqual(t *testing.T, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParse_SimpleModule(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `print("hello")`, nil
		}
		return "", errors.New("not found")
	}

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := bundle.Modules["main"]; !ok {
		t.Fatal("main module not present")
	}

	if bundle.Entry != "main" {
		t.Fatal("entry module mismatch")
	}
}

func TestParse_SingleDependency(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("foo")`,
			"foo":  `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"main", "foo"}
	got := make([]string, 0, len(bundle.Modules))
	for k := range bundle.Modules {
		got = append(got, k)
	}

	mustEqual(t, got, expected)
}

func TestParse_TransitiveDependencies(t *testing.T) {
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

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"main", "a", "b"}
	got := make([]string, 0, len(bundle.Modules))
	for k := range bundle.Modules {
		got = append(got, k)
	}

	mustEqual(t, got, expected)
}

func TestParse_DiamondDependencies(t *testing.T) {
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

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	for _, mod := range []string{"main", "b", "c", "d"} {
		if _, ok := bundle.Modules[mod]; !ok {
			t.Fatalf("module %q missing", mod)
		}
	}
}

func TestParse_Cycle(t *testing.T) {
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

	bundle, err := Parse("a", lookup)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := bundle.Modules["a"]; !ok {
		t.Fatal("module a missing")
	}
	if _, ok := bundle.Modules["b"]; !ok {
		t.Fatal("module b missing")
	}
}

func TestParse_SelfCycle(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "a" {
			return `require("a")`, nil
		}
		return "", errors.New("not found")
	}

	bundle, err := Parse("a", lookup)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := bundle.Modules["a"]; !ok {
		t.Fatal("module a missing")
	}
}

func TestParse_MultipleRequiresDedup(t *testing.T) {
	lookup := func(name string) (string, error) {
		files := map[string]string{
			"main": `require("a"); require("b"); require("a")`,
			"a":    `return {}`,
			"b":    `return {}`,
		}
		src, ok := files[name]
		if !ok {
			return "", errors.New("not found")
		}
		return src, nil
	}

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"main", "a", "b"}
	got := make([]string, 0, len(bundle.Modules))
	for k := range bundle.Modules {
		got = append(got, k)
	}

	mustEqual(t, got, expected)
}

func TestParse_Normalization(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require("foo/bar")`, nil
		}
		if name == "foo.bar" {
			return `return {}`, nil
		}
		return "", errors.New("not found")
	}

	bundle, err := Parse("main", lookup)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := bundle.Modules["foo.bar"]; !ok {
		t.Fatal("normalized module foo.bar missing")
	}
}

func TestParse_DynamicRequireError(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require(x)`, nil
		}
		return "", errors.New("not found")
	}

	_, err := Parse("main", lookup)
	if err == nil {
		t.Fatal("expected error for dynamic require")
	}
}

func TestParse_ConcatenatedRequireError(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require("a" .. "b")`, nil
		}
		return "", errors.New("not found")
	}

	_, err := Parse("main", lookup)
	if err == nil {
		t.Fatal("expected error for concatenated require")
	}
}

func TestParse_MissingModuleError(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "main" {
			return `require("missing")`, nil
		}
		return "", errors.New("not found")
	}

	_, err := Parse("main", lookup)
	if err == nil {
		t.Fatal("expected error for missing module")
	}
}
