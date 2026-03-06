package main

// Package main implements the go-lib scaffold CLI.
// It generates a production-ready service skeleton wired to go-lib conventions.
//
// Usage:
//
// go run ./scaffold --name orders --module github.com/acme/orders [--out .] [--dry-run]/

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

const (
	defaultGoLibModule = "github.com/nojyerac/go-lib"
	defaultGoVersion   = "1.25.1"
)

// Options configure what the generator produces.
type Options struct {
	// Name is the short service name, e.g. "orders".
	Name string
	// Module is the fully-qualified Go module path, e.g. "github.com/acme/orders".
	Module string
	// GoLibModule is the import path for go-lib (defaults to defaultGoLibModule).
	GoLibModule string
	// GoVersion is the minimum Go version written into go.mod (defaults to defaultGoVersion).
	GoVersion string
}

func stringsToTitle(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// NameTitle returns Name with a capitalised first letter (for use in generated
// Go identifiers and display strings).
func (o Options) NameTitle() string {
	return stringsToTitle(o.Name)
}

// NameUpper returns Name in UPPER_CASE (used as the env-var prefix for viper).
func (o Options) NameUpper() string {
	return strings.ToUpper(o.Name)
}

// NameLower returns Name in lower_case (used for some file paths).
func (o Options) NameLower() string {
	return strings.ToLower(o.Name)
}

// fileEntries maps an output path template string to the name of the embedded
// template that produces the file's content. The path strings are themselves
// rendered as text/template so they can reference {{.Name}} etc.
var fileEntries = []struct {
	pathTmpl string
	tmplName string
}{
	{".github/workflows/ci.yml", "github_workflows_ci.yml.tmpl"},
	{"api/openapi.yml", "api_openapi.yml.tmpl"},
	{"api/example.proto", "api_example.proto.tmpl"},
	{"cmd/{{.Name}}/main.go", "cmd_main.go.tmpl"},
	{"config/config.go", "config_config.go.tmpl"},
	{"data/data.go", "data_data.go.tmpl"},
	{"data/data_suite_test.go", "data_data_suite_test.go.tmpl"},
	{"data/db/db.go", "data_db_db.go.tmpl"},
	{"data/db/db_suite_test.go", "data_db_db_suite_test.go.tmpl"},
	{"scripts/generate.sh", "scripts_generate.sh.tmpl"},
	{"scripts/lint.sh", "scripts_lint.sh.tmpl"},
	{"scripts/test.sh", "scripts_test.sh.tmpl"},
	{"transport/http/http.go", "transport_http_http.go.tmpl"},
	{"transport/http/http_suite_test.go", "transport_http_http_suite_test.go.tmpl"},
	{"transport/rpc/rpc.go", "transport_rpc_rpc.go.tmpl"},
	{"transport/rpc/rpc_suite_test.go", "transport_rpc_rpc_suite_test.go.tmpl"},
	{".golangci.yml", "golangci.yml.tmpl"},
	{".mockery.yml", "mockery.yml.tmpl"},
	{"go.mod", "go.mod.tmpl"},
	{"Dockerfile", "Dockerfile.tmpl"},
	{"Makefile", "Makefile.tmpl"},
	{"README.md", "README.md.tmpl"},
}

// Generator renders service skeletons from the embedded template set.
type Generator struct {
	tmpl *template.Template
}

// NewGenerator parses all embedded templates and returns a ready Generator.
func NewGenerator() (*Generator, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
		"Title":   stringsToTitle,
	}
	tmpl, err := template.New("root").
		Funcs(funcMap).
		ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	return &Generator{tmpl: tmpl}, nil
}

// Generate executes every template and returns a map of
// (relative output path) → (file content string).
func (g *Generator) Generate(opts Options) (map[string]string, error) {
	if opts.GoLibModule == "" {
		opts.GoLibModule = defaultGoLibModule
	}
	if opts.GoVersion == "" {
		opts.GoVersion = defaultGoVersion
	}

	out := make(map[string]string, len(fileEntries))
	for _, entry := range fileEntries {
		path, err := renderString(entry.pathTmpl, opts)
		if err != nil {
			return nil, fmt.Errorf("render path %q: %w", entry.pathTmpl, err)
		}

		var buf bytes.Buffer
		if err := g.tmpl.ExecuteTemplate(&buf, entry.tmplName, opts); err != nil {
			return nil, fmt.Errorf("execute template %q: %w", entry.tmplName, err)
		}
		out[path] = buf.String()
	}
	return out, nil
}

// renderString treats s as a text/template, executes it with data, and returns
// the result. Used to expand {{.Name}} in output-path patterns.
func renderString(s string, data any) (string, error) {
	t, err := template.New("path").Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
