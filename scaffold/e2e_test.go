package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// goLibRoot returns the absolute path of the go-lib module root so the
// generated service can point its replace directive at it during compilation.
func goLibRoot(t *testing.T) string {
	t.Helper()
	// The scaffold package lives at <go-lib root>/scaffold/, so go two levels up
	// from this file's directory.
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file == <go-lib>/scaffold/e2e_test.go → parent is <go-lib>/scaffold/ → parent is <go-lib>/
	return filepath.Dir(filepath.Dir(file))
}

// generateToDir calls Generate and writes all produced files under outDir/name,
// exactly mirroring what main() does.
func generateToDir(t *testing.T, opts Options, outDir string) string {
	t.Helper()
	gen, err := NewGenerator()
	if err != nil {
		t.Fatalf("NewGenerator: %v", err)
	}
	files, err := gen.Generate(opts)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	root := filepath.Join(outDir, opts.Name)
	for relPath, content := range files {
		dest := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(dest), err)
		}
		if err := os.WriteFile(dest, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", dest, err)
		}
	}
	return root
}

// TestE2E_WritesExpectedFileTree generates a service to a temp directory and
// asserts that every expected file is present and non-empty.
func TestE2E_WritesExpectedFileTree(t *testing.T) {
	opts := Options{
		Name:   "alpha",
		Module: "github.com/e2e/alpha",
	}
	dir := t.TempDir()
	root := generateToDir(t, opts, dir)

	wantPaths := []string{
		"api/example.proto",
		"api/openapi.yml",
		"cmd/alpha/main.go",
		"config/config.go",
		"data/data.go",
		"data/data_suite_test.go",
		"data/db/db.go",
		"data/db/db_suite_test.go",
		"transport/rpc/rpc.go",
		"transport/rpc/rpc_suite_test.go",
		"transport/http/http.go",
		"transport/http/http_suite_test.go",
		"go.mod",
		"Dockerfile",
		"Makefile",
		"README.md",
		"scripts/generate.sh",
		"scripts/lint.sh",
		"scripts/test.sh",
		".github/workflows/ci.yml",
	}

	for _, rel := range wantPaths {
		full := filepath.Join(root, rel)
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("missing expected file %s: %v", rel, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s exists but is empty", rel)
		}
	}
}

// TestE2E_FileContentsAreCoherent spot-checks that critical substitutions
// appear in the written files.
func TestE2E_FileContentsAreCoherent(t *testing.T) {
	opts := Options{
		Name:   "billing",
		Module: "github.com/acme/billing",
	}
	dir := t.TempDir()
	root := generateToDir(t, opts, dir)

	checks := []struct {
		file    string
		contain []string
	}{
		{"config/config.go", []string{opts.GoLibModule}},
		{"transport/http/http.go", []string{opts.Name, opts.GoLibModule}},
		{"go.mod", []string{"module " + opts.Module, opts.GoLibModule}},
		{"Dockerfile", []string{"billing", "BILLING_"}},
		{"Makefile", []string{"billing", "BILLING_"}},
		{"README.md", []string{"billing", "BILLING_"}},
		{".github/workflows/ci.yml", []string{"billing", "CI"}},
	}

	for _, tc := range checks {
		raw, err := os.ReadFile(filepath.Join(root, tc.file))
		if err != nil {
			t.Errorf("cannot read %s: %v", tc.file, err)
			continue
		}
		content := string(raw)
		for _, want := range tc.contain {
			if !strings.Contains(content, want) {
				t.Errorf("%s: expected to contain %q", tc.file, want)
			}
		}
	}
}

