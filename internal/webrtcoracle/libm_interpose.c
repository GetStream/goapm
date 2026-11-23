//go:build (webrtc_oracle || webrtc_src_oracle) && !webrtc_oracle_stock

// Interposes the libm entry points the ported modules reach, so that the C++
// reference uses exactly the Go implementations in internal/cmath (host libm
// float results are not correctly rounded and differ between libms). How
// clang lowers powf(2.f, x) depends on the target libm: exp2f, or powf
// itself. Both spellings are covered; powf with another base aborts, since
// no Go helper mirrors it. src_main_test.go checks that the test binary of
// the source build imports none of these from libm. The webrtc_oracle_stock
// tag leaves them out, to measure goapm against the C++ on the host libm.
//
//
// fft4g.cc builds the tables of the 256-point FFT with (float)cos(x) and
// (float)sin(x) of a float x, which clang narrows to cosf and sinf, and
// pairs to __sincosf_stret on Darwin and sincosf on glibc. A build that
// keeps the double functions gets the float64 results the Go helpers round.
#include <stdio.h>
#include <stdlib.h>

float goapmExp2f(float);
float goapmTanhf(float);
float goapmCosf(float);
float goapmSinf(float);

float exp2f(float x) { return goapmExp2f(x); }
float tanhf(float x) { return goapmTanhf(x); }
float cosf(float x) { return goapmCosf(x); }
float sinf(float x) { return goapmSinf(x); }

float powf(float base, float x) {
  if (base == 2.0f) return goapmExp2f(x);
  fprintf(stderr, "goapm oracle: powf(%a, %a) has no Go counterpart\n", base, x);
  abort();
}

#if defined(__APPLE__)
struct goapm_sincosf {
  float sin, cos;
};
struct goapm_sincosf __sincosf_stret(float x) {
  struct goapm_sincosf r = {goapmSinf(x), goapmCosf(x)};
  return r;
}
#else
void sincosf(float x, float* s, float* c) {
  *s = goapmSinf(x);
  *c = goapmCosf(x);
}
#endif

// The arguments the sources pass to the double functions are floats.
static float goapm_narrow(const char* name, double x) {
  if ((double)(float)x != x) {
    fprintf(stderr, "goapm oracle: %s(%a) has no Go counterpart\n", name, x);
    abort();
  }
  return (float)x;
}
double cos(double x) { return goapmCosf(goapm_narrow("cos", x)); }
double sin(double x) { return goapmSinf(goapm_narrow("sin", x)); }
