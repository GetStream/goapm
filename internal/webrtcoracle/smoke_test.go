//go:build webrtc_oracle || webrtc_src_oracle

package webrtcoracle

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand/v2"
	"testing"
)

// TestSmokeContractionOff checks that the C++ of the oracle does not fuse a
// multiplication and an addition, like libwebrtc (-ffp-contract=off).
func TestSmokeContractionOff(t *testing.T) {
	// a*b is 1 - 2^-46 exactly and rounds to 1, so the unfused sum is 0.
	a, b, c := float32(1+1.0/(1<<23)), float32(1-1.0/(1<<23)), float32(-1)
	if got := SmokeMulAdd(a, b, c); got != 0 {
		t.Fatalf("a*b + c = %g: the C++ is compiled with floating-point contraction", got)
	}
	if fused := math.FMA(float64(a), float64(b), float64(c)); fused == 0 {
		t.Fatal("the probe does not tell a fused multiply-add from an unfused one")
	}
}

// smokeSignal returns frames of a 440 Hz tone that starts after a second, on
// stationary noise with a constant offset, per channel.
func smokeSignal(rateHz, channels, frames int) []float32 {
	rng := rand.New(rand.NewPCG(1, uint64(rateHz+channels)))
	n := rateHz / 100
	x := make([]float32, frames*channels*n)
	for f := 0; f < frames; f++ {
		for ch := 0; ch < channels; ch++ {
			s := x[(f*channels+ch)*n:][:n]
			for i := range s {
				v := 0.02*(rng.Float64()*2-1) + 0.01
				if f >= 100 && f < 150 {
					v += 0.3 * math.Sin(2*math.Pi*440*float64(f*n+i)/float64(rateHz))
				}
				s[i] = float32(v)
			}
		}
	}
	return x
}

// smokeHashes are the FNV-1a hashes of the output bits of TestSmokeEndToEnd,
// by sample rate and number of channels. They are the same for the prebuilt
// libwebrtc and for the source build on every platform.
var smokeHashes = map[[2]int]uint64{
	{16000, 1}: 0x66c95d26b1c58ab7,
	{16000, 2}: 0xbf8973723264f142,
	{32000, 1}: 0x7877c03fa7f035c5,
	{32000, 2}: 0xf81e92ca76e44e51,
	{48000, 1}: 0x2c466904ed52ca1f,
	{48000, 2}: 0xb0635e3bcd89761d,
}

// TestSmokeEndToEnd runs the C++ capture chain that goapm ports first
// (AudioBuffer, band splitting, HighPassFilter, NoiseSuppressor) and checks
// what any working build of it does: the output is finite and reproducible,
// the constant offset is gone and the stationary noise is attenuated. With
// the Go transcendentals interposed, the output is also the same bit for
// bit in every build of the oracle.
func TestSmokeEndToEnd(t *testing.T) {
	for _, rateHz := range []int{16000, 32000, 48000} {
		for channels := 1; channels <= 2; channels++ {
			t.Run(fmt.Sprintf("%dHz_%dch", rateHz, channels), func(t *testing.T) {
				const frames = 300
				n := rateHz / 100
				in := smokeSignal(rateHz, channels, frames)
				calls := LibmCalls()
				out := SmokeProcess(rateHz, channels, in)
				if !StockLibm && LibmCalls() == calls {
					t.Fatal("the C++ noise suppressor did not call the Go transcendentals")
				}
				again := SmokeProcess(rateHz, channels, in)
				for i := range out {
					if math.Float32bits(out[i]) != math.Float32bits(again[i]) {
						t.Fatalf("sample %d differs between two runs: %g, %g", i, out[i], again[i])
					}
					if v := float64(out[i]); math.IsNaN(v) || math.IsInf(v, 0) {
						t.Fatalf("sample %d is %g", i, out[i])
					}
				}
				h := fnv.New64a()
				var b [4]byte
				for _, v := range out {
					u := math.Float32bits(v)
					b[0], b[1], b[2], b[3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
					h.Write(b[:])
				}
				if want := smokeHashes[[2]int{rateHz, channels}]; StockLibm {
					t.Logf("host libm: output hash %#x, with the Go transcendentals %#x", h.Sum64(), want)
				} else if h.Sum64() != want {
					t.Errorf("output hash %#x, want %#x", h.Sum64(), want)
				}
				// The last second: noise only.
				var sumIn, sumOut, energyIn, energyOut float64
				tail := (frames - 100) * channels * n
				for i := tail; i < len(in); i++ {
					sumIn += float64(in[i])
					sumOut += float64(out[i])
				}
				meanIn, meanOut := sumIn/float64(len(in)-tail), sumOut/float64(len(in)-tail)
				for i := tail; i < len(in); i++ {
					energyIn += (float64(in[i]) - meanIn) * (float64(in[i]) - meanIn)
					energyOut += (float64(out[i]) - meanOut) * (float64(out[i]) - meanOut)
				}
				if math.Abs(meanOut) > 0.1*math.Abs(meanIn) {
					t.Errorf("mean %g in, %g out: the high-pass filter did not remove the offset", meanIn, meanOut)
				}
				if db := 10 * math.Log10(energyOut/energyIn); db > -6 || db < -30 {
					t.Errorf("noise changed by %.1f dB, want an attenuation of 6 to 30 dB", db)
				}
			})
		}
	}
}
