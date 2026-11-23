//go:build webrtc_src_oracle && !webrtc_oracle && arm64 && !purego

package webrtcoracle

// #cgo CFLAGS: -DWEBRTC_HAS_NEON
// #cgo CXXFLAGS: -DWEBRTC_HAS_NEON
import "C"