// TestE2E_NoStrayTemplateSyntax checks that no unrendered Go template markers
// remain in any generated file.
func TestE2E_NoStrayTemplateSyntax(t *testing.T) {
	opts := Options{
		Name:   "gamma",
		Module: "github.com/e2e/gamma",
	}
	dir := t.TempDir()
	root := generateToDir(t, opts, dir)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(raw)
		if strings.Contains(content, "{{.") {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("file %s still contains unrendered template marker {{.", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestE2E_DeterminsticLayout verifies that two calls with the same inputs
// produce an identical set of files on disk (same paths, same content).
func TestE2E_DeterminsticLayout(t *testing.T) {
	opts := Options{
		Name:   "delta",
		Module: "github.com/e2e/delta",
	}

	dir1 := t.TempDir()
	dir2 := t.TempDir()
	root1 := generateToDir(t, opts, dir1)
	root2 := generateToDir(t, opts, dir2)

	collectFiles := func(root string) map[string]string {
		m := make(map[string]string)
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			raw, _ := os.ReadFile(path)
			m[rel] = string(raw)
			return nil
		})
		return m
	}

	files1 := collectFiles(root1)
	files2 := collectFiles(root2)

	keys1 := make([]string, 0, len(files1))
	for k := range files1 {
		keys1 = append(keys1, k)
	}
	sort.Strings(keys1)

	for _, rel := range keys1 {
		if files1[rel] != files2[rel] {
			t.Errorf("content differs between runs for %s", rel)
		}
	}
	if len(files1) != len(files2) {
		t.Errorf("file counts differ: first=%d second=%d", len(files1), len(files2))
	}
}

// TestE2E_DryRun executes the scaffold binary with --dry-run and confirms it
// prints paths but writes nothing to disk.
func TestE2E_DryRun(t *testing.T) {
	// Build the binary into a temp dir so we can exec it.
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "scaffold")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join(goLibRoot(t), "scaffold")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build scaffold binary: %v\n%s", err, out)
	}

	outDir := t.TempDir()
	cmd := exec.Command(binaryPath,
		"--name", "drytest",
		"--module", "github.com/e2e/drytest",
		"--out", outDir,
		"--dry-run",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dry-run exited with error: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "dry run") {
		t.Errorf("expected 'dry run' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "drytest") {
		t.Errorf("expected service name in dry-run output, got:\n%s", output)
	}

	// Nothing should have been written on disk.
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Errorf("--dry-run wrote files to disk: %v", names)
	}
}

// TestE2E_InvalidInput verifies the CLI rejects missing required flags.
func TestE2E_InvalidInput(t *testing.T) {
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "scaffold")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = filepath.Join(goLibRoot(t), "scaffold")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build scaffold binary: %v\n%s", err, out)
	}

	cases := []struct {
		desc string
		args []string
	}{
		{"missing both flags", []string{}},
		{"missing --module", []string{"--name", "foo"}},
		{"missing --name", []string{"--module", "github.com/foo/foo"}},
		{"name with slash", []string{"--name", "bad/name", "--module", "github.com/x/x"}},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			if err := cmd.Run(); err == nil {
				t.Errorf("expected non-zero exit for %q, but got success", tc.desc)
			}
		})
	}
}

// TestE2E_Compiles generates a service, injects a replace directive pointing
// at the local go-lib source, and verifies the generated code compiles.
// Skipped in short mode because it invokes the Go toolchain.
func TestE2E_Compiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation test in -short mode")
	}

	opts := Options{
		Name:        "epsilon",
		Module:      "github.com/e2e/epsilon",
		GoLibModule: defaultGoLibModule,
		GoVersion:   defaultGoVersion,
	}
	dir := t.TempDir()
	root := generateToDir(t, opts, dir)
	libRoot := goLibRoot(t)

	// Inject replace directive so the generated module resolves go-lib locally.
	// this is a test-only replace directive, not a security-sensitive operation
	editCmd := exec.Command("go", "mod", "edit", //nolint:gosec // see above
		"-replace="+opts.GoLibModule+"="+libRoot,
	)
	editCmd.Dir = root
	if out, err := editCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod edit: %v\n%s", err, out)
	}

	// Seed go.sum from go-lib's own go.sum so tidy works without downloading
	// the full transitive graph from scratch.
	libSum, err := os.ReadFile(filepath.Join(libRoot, "go.sum"))
	if err != nil {
		t.Fatalf("read go-lib go.sum: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.sum"), libSum, 0o600); err != nil {
		t.Fatalf("write go.sum: %v", err)
	}

	generateCmd := exec.Command("bash", "scripts/generate.sh")
	generateCmd.Dir = root
	if out, err := generateCmd.CombinedOutput(); err != nil {
		t.Fatalf("generate.sh: %v\n%s", err, out)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = root
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = root
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build ./...: %v\n%s", err, out)
	}
}
