# Agent Instructions

goapm is a pure-Go port of WebRTC's audio processing module
(`modules/audio_processing`) targeting **bit-exact parity with libwebrtc
m141** (`branch-heads/7390`, commit `d2eaa5570fc9959f8dbde32912a16366b8ee75f4`).
AEC3 is not ported here: it comes from
[goaec3](https://github.com/GetStream/goaec3), whose layout and discipline
goapm follows (`../goaec3/AGENTS.md`).

State current behavior in the present tense in code comments and docs. Do not
write changelog narrative. Comments follow the Go standard library: full
sentences that say what the code does and why, starting with the name they
document.

## Goals

- Bit-exact with libwebrtc m141, proven against the C++ oracle module by
  module.
- Faster than libwebrtc: archsimd kernels where the C++ has SIMD, and a fast
  pure-Go fallback.
- Invisible to the garbage collector: pointer-free state in few objects.
- Always safe: no `unsafe`, no assembly, no cgo in the library, zero
  allocations per frame, every vector instruction behind a test of its CPU
  feature.
- Work the C++ does is skipped only if the outputs and the state are
  provably identical, and a test proves it.

## Prime directive: parity, proven against a live C++ oracle

- The reference sources live in `tmp_check/webrtc-m141/` (sparse checkout; not
  committed; `scripts/fetch-webrtc.sh`). `tmp_check/libwebrtc` is a symlink to
  the prebuilt `libwebrtc.a` (m141, darwin/arm64, NEON, built with
  `-ffp-contract=off`). `tmp_check` may be a symlink to goaec3's: both
  repositories pin the same commit and the same paths.
- `internal/webrtcoracle` (cgo) exposes C++ entry points to the parity tests:
  the `webrtc_oracle` tag links the prebuilt library, the `webrtc_src_oracle`
  tag compiles the reference from the checked-out sources. Every translation
  unit starts with `prelude.h`: clang, libc++, `NDEBUG`,
  `#pragma clang fp contract(off)`. libc++ matters: the sources call `tanh`
  on a float unqualified, which is `tanhf` with libc++ (libwebrtc's standard
  library on every platform) and the double `tanh` with libstdc++.
- Every ported module gets a differential test against the oracle before the
  next module builds on it. Kernels and modules must be **bit-exact**
  (`math.Float32bits` equality), state included where the oracle can read
  it. **Never weaken a gate to make work pass.**
- The shims (`*_shim.cc`, `*_shim.h`), their Go wrappers and the parity
  tests carry the constraint `webrtc_oracle || webrtc_src_oracle` and run
  unchanged in both builds.

## Lanes

- `purego` build tag: the scalar C++ path. Bit-exact on every architecture.
- Default build: reproduces the SIMD path libwebrtc selects on that
  architecture (NEON on arm64, SSE2/AVX2 on amd64). SIMD variants in WebRTC
  reorder reductions, so the Go code for each lane follows that lane's
  operation order exactly.
- A module that is scalar in the C++ (the high-pass filter, the noise
  suppressor) has one lane and the same code in every build. SIMD goes only
  where the C++ has SIMD, or where a vector form is provably the scalar
  order (independent elements, no reduction).
- C++ classes that take a CPU-feature argument keep it in Go so tests can
  drive every lane against the oracle in one binary.

## SIMD: GOEXPERIMENT=simd

- Toolchain: Go 1.27.4 (latest 1.27.x). Kernels are written in Go with
  `simd/archsimd`, which exists only when building with `GOEXPERIMENT=simd`
  and covers both arm64 (NEON) and amd64 (SSE/AVX/AVX2/AVX-512) in Go 1.27.
- Kernel file pattern:
  - `foo_simd.go`: `//go:build goexperiment.simd && (arm64 || amd64) && !purego`
    (split into `foo_simd_arm64.go` / `foo_simd_amd64.go` when the lane order
    differs per arch).
  - `foo_nosimd.go`: `//go:build !goexperiment.simd || purego || !(arm64 || amd64)`
    — routes to the pure-Go lane references.
- Every lane also has a pure-Go reference (`fooNeonRef`, `fooSse2Ref`, ...).
  It is the spec, the fallback when the experiment is off, and the oracle that
  the archsimd kernel is tested against bit-exactly. A build without
  `GOEXPERIMENT=simd` is still bit-exact; only slower.
- archsimd semantics: `a.Mul(b).Add(c)` compiles to separate multiply and add
  (unfused, like `vmlaq_f32`/`_mm_mul_ps`+`_mm_add_ps`); `a.MulAdd(b, c)` is a
  real FMA (like `vfmaq_f32`/`_mm256_fmadd_ps`). Use each exactly where the C++
  does.
- No hand-written assembly. Kernels are archsimd; a kernel archsimd cannot
  express stays pure Go.
- Every vector instruction runs behind a test of its CPU feature: a branch
  on a package-level variable set from `archsimd.X86`, in the function itself
  or in every caller. The first archsimd kernel brings goaec3's guard tests
  with it (`TestSimdInstructionsGated`, `TestNoLegacySSEWithDirtyUpperState`)
  and the `avx-only` CI job (the tests on an emulated CPU with AVX and no
  AVX2).
- 256-bit code needs AVX2, also where its instructions are AVX, and ends with
  `archsimd.ClearAVXUpperBits()`. The 128-bit kernels run with AVX alone;
  some 128-bit archsimd operations compile to AVX2 instructions (the
  broadcasts, `Abs`, `Neg`) and are replaced as in goaec3.
- CPU detection is one routine for builds with and without
  `GOEXPERIMENT=simd`, so both builds pick the same lane. The library never
  reads the environment; tests force a lane through an internal hook.
- The performance rules of goaec3's AGENTS.md (addressing, closures,
  transition penalties, what bounds a kernel on arm64 and amd64) apply to
  every kernel written here.

