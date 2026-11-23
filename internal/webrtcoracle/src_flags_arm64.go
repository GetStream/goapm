//go:build webrtc_src_oracle && !webrtc_oracle && arm64

package webrtcoracle

// libwebrtc's GN build defines WEBRTC_ARCH_ARM64 and WEBRTC_HAS_NEON on arm64
// (rtc_base/system/arch.h defines neither); the purego build leaves out
// WEBRTC_HAS_NEON, which compiles the scalar paths.

// #cgo CFLAGS: -DWEBRTC_ARCH_ARM64
// #cgo CXXFLAGS: -DWEBRTC_ARCH_ARM64
import "C"
