//go:build webrtc_src_oracle && !webrtc_oracle

// The infrastructure the webrtc_src_oracle build replaces: cpu_info, which
// answers from the instruction sets the build targets. Only the resamplers
// of AudioBuffer read it, and the oracle never configures an AudioBuffer
// that resamples.
#include "prelude.h"

#include "rtc_base/cpu_info.h"

namespace webrtc {
namespace cpu_info {

uint32_t DetectNumberOfCores() { return 1; }

bool Supports(ISA isa) {
  switch (isa) {
    case ISA::kSSE2:
#if defined(__x86_64__)
      return true;
#else
      return false;
#endif
    case ISA::kNeon:
#if defined(WEBRTC_HAS_NEON)
      return true;
#else
      return false;
#endif
    default:
      return false;
  }
}

}  // namespace cpu_info
}  // namespace webrtc