## Memory safety: no unsafe

- No non-test file imports `unsafe`. `TestNoUnsafe` in the root package scans
  every Go file of the module, whatever its build tags, and fails on the
  import. The only exception is `internal/webrtcoracle`, the cgo test oracle,
  which no shipped package imports.
- Unchecked access is obtained by proof, not by pointer arithmetic: convert a
  slice to an array pointer over a window with constant bounds
  (`(*[160]float32)(band)`), then index it with a constant, a masked value,
  or a loop counter with constant limits. Check with
  `go build -gcflags=-d=ssa/check_bce ./internal/...`.
- Reinterpreting bits goes through `math.Float32bits`/`Float32frombits` or
  archsimd's `ToBits`/`BitsToFloat32`, never through a pointer cast.

## Float discipline

- All signal-domain storage and arithmetic is `float32` (C `float`). Use
  `float64` only where the C++ uses `double`.
- Go may fuse `x*y + z` into an FMA on arm64, also across statements. libwebrtc
  is built with contraction off, so wrap every product that feeds an add/sub
  in `float32(...)`: `float32(a*b) + float32(c*d)`, including products stored
  in locals (`a := float32(x * y)` before `1 - a`) and compound assignments
  (`s += float32(a * b)`). Use a real FMA only where the C++ calls one.
- C++ evaluates `a * b * c` left to right and `float op int` in float; keep
  the order and the types. `std::min(a, b)` is `b < a ? b : a` and
  `std::max(a, b)` is `a < b ? b : a`: with a NaN operand the result depends
  on the order, so port them as written (`cmath.Min`, `cmath.Max`), not with
  Go's `min` and `max`.
- Transcendentals: `sqrtf` is exact via `float32(math.Sqrt(float64(x)))`.
  Every other libm call the C++ makes goes through a helper in
  `internal/cmath` that returns the rounded result of the float64 function
  (`Exp2f` for `powf(2.f, x)`, `Tanhf`, `Cosf`, `Sinf`, `Atanf`). The oracle binary
  interposes those same Go functions for the libm symbols
  (`internal/webrtcoracle/libm.go`, `libm_interpose.c`), because host libm
  float results are not correctly rounded and differ between Apple libm and
  glibc. Any new libm call needs the same treatment: a helper, an interposer,
  and its name in the import check of `src_main_test.go`.
- The `webrtc_oracle_stock` tag links the host libm instead of the
  interposer; the tests then log how far goapm is from libwebrtc on that
  libm instead of failing.

## Naming

- One Go file per C++ file (`noise_estimator.cc` → `noise_estimator.go`).
- C++ class → Go type of the same name; constructor `NewX`; methods keep their
  C++ names (exported). `kFoo` constants drop the `k` (`FftSizeBy2Plus1`).
  `std::array<float, N>` → `[N]float32`; `size_t`/`int` → `int`.
- Config structs keep the C++ snake_case field names in CamelCase.
- ApmDataDumper, metrics/histogram reporting and logging are not ported.
  Field trials are fixed at their defaults.

## Allocation discipline

Steady-state processing is allocation-free. Lock every hot path with a
`testing.AllocsPerRun(...) == 0` test after warm-up.

## Garbage collector footprint

A processor is long-lived state the garbage collector marks on every cycle.
What the mark costs: the heap objects, the pointer words it reads (every
pointer word of a live object, nil or not) and the length of an object up to
its last pointer word. A pointer-free object is marked without being read,
whatever its size.

