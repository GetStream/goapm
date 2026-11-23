// Package cmath holds the float transcendentals of the signal path.
//
// libwebrtc calls the host libm (clang lowers powf(2.f, x) to exp2f, and
// (float)cos(x) of a float x to cosf), whose float results are not correctly
// rounded and differ between platforms. The port uses the rounded result of
// the float64 function, and the oracle binary interposes these functions
// for the libm symbols.
package cmath

import "math"

// Exp2f returns 2**x rounded to float32.
func Exp2f(x float32) float32 { return float32(math.Exp2(float64(x))) }

// Tanhf returns the hyperbolic tangent of x rounded to float32.
func Tanhf(x float32) float32 { return float32(math.Tanh(float64(x))) }

// Sqrtf returns the square root of x, correctly rounded like C's sqrtf.
func Sqrtf(x float32) float32 { return float32(math.Sqrt(float64(x))) }

// Cosf returns the cosine of x rounded to float32.
func Cosf(x float32) float32 { return float32(math.Cos(float64(x))) }

// Sinf returns the sine of x rounded to float32.
func Sinf(x float32) float32 { return float32(math.Sin(float64(x))) }

// Min returns std::min(a, b): b if b < a, else a. With a NaN operand the
// result is a.
func Min(a, b float32) float32 {
	if b < a {
		return b
	}
	return a
}

// Max returns std::max(a, b): b if a < b, else a. With a NaN operand the
// result is a.
func Max(a, b float32) float32 {
	if a < b {
		return b
	}
	return a
}
