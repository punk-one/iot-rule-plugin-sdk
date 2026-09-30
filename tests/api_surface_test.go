package sdk_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestPublicAPISurface(t *testing.T) {
	want := loadPublicAPISpec(t)
	got := collectExportedDecls(t)

	if !equalSet(got, want) {
		t.Fatalf("public API mismatch\nwant: %s\n got: %s", setString(want), setString(got))
	}
}

func collectExportedDecls(t *testing.T) map[string]struct{} {
	t.Helper()

	matches, err := filepath.Glob("../*.go")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}

	fset := token.NewFileSet()
	got := make(map[string]struct{})

	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s failed: %v", path, err)
		}

		for _, decl := range file.Decls {
			switch typed := decl.(type) {
			case *ast.FuncDecl:
				if typed.Recv == nil && typed.Name.IsExported() {
					got[typed.Name.Name] = struct{}{}
				}
			case *ast.GenDecl:
				for _, spec := range typed.Specs {
					switch defined := spec.(type) {
					case *ast.TypeSpec:
						if defined.Name.IsExported() {
							got[defined.Name.Name] = struct{}{}
						}
					case *ast.ValueSpec:
						for _, name := range defined.Names {
							if name.IsExported() {
								got[name.Name] = struct{}{}
							}
						}
					}
				}
			}
		}
	}

	return got
}

func loadPublicAPISpec(t *testing.T) map[string]struct{} {
	t.Helper()

	data, err := os.ReadFile("../public_api.json")
	if err != nil {
		t.Fatalf("read public_api.json failed: %v", err)
	}

	var spec struct {
		Symbols []string `json:"symbols"`
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatalf("unmarshal public_api.json failed: %v", err)
	}
	if len(spec.Symbols) == 0 {
		t.Fatal("public_api.json symbols is empty")
	}

	want := make(map[string]struct{}, len(spec.Symbols))
	for _, symbol := range spec.Symbols {
		if symbol == "" {
			t.Fatal("public_api.json contains empty symbol")
		}
		want[symbol] = struct{}{}
	}
	return want
}

func equalSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if _, ok := b[key]; !ok {
			return false
		}
	}
	return true
}

func setString(s map[string]struct{}) string {
	keys := make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}
