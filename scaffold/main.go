package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	var (
		name   = flag.String("name", "", "service name, e.g. orders (required)")
		module = flag.String("module", "", "Go module path, e.g. github.com/acme/orders (required)")
		outDir = flag.String("out", ".", "parent directory to write the generated service into")
		dryRun = flag.Bool("dry-run", false, "print output file paths without writing anything")
		libMod = flag.String("golib-module", defaultGoLibModule, "go-lib module import path override")
		goVer  = flag.String("go-version", defaultGoVersion, "minimum Go version in generated go.mod")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: scaffold [flags]\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n  go run ./scaffold --name orders --module github.com/acme/orders\n")
	}
	flag.Parse()

	if *name == "" || *module == "" {
		flag.Usage()
		os.Exit(1)
	}
	if strings.ContainsAny(*name, "/\\ ") {
		fmt.Fprintln(os.Stderr, "error: --name must not contain path separators or spaces")
		os.Exit(1)
	}

	gen, err := NewGenerator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing generator: %v\n", err)
		os.Exit(1)
	}

	files, err := gen.Generate(Options{
		Name:        *name,
		Module:      *module,
		GoLibModule: *libMod,
		GoVersion:   *goVer,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "generation error: %v\n", err)
		os.Exit(1)
	}

	// Sort paths for deterministic output.
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	if *dryRun {
		fmt.Println("-- dry run (no files written) --")
		for _, p := range paths {
			fmt.Printf("  %s/%s/%s\n", filepath.Clean(*outDir), *name, p)
		}
		return
	}

	root := filepath.Join(filepath.Clean(*outDir), *name)
	for _, p := range paths {
		dest := filepath.Join(root, p)

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", filepath.Dir(dest), err)
			os.Exit(1)
		}
		var flags os.FileMode = 0o644
		if strings.HasSuffix(dest, ".sh") {
			flags = 0o755
		}
		if err := os.WriteFile(dest, []byte(files[p]), flags); err != nil { // nolint:gosec // file permissions are not sensitive
			fmt.Fprintf(os.Stderr, "write %s: %v\n", dest, err)
			os.Exit(1)
		}
		fmt.Printf("  created  %s\n", dest)
	}
	fmt.Printf("\nService %q scaffolded into %s\nNext steps:\n", *name, root)
	fmt.Printf("cd %s && make generate &&"+
		" go mod tidy && make lint && make test\n", root)
}
