//go:build webrtc_oracle || webrtc_src_oracle

package webrtcoracle

// #include "smoke_shim.h"
import "C"

import "unsafe"

// SmokeMulAdd returns a*b + c as the C++ of the oracle computes it.
func SmokeMulAdd(a, b, c float32) float32 {
	return float32(C.goapm_smoke_muladd(C.float(a), C.float(b), C.float(c)))
}

// SmokeProcess runs 10 ms frames through the C++ AudioBuffer, its band
// splitting, HighPassFilter and NoiseSuppressor (default config). in holds
// the frames one after the other, each as channels runs of rateHz/100
// samples in [-1, 1]; the result has the same layout.
func SmokeProcess(rateHz, channels int, in []float32) []float32 {
	frame := rateHz / 100 * channels
	frames := len(in) / frame
	if frames == 0 || len(in) != frames*frame {
		panic("webrtcoracle: SmokeProcess needs whole frames")
	}
	out := make([]float32, len(in))
	C.goapm_smoke_process(C.int(rateHz), C.int(channels), C.int(frames),
		(*C.float)(unsafe.Pointer(&in[0])), (*C.float)(unsafe.Pointer(&out[0])))
	return out
}
