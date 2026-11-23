#ifndef GOAPM_SMOKE_SHIM_H
#define GOAPM_SMOKE_SHIM_H
#ifdef __cplusplus
extern "C" {
#endif

// a * b + c as a translation unit of the oracle compiles it.
float goapm_smoke_muladd(float a, float b, float c);

// Runs frames 10 ms frames of channels channels at rate Hz through
// AudioBuffer, its band splitting, HighPassFilter and NoiseSuppressor, as
// AudioProcessingImpl chains them. in and out hold the frames one after the
// other, each as channels runs of rate/100 samples in [-1, 1].
void goapm_smoke_process(int rate, int channels, int frames, const float* in, float* out);

#ifdef __cplusplus
}
#endif
#endif
