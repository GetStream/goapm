package goapm

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestNoUnsafe fails if a non-test Go file of the module imports "unsafe".
// It reads the import clauses of every file, whatever its build constraints,
// so it covers every lane, architecture and GOEXPERIMENT setting at once.
// internal/webrtcoracle, the cgo harness that links the C++ reference into
// the parity tests, is the one exception; no shipped package depends on it.
func TestNoUnsafe(t *testing.T) {
	const oracle = "internal/webrtcoracle"
	fset := token.NewFileSet()
	files := 0
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != "." && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") ||
				name == "testdata" || name == "tmp_check" || filepath.ToSlash(path) == oracle) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		files++
		for _, imp := range f.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if p == "unsafe" {
				t.Errorf("%s imports unsafe", fset.Position(imp.Pos()))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 2 {
		t.Fatalf("scanned only %d files; the walk did not cover the module", files)
	}
}
