package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type apiSpec struct {
	Symbols []string `json:"symbols"`
}

func main() {
	spec, err := loadSpec("public_api.json")
	if err != nil {
		panic(err)
	}
	sort.Strings(spec.Symbols)

	content, err := renderMarkdown(spec)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("PUBLIC_API.md", []byte(content), 0644); err != nil {
		panic(err)
	}
}

func loadSpec(path string) (*apiSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var spec apiSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}
	return &spec, nil
}

func renderMarkdown(spec *apiSpec) (string, error) {
	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal spec: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("# Public API Surface\n\n")
	builder.WriteString("Generated from `public_api.json` via `go run ./internal/tools/public_api_export.go`.\n\n")
	builder.WriteString("## Machine-readable Source\n\n")
	builder.WriteString("```json\n")
	builder.Write(jsonData)
	builder.WriteString("\n```\n\n")
	builder.WriteString("## Symbols\n\n")
	for _, symbol := range spec.Symbols {
		builder.WriteString("- `")
		builder.WriteString(symbol)
		builder.WriteString("`\n")
	}
	return builder.String(), nil
}
