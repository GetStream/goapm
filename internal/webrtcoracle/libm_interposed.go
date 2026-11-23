//go:build (webrtc_oracle || webrtc_src_oracle) && !webrtc_oracle_stock

package webrtcoracle

// StockLibm reports whether the C++ reference calls the host libm instead
// of the Go implementations in internal/cmath.
const StockLibm = false
