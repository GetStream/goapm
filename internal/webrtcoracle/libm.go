//go:build webrtc_oracle || webrtc_src_oracle

package webrtcoracle

// The libm symbols that call these exports are defined in libm_interpose.c.

import "C"

import (
	"sync/atomic"

	"github.com/GetStream/goapm/internal/cmath"
)

// libmCalls counts the calls the C++ makes into the Go transcendentals.
var libmCalls atomic.Int64

// LibmCalls returns the number of libm calls of the C++ reference that the
// Go implementations in internal/cmath answered.
func LibmCalls() int64 { return libmCalls.Load() }

//export goapmExp2f
func goapmExp2f(x C.float) C.float {
	libmCalls.Add(1)
	return C.float(cmath.Exp2f(float32(x)))
}

//export goapmTanhf
func goapmTanhf(x C.float) C.float {
	libmCalls.Add(1)
	return C.float(cmath.Tanhf(float32(x)))
}

//export goapmCosf
func goapmCosf(x C.float) C.float {
	libmCalls.Add(1)
	return C.float(cmath.Cosf(float32(x)))
}

//export goapmSinf
func goapmSinf(x C.float) C.float {
	libmCalls.Add(1)
	return C.float(cmath.Sinf(float32(x)))
}
