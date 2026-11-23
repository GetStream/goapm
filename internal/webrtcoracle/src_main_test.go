//go:build webrtc_src_oracle && !webrtc_oracle

package webrtcoracle

import (
	"debug/elf"
	"debug/macho"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

// srcLibmImports lists the libm functions the test binary imports that the
// oracle interposes (libm_interpose.c), or that no Go helper mirrors: any
// of them would mean a C++ call reaches the host libm.
func srcLibmImports() ([]string, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	var syms []string
	if runtime.GOOS == "darwin" {
		f, err := macho.Open(exe)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if syms, err = f.ImportedSymbols(); err != nil {
			return nil, err
		}
	} else {
		f, err := elf.Open(exe)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		imp, err := f.ImportedSymbols()
		if err != nil {
			return nil, err
		}
		for _, s := range imp {
			syms = append(syms, s.Name)
		}
	}
	interposed := map[string]bool{"exp2f": true, "tanhf": true, "powf": true, "pow": true, "exp2": true,
		"tanh": true, "__exp10f": true, "exp10f": true, "log10f": true, "expf": true, "logf": true,
		"log2f": true, "sinf": true, "cosf": true, "sin": true, "cos": true, "sincosf": true,
		"sincos": true, "__sincosf_stret": true, "__sincos_stret": true, "atanf": true, "atan": true,
		"exp": true, "log": true, "log10": true, "log2": true}
	var bad []string
	for _, s := range syms {
		if interposed[strings.TrimPrefix(s, "_")] {
			bad = append(bad, s)
		}
	}
	return bad, nil
}

func TestMain(m *testing.M) {
	bad, err := srcLibmImports()
	switch {
	case StockLibm && (err != nil || len(bad) == 0):
		// The stock build must reach the host libm, or it measures nothing.
		fmt.Fprintln(os.Stderr, "webrtc_src_oracle: the webrtc_oracle_stock build imports no libm function", err)
		os.Exit(1)
	case StockLibm:
		fmt.Fprintln(os.Stderr, "webrtc_src_oracle: host libm imports", bad)
	case err != nil || len(bad) > 0:
		fmt.Fprintln(os.Stderr, "webrtc_src_oracle: libm imports", bad, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "webrtc_src_oracle: %s/%s, purego %v\n", runtime.GOOS, runtime.GOARCH, purego)
	os.Exit(m.Run())
}
