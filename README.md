# goapm

Pure-Go port of WebRTC's audio processing module.

- Noise suppression, gain control, high-pass filter and the pipeline that
  chains them, bit-exact with libwebrtc m141.
- Echo cancellation through [goaec3](https://github.com/GetStream/goaec3).
- Memory-safe: no `unsafe`, no cgo, no assembly.

> **Status: in development.** Nothing is usable yet.

## Parity

Every module is compared bit for bit with libwebrtc, both prebuilt and built
from source, before the next one builds on it. Details in
[AGENTS.md](AGENTS.md).

```sh
make test test-simd test-purego   # unit tests
make fetch-webrtc oracle-src      # parity against libwebrtc built from source
```

Requires Go 1.27+.

## License

BSD 3-Clause, see [LICENSE](LICENSE). Derived from WebRTC. Security
policy in [SECURITY.md](SECURITY.md).
