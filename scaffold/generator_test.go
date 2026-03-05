package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenOpts are the fixed inputs used to produce every golden snapshot.
var goldenOpts = Options{
	Name:        "example",
	Module:      "github.com/example/example",
	GoLibModule: defaultGoLibModule,
	GoVersion:   defaultGoVersion,
}

// TestGenerate_Golden asserts that Generate produces output identical to the
// checked-in snapshots under testdata/golden/.
//
// To regenerate snapshots after intentional template changes, run:
//
//	UPDATE_GOLDEN=1 go test ./scaffold/... -run TestGenerate_Golden
func TestGenerate_Golden(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	files, err := gen.Generate(goldenOpts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	update := os.Getenv("UPDATE_GOLDEN") == "1"

	for relPath, got := range files {
		// Golden file path mirrors the relative output path, with path
		// separators replaced so we keep a flat testdata directory structure.
		goldenName := strings.ReplaceAll(relPath, string(filepath.Separator), "__")
		goldenName = strings.ReplaceAll(goldenName, "/", "__")
		goldenPath := filepath.Join("testdata", "golden", goldenName)

		if update {
			if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", filepath.Dir(goldenPath), err)
			}
			if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
				t.Fatalf("write golden %s: %v", goldenPath, err)
			}
			t.Logf("updated golden: %s", goldenPath)
			continue
		}

		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Errorf("missing golden file %s — run with UPDATE_GOLDEN=1 to create it", goldenPath)
			continue
		}
		if got != string(want) {
			t.Errorf("output mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
				relPath, want, got)
		}
	}
}

// TestGenerate_AllFilesPresent checks that every expected output file is
// produced for a given set of inputs.
func TestGenerate_AllFilesPresent(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	files, err := gen.Generate(goldenOpts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	expected := []string{
		"cmd/example/main.go",
		"internal/app/app.go",
		"internal/app/app_test.go",
		"config/config.go",
		"transport/server.go",
		"go.mod",
		"Dockerfile",
		"Makefile",
		"README.md",
		"scripts/lint.sh",
		"scripts/test.sh",
		".github/workflows/ci.yml",
	}

	for _, want := range expected {
		if _, ok := files[want]; !ok {
			t.Errorf("expected file %q missing from generator output", want)
		}
	}

	if t.Failed() {
		t.Log("got keys:")
		for k := range files {
			t.Log("  ", k)
		}
	}
}

// TestGenerate_NameIsSubstituted verifies that the service name is correctly
// injected into generated content.
func TestGenerate_NameIsSubstituted(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	opts := Options{
		Name:   "myservice",
		Module: "github.com/acme/myservice",
	}
	files, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// The app entrypoint should reference the module path.
	appGo := files["internal/app/app.go"]
	if !strings.Contains(appGo, opts.Module) {
		t.Errorf("app.go does not contain module %q", opts.Module)
	}
	if !strings.Contains(appGo, opts.Name) {
		t.Errorf("app.go does not contain service name %q", opts.Name)
	}

	// The cmd entrypoint must be at cmd/myservice/main.go.
	if _, ok := files["cmd/myservice/main.go"]; !ok {
		t.Error("expected cmd/myservice/main.go to be present")
	}

	// Makefile must contain the service name.
	if !strings.Contains(files["Makefile"], opts.Name) {
		t.Errorf("Makefile does not contain service name %q", opts.Name)
	}
}

// TestGenerate_Deterministic ensures identical inputs produce identical outputs.
func TestGenerate_Deterministic(t *testing.T) {
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}

	first, err := gen.Generate(goldenOpts)
	if err != nil {
		t.Fatalf("first Generate: %v", err)
	}
	second, err := gen.Generate(goldenOpts)
	if err != nil {
		t.Fatalf("second Generate: %v", err)
	}

	for path, a := range first {
		b, ok := second[path]
		if !ok {
			t.Errorf("second run missing file %q", path)
			continue
		}
		if a != b {
			t.Errorf("output for %q differs between runs", path)
		}
	}
}

// TestOptions_NameTitle checks the PascalCase helper.
func TestOptions_NameTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"orders", "Orders"},
		{"myService", "MyService"},
		{"", ""},
		{"a", "A"},
	}
	for _, tc := range cases {
		got := (Options{Name: tc.in}).NameTitle()
		if got != tc.want {
			t.Errorf("NameTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestOptions_NameUpper checks the UPPER_CASE helper.
func TestOptions_NameUpper(t *testing.T) {
	cases := []struct{ in, want string }{
		{"orders", "ORDERS"},
		{"myService", "MYSERVICE"},
	}
	for _, tc := range cases {
		got := (Options{Name: tc.in}).NameUpper()
		if got != tc.want {
			t.Errorf("NameUpper(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
