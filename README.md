# goapm

Pure-Go port of WebRTC's audio processing module.

- Noise suppression, gain control, high-pass filter and the pipeline that
  chains them, bit-exact with libwebrtc m141.
- Echo cancellation through [goaec3](https://github.com/GetStream/goaec3).
- Memory-safe: no `unsafe`, no cgo, no assembly.

> **Status: in development.** Nothing is usable yet.