- The state of a component is a pointer-free struct that its parent holds by
  value. Tables are package-level arrays, not per-instance slices.
- What depends on the number of channels is one pointer-free slice of
  per-channel structs, not a slice of pointers.
- Multi-dimensional buffers are one pointer-free array with views computed
  from the indices, not slices of slices.
- A component copies its config rather than pointing at it. An optional
  implementation is a tagged value, not an interface.
- Tests lock the shape: a `Test...PointerFree` test per state type, and a
  test of the number of heap objects of each constructor.
- The full rules, and the arena that holds the arrays of a whole pipeline in
  one object, are in goaec3's AGENTS.md ("Garbage collector footprint");
  they apply to the pipeline that chains the modules here.

## Process

- Work on a branch; open a pull request at every safe point (one module
  with its parity test). A pull request merges on green CI; one that changes
  performance also carries a run of `bench.yml` (Go against the source-built
  libwebrtc on real x86 and arm64 runners, base and head on the same
  machine).
- Timings on a shared Mac and under Rosetta do not predict x86: decide from
  the runner benchmarks.
- Keep the README short and plain, and in sync with what is merged.
- Untrusted input never panics: fuzz targets cover every public entry point
  and every parser of a frame shape.

## Build and test

- Use Go 1.27.4 (`$HOME/sdk/go1.27.4/bin/go`, or `go1.27.4`). The `Makefile`
  wraps every command below. CI (`.github/workflows/ci.yml`): `lint`
  (gofmt, `go vet` in every build configuration, generated files current,
  actionlint); `vuln` (govulncheck); `test` (default, simd, purego on
  linux/amd64 and linux/arm64); `test-32bit`; `race` (`go test -race -short
  ./...` without and with simd); `oracle-src` (the source oracle on amd64 and
  arm64, with and without simd and in the scalar lane; the simd jobs also
  log the divergence on glibc's libm). Runs on main, nightly and manual runs
  are never cancelled; a new push to a pull request cancels its run.
  `.github/workflows/macos.yml` runs separately (one macos-15 job: the tests
  in all three builds, the source oracle and the Apple libm divergence), so a
  macOS job waiting for a runner never holds up ci.
  `.github/workflows/bench.yml` times Go against the source-built libwebrtc
  (push to `perf/**` or manual). Actions are pinned to commits; workflows
  run with `contents: read`.
- `go test ./...` — default lane, pure-Go lane references (`make test`).
- `GOEXPERIMENT=simd go test ./...` — default lane with archsimd kernels
  (`make test-simd`).
- `go test -tags purego ./...` — scalar lane (`make test-purego`).
- `GOARCH=amd64 go test ./...` — amd64 under Rosetta (`make test-amd64`).
- `go test -race -short ./...`, with and without `GOEXPERIMENT=simd`
  (`make test-race`).
- `make lint` — gofmt, `go vet` in every build configuration, and the
  generated source-oracle files are current.
- `make oracle` — parity against the prebuilt libwebrtc (darwin/arm64 only;
  needs `tmp_check/libwebrtc`).
- `scripts/fetch-webrtc.sh` (`make fetch-webrtc`) — sparse checkout of the
  pinned WebRTC sources and the third_party subtrees they include
  (abseil-cpp, rnnoise, pffft) into `tmp_check/webrtc-m141`; idempotent.
- `make oracle-src` — parity against the C++ reference compiled from
  `tmp_check/webrtc-m141` on any OS/arch (`CC=clang CXX=clang++ go test -tags
  webrtc_src_oracle ./internal/webrtcoracle/`); `make oracle-src-scalar` adds
  `-tags purego`. Needs clang, and libc++ on Linux (`libc++-dev`,
  `libc++abi-dev`). The `src_gen_*` files are generated by
  `scripts/gen-src-oracle.sh` (one translation unit per WebRTC source);
  regenerate after adding a source.
- `make oracle-src-rosetta` — the x86_64 source build under Rosetta 2 on
  Apple silicon.
- `make oracle-stock`, `make oracle-src-stock` — the oracle on the host libm
  (logs, does not gate).
- `make bench` — the `Benchmark<Module>VsLibwebrtc` benchmarks of
  `internal/webrtcoracle`.

## Layout

- `internal/cmath` — the float transcendentals and `std::min`/`std::max`.
- `internal/webrtcoracle` — cgo oracle harness (tags `webrtc_oracle`,
  `webrtc_src_oracle`; `webrtc_oracle_stock` with either links the host
  libm instead of the interposer).
- One package per C++ directory as modules are ported
  (`modules/audio_processing/ns` → `internal/ns`).
- root package `goapm` — public API.
