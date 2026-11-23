//go:build webrtc_src_oracle && !webrtc_oracle

// Package webrtcoracle is the cgo test oracle of goapm: it exposes the C++
// audio processing module of libwebrtc m141 to the parity tests. The
// webrtc_oracle tag links the prebuilt libwebrtc.a (oracle.go); the
// webrtc_src_oracle tag compiles the reference from the WebRTC sources
// (src_oracle.go). No shipped package imports it.
package webrtcoracle

// The source build compiles the noise suppressor, the high-pass filter,
// AudioBuffer with its band splitting and their common_audio and rtc_base
// dependencies from tmp_check/webrtc-m141 (scripts/fetch-webrtc.sh) on
// darwin and linux, arm64 and amd64. Logging, metrics and cpu_info are the
// only replaced pieces (src_stubs.cc).
//
// The translation units are generated (src_gen_*, scripts/gen-src-oracle.sh);
// the shims (*_shim.cc) and their Go wrappers are shared with the
// webrtc_oracle build. The purego tag leaves out WEBRTC_HAS_NEON, which
// compiles the scalar paths.
//
// Every translation unit starts with prelude.h, which requires clang and
// turns floating-point contraction off for the whole unit, as libwebrtc's
// -ffp-contract=off does (cgo does not accept that flag without
// CGO_CFLAGS_ALLOW).

/*
#cgo CFLAGS: -DNDEBUG -DWEBRTC_POSIX -DWEBRTC_ENABLE_AVX2 -I${SRCDIR}/../../tmp_check/webrtc-m141
#cgo CXXFLAGS: -std=c++20 -DNDEBUG -DWEBRTC_POSIX -DWEBRTC_ENABLE_AVX2 -DWEBRTC_APM_DEBUG_DUMP=0 -DRTC_DISABLE_LOGGING -DRTC_DISABLE_METRICS -DRTC_DISABLE_TRACE_EVENTS -Wno-nullability-completeness -I${SRCDIR}/../../tmp_check/webrtc-m141 -I${SRCDIR}/../../tmp_check/webrtc-m141/third_party/abseil-cpp
#cgo darwin CFLAGS: -DWEBRTC_MAC
#cgo darwin CXXFLAGS: -DWEBRTC_MAC
#cgo darwin LDFLAGS: -lc++
#cgo linux CFLAGS: -DWEBRTC_LINUX
#cgo linux CXXFLAGS: -DWEBRTC_LINUX
#cgo linux LDFLAGS: -lstdc++ -lm
*/
import "C"

// SourceBuild reports whether the C++ reference is compiled from the WebRTC
// sources.
const SourceBuild = true
