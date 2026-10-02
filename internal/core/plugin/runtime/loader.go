package runtime

import (
	"fmt"
	"reflect"
	"strings"

	"repo-nest/internal/core/plugin"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// exportedTypes is the set of host symbols exposed to plugin scripts via the
// import path "repo-nest/internal/core/plugin". Scripts use plugin.Context and
// plugin.ImportDoc.
var exportedTypes = interp.Exports{
	"repo-nest/internal/core/plugin/plugin": {
		"Context":   reflect.ValueOf((*Context)(nil)),
		"Event":     reflect.ValueOf(plugin.Event{}),
		"ImportDoc": reflect.ValueOf(plugin.ImportDoc{}),
	},
}

// safeSymbolPkgs is the allowlist of stdlib packages a plugin may import.
//
// A plugin already holds full host authority by design: it receives
// plugin.Context with direct database access, and loading it is an explicit
// act of trust in its source. But the interpreter must not ALSO hand out the
// packages that reach past RepoNest entirely — os/os/exec/syscall/unsafe
// escape to the filesystem and process table, net*/net/http reach the
// network — because then a single dropped-in plugin.go becomes arbitrary
// code execution in the app process with no product-level gate.
// Extend deliberately, never wholesale.
var safeSymbolPkgs = map[string]bool{
	"bufio":            true,
	"bytes":            true,
	"cmp":              true,
	"context":          true,
	"encoding/base64":  true,
	"encoding/csv":     true,
	"encoding/hex":     true,
	"encoding/json":    true,
	"encoding/utf8":    true,
	"errors":           true,
	"fmt":              true,
	"hash":             true,
	"hash/crc32":       true,
	"hash/crc64":       true,
	"hash/fnv":         true,
	"io":               true,
	"maps":             true,
	"math":             true,
	"math/bits":        true,
	"math/rand":        true,
	"net/url":          true,
	"path":             true,
	"path/filepath":    true,
	"regexp":           true,
	"slices":           true,
	"sort":             true,
	"strconv":          true,
	"strings":          true,
	"sync":             true,
	"sync/atomic":      true,
	"text/tabwriter":   true,
	"text/template":    true,
	"time":             true,
	"unicode":          true,
	"unicode/utf16":    true,
	"unicode/utf8":     true,
}

// safeSymbols filters stdlib.Symbols down to safeSymbolPkgs. yaegi keys its
// stdlib entries as "<import path>/<package name>" ("fmt/fmt", "os/exec/exec"),
// so the trailing element is stripped to get the path a plugin would import.
// The package drop is silent — a plugin importing an excluded package fails
// at Eval time with an "unable to find source" error, which loadPlugin
// records as a normal per-plugin load failure.
func safeSymbols() map[string]map[string]reflect.Value {
	out := make(map[string]map[string]reflect.Value, len(safeSymbolPkgs))
	for key, symbols := range stdlib.Symbols {
		i := strings.LastIndexByte(key, '/')
		if i < 0 || !safeSymbolPkgs[key[:i]] {
			// No slash: the "." root entry and similar non-package keys.
			continue
		}
		out[key] = symbols
	}
	return out
}

// script wraps a yaegi interpreter bound to one plugin source file.
type script struct {
	i *interp.Interpreter
}

// compileScript evaluates a plugin source file and prepares it for symbol
// lookup. Compilation errors are returned verbatim.
func compileScript(src string) (*script, error) {
	i := interp.New(interp.Options{})
	i.Use(safeSymbols())
	i.Use(exportedTypes)

	if _, err := i.Eval(src); err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}
	return &script{i: i}, nil
}

// funcValue looks up a package-level symbol such as "main.Name".
func (s *script) funcValue(sym string) (reflect.Value, error) {
	v, err := s.i.Eval(sym)
	if err != nil {
		return reflect.Value{}, err
	}
	if !v.IsValid() || v.Kind() != reflect.Func {
		return reflect.Value{}, fmt.Errorf("%s is not a function", sym)
	}
	return v, nil
}
