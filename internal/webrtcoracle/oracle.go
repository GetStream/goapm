//go:build webrtc_oracle

// Package webrtcoracle is the cgo test oracle of goapm: it exposes the C++
// audio processing module of libwebrtc m141 to the parity tests. The
// webrtc_oracle tag links the prebuilt libwebrtc.a (oracle.go); the
// webrtc_src_oracle tag compiles the reference from the WebRTC sources
// (src_oracle.go). No shipped package imports it.
package webrtcoracle

// The prebuilt library is tmp_check/libwebrtc (m141, darwin/arm64, NEON,
// built with -ffp-contract=off). The shims (*_shim.cc) and their Go wrappers
// are shared with the webrtc_src_oracle build.

/*
#cgo CFLAGS: -DNDEBUG
#cgo CXXFLAGS: -std=c++20 -DNDEBUG -Wno-nullability-completeness -DWEBRTC_APM_DEBUG_DUMP=0 -DWEBRTC_POSIX -DWEBRTC_MAC -DWEBRTC_HAS_NEON -I${SRCDIR}/../../tmp_check/libwebrtc/include -I${SRCDIR}/../../tmp_check/libwebrtc/include/third_party/abseil-cpp
// libwebrtc's GN build defines WEBRTC_ARCH_ARM64 (rtc_base/system/arch.h does
// not); without it NEON header code compiles a different path than
// libwebrtc.a.
#cgo arm64 CXXFLAGS: -DWEBRTC_ARCH_ARM64
#cgo LDFLAGS: ${SRCDIR}/../../tmp_check/libwebrtc/lib/libwebrtc.a -lc++ -framework Foundation -framework CoreAudio -framework AudioToolbox -framework CoreMedia -framework CoreVideo -framework AVFoundation -framework AppKit -framework IOSurface -framework Metal -framework VideoToolbox -framework ApplicationServices
*/
import "C"

// SourceBuild reports whether the C++ reference is compiled from the WebRTC
// sources.
const SourceBuild = false
