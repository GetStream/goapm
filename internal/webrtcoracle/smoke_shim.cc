//go:build webrtc_oracle || webrtc_src_oracle

#include "prelude.h"

#include <vector>

#include "api/audio/audio_processing.h"
#include "modules/audio_processing/audio_buffer.h"
#include "modules/audio_processing/high_pass_filter.h"
#include "modules/audio_processing/ns/noise_suppressor.h"
#include "modules/audio_processing/ns/ns_config.h"
#include "smoke_shim.h"

extern "C" float goapm_smoke_muladd(float a, float b, float c) {
  volatile float va = a, vb = b, vc = c;
  return va * vb + vc;
}

extern "C" void goapm_smoke_process(int rate, int channels, int frames, const float* in, float* out) {
  const size_t num_channels = static_cast<size_t>(channels);
  const size_t frame_size = static_cast<size_t>(rate / 100);
  const webrtc::StreamConfig stream(rate, num_channels);
  webrtc::AudioBuffer audio(rate, num_channels, rate, num_channels, rate, num_channels);
  // The filter runs on the lowest band, as with a multi-band capture stream.
  webrtc::HighPassFilter hpf(16000, num_channels);
  webrtc::NoiseSuppressor ns(webrtc::NsConfig(), rate, num_channels);
  // A 16 kHz stream has one band and no splitting filter.
  const bool multi_band = rate > 16000;
  std::vector<const float*> src(num_channels);
  std::vector<float*> dst(num_channels);
  for (int f = 0; f < frames; ++f) {
    for (size_t ch = 0; ch < num_channels; ++ch) {
      src[ch] = in + (static_cast<size_t>(f) * num_channels + ch) * frame_size;
      dst[ch] = out + (static_cast<size_t>(f) * num_channels + ch) * frame_size;
    }
    audio.CopyFrom(src.data(), stream);
    if (multi_band) audio.SplitIntoFrequencyBands();
    hpf.Process(&audio, true);
    ns.Analyze(audio);
    ns.Process(&audio);
    if (multi_band) audio.MergeFrequencyBands();
    audio.CopyTo(stream, dst.data());
  }
}
