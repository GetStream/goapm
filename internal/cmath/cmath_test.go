package cmath

import (
	"math"
	"testing"
)

func TestExp2f(t *testing.T) {
	for e := -149; e <= 127; e++ {
		if got, want := Exp2f(float32(e)), float32(math.Ldexp(1, e)); got != want {
			t.Fatalf("Exp2f(%d) = %g, want %g", e, got, want)
		}
	}
	if got := Exp2f(128); !math.IsInf(float64(got), 1) {
		t.Fatalf("Exp2f(128) = %g, want +Inf", got)
	}
	if got := Exp2f(-150); got != 0 {
		t.Fatalf("Exp2f(-150) = %g, want 0", got)
	}
	if got := Exp2f(float32(math.NaN())); got == got {
		t.Fatalf("Exp2f(NaN) = %g, want NaN", got)
	}
}

func TestTanhf(t *testing.T) {
	for _, c := range []struct{ x, want float32 }{
		{0, 0}, {float32(math.Inf(1)), 1}, {float32(math.Inf(-1)), -1}, {20, 1}, {-20, -1},
	} {
		if got := Tanhf(c.x); got != c.want {
			t.Fatalf("Tanhf(%g) = %g, want %g", c.x, got, c.want)
		}
	}
	if got := Tanhf(float32(math.Copysign(0, -1))); got != 0 || !math.Signbit(float64(got)) {
		t.Fatalf("Tanhf(-0) = %g, want -0", got)
	}
	// tanh is odd and monotonic.
	prev := float32(-1)
	for x := float32(-10); x <= 10; x += 1.0 / 64 {
		y := Tanhf(x)
		if y < prev {
			t.Fatalf("Tanhf(%g) = %g < Tanhf of a smaller argument %g", x, y, prev)
		}
		if y != -Tanhf(-x) {
			t.Fatalf("Tanhf(%g) != -Tanhf(%g)", x, -x)
		}
		prev = y
	}
}

func TestSqrtf(t *testing.T) {
	for i := 0; i < 1000; i++ {
		x := float32(i) * float32(i)
		if got := Sqrtf(x); got != float32(i) {
			t.Fatalf("Sqrtf(%g) = %g", x, got)
		}
	}
}

func TestMinMax(t *testing.T) {
	nan := float32(math.NaN())
	if Min(1, 2) != 1 || Min(2, 1) != 1 || Max(1, 2) != 2 || Max(2, 1) != 2 {
		t.Fatal("Min or Max of two numbers")
	}
	// The first operand wins when either is NaN, as in std::min and std::max.
	if Min(3, nan) != 3 || Max(3, nan) != 3 {
		t.Fatal("Min or Max with NaN as the second operand")
	}
	if v := Min(nan, 3); v == v {
		t.Fatalf("Min(NaN, 3) = %g, want NaN", v)
	}
	if v := Max(nan, 3); v == v {
		t.Fatalf("Max(NaN, 3) = %g, want NaN", v)
	}
	// Equal operands return the first: Min(+0, -0) is +0, Max(-0, +0) is -0.
	negZero := float32(math.Copysign(0, -1))
	if math.Signbit(float64(Min(0, negZero))) || !math.Signbit(float64(Max(negZero, 0))) {
		t.Fatal("Min or Max of zeros of both signs")
	}
}

func TestAtanf(t *testing.T) {
	// fft4g.cc derives its table step from atanf(1.0f), pi/4.
	if got := Atanf(1); math.Float32bits(got) != 0x3f490fdb {
		t.Fatalf("Atanf(1) = %#x, want 0x3f490fdb", math.Float32bits(got))
	}
	if Atanf(0) != 0 || Atanf(-1) != -Atanf(1) {
		t.Fatal("Atanf is not odd through 0")
	}
}
