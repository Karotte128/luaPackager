# luapackager

A simple, modular Lua bundler written in Go. It statically resolves `require()` calls and produces a single Lua file containing all internal dependencies, while allowing explicit exclusion of external modules.

---

## Features

* Recursive dependency resolution
* Pure Lua output (no minification or obfuscation)
* Pluggable module lookup (filesystem, HTTP, etc.)
* Support for external modules via annotations:

  * `--!external module`
  * `require("module") --!external`
* Deterministic bundling (stable ordering)

---

## Installation

```bash
git clone https://github.com/karotte128/luapackager.git
cd luapackager
go build -o luapackager
```

---

## CLI Usage

```bash
./luapackager -entry <module> -dir <path> -out <file>
```

### Flags

| Flag     | Description                        | Default      |
| -------- | ---------------------------------- | ------------ |
| `-entry` | Entry module name (required)       | —            |
| `-dir`   | Base directory for Lua modules     | `.`          |
| `-out`   | Output file for bundled Lua script | `bundle.lua` |

---

### Example

```bash
./luapackager -entry main -dir ./lua_modules -out bundle.lua
```

This will:

* Load `main.lua` from `./lua_modules`
* Resolve all dependencies
* Generate a single bundled file `bundle.lua`

---

## External Modules

You can mark modules as external so they are **not bundled** but still required at runtime.

### Module-level

```lua
--!external socket

local socket = require("socket")
```

### Inline

```lua
local socket = require("socket") --!external
```

---

## Go API Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/karotte128/luapackager/bundler"
    "github.com/karotte128/luapackager/parser"
)

func main() {
    // create a lookup function, searching in the current directory
    lookup := fileLookup("./")

    // parse the lua files for dependencies
    input, err := parser.Parse("main", lookup)
    if err != nil {
        log.Fatal(err)
    }

    // bundle all dependencies into a single file
    output, err := bundler.Bundle(input)
    if err != nil {
        log.Fatal(err)
    }

    // printing the bundled output
    fmt.Println(output)
}

// fileLookup returns a LookupFunc that reads modules from a base directory.
func fileLookup(baseDir string) parser.LookupFunc {
	return func(module string) (string, error) {
		// Convert module name to path: foo.bar -> foo/bar.lua
		path := filepath.Join(baseDir, strings.ReplaceAll(module, ".", string(filepath.Separator))+".lua")

		// Read file contents
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("module %q not found at %q: %w", module, path, err)
		}

		return string(data), nil
	}
}
```

---

## Custom Lookup

You can provide your own module resolution logic:

```go
lookup := func(name string) (string, error) {
    // Load from filesystem, HTTP, memory, etc.
}
```

This makes the parser fully extensible.

---

## Project Structure

* `parser` — resolves dependencies and builds module graph
* `bundler` — generates final Lua output
* CLI — simple wrapper around parser + bundler

---

## Limitations

* Only supports **static `require("...")`**
* Dynamic requires (`require(x)`) will cause an error
* Regex-based parsing (may not handle all Lua edge cases)

---

## Future Improvements

* Lua tokenizer / AST parsing
* Multiple lookup paths (like `package.path`)
* Debug/trace output
* Plugin system for transformations

---

## License

MIT
