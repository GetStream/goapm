.PHONY: test test-purego test-simd test-amd64 test-all test-race oracle oracle-src oracle-src-scalar
.PHONY: oracle-stock oracle-src-stock oracle-src-rosetta fetch-webrtc gen-src-oracle lint bench vuln

GO ?= go
# The C++ reference builds need clang: libwebrtc is built by clang with
# contraction off, which the sources request through a clang pragma.
CLANG_CC ?= clang
CLANG_CXX ?= clang++
ORACLE_TIMEOUT ?= 40m
SIMD := GOEXPERIMENT=simd
SRC_TAGS := webrtc_src_oracle
ROSETTA_ENV := CGO_ENABLED=1 GOARCH=amd64 CC="$(CLANG_CC) -arch x86_64" CXX="$(CLANG_CXX) -arch x86_64"

# Default lane, pure-Go lane references (no GOEXPERIMENT).
test:
	$(GO) test ./...

# Scalar lane.
test-purego:
	$(GO) test -tags purego ./...

# Default lane with the archsimd kernels.
test-simd:
	$(SIMD) $(GO) test ./...

# amd64 lanes (under Rosetta 2 on Apple silicon).
test-amd64:
	GOARCH=amd64 $(GO) test ./...
	GOARCH=amd64 $(SIMD) $(GO) test ./...

test-all: test test-purego test-simd test-amd64

# The race detector over the short test suite, without and with archsimd.
test-race:
	$(GO) test -race -short ./...
	$(SIMD) $(GO) test -race -short ./...

# Known vulnerabilities in the code the packages and their tests reach,
# without and with archsimd.
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@v1.10.0
vuln:
	$(GO) run $(GOVULNCHECK) -test ./...
	$(SIMD) $(GO) run $(GOVULNCHECK) -test ./...

# Sparse checkout of the pinned WebRTC m141 sources and the third_party
# subtrees they include into tmp_check/webrtc-m141.
fetch-webrtc:
	bash scripts/fetch-webrtc.sh

# Regenerate the src_gen_* files of the webrtc_src_oracle build.
gen-src-oracle:
	bash scripts/gen-src-oracle.sh

# Parity against the prebuilt libwebrtc.a (darwin/arm64, needs
# tmp_check/libwebrtc), with and without the archsimd kernels and in the
# scalar lane.
oracle:
	$(GO) test -timeout $(ORACLE_TIMEOUT) -tags webrtc_oracle ./internal/webrtcoracle/
	$(SIMD) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags webrtc_oracle ./internal/webrtcoracle/
	$(GO) test -timeout $(ORACLE_TIMEOUT) -tags 'webrtc_oracle purego' ./internal/webrtcoracle/

# The prebuilt libwebrtc on Apple libm, without the interposer that
# substitutes the Go transcendentals: the tests log what differs.
oracle-stock:
	$(SIMD) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags 'webrtc_oracle webrtc_oracle_stock' -v ./internal/webrtcoracle/

# The same against the source build on the host libm.
oracle-src-stock: fetch-webrtc
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(SIMD) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags '$(SRC_TAGS) webrtc_oracle_stock' -v ./internal/webrtcoracle/

# Parity against the C++ reference built from the WebRTC sources, on this
# machine's architecture, with and without archsimd.
oracle-src: fetch-webrtc
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags $(SRC_TAGS) ./internal/webrtcoracle/
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(SIMD) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags $(SRC_TAGS) ./internal/webrtcoracle/

# The scalar lane of the source build.
oracle-src-scalar: fetch-webrtc
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags '$(SRC_TAGS) purego' ./internal/webrtcoracle/

# The x86_64 source build under Rosetta 2 on Apple silicon.
oracle-src-rosetta: fetch-webrtc
	$(ROSETTA_ENV) $(SIMD) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags $(SRC_TAGS) ./internal/webrtcoracle/
	$(ROSETTA_ENV) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags $(SRC_TAGS) ./internal/webrtcoracle/
	$(ROSETTA_ENV) $(GO) test -timeout $(ORACLE_TIMEOUT) -tags '$(SRC_TAGS) purego' ./internal/webrtcoracle/

lint:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi
	$(GO) vet ./...
	$(SIMD) $(GO) vet ./...
	$(GO) vet -tags purego ./...
	GOARCH=arm64 $(SIMD) $(GO) vet ./...
	GOARCH=amd64 $(SIMD) $(GO) vet ./...
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(GO) vet -tags $(SRC_TAGS) ./...
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(SIMD) $(GO) vet -tags '$(SRC_TAGS) purego' ./...
	CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(SIMD) $(GO) vet -tags '$(SRC_TAGS) webrtc_oracle_stock' ./...
	@if [ -d tmp_check/webrtc-m141/modules ]; then bash scripts/gen-src-oracle.sh -check; else echo "lint: tmp_check/webrtc-m141 missing, gen-src-oracle -check skipped"; fi

# Go against libwebrtc (ns per 10 ms frame); the prebuilt library on
# darwin/arm64, the source build elsewhere.
bench:
	@if [ -f tmp_check/libwebrtc/lib/libwebrtc.a ]; then \
		$(SIMD) $(GO) test -tags webrtc_oracle -run '^$$' -bench VsLibwebrtc -benchtime 2000x -count 5 ./internal/webrtcoracle/; \
	else \
		CC=$(CLANG_CC) CXX=$(CLANG_CXX) $(SIMD) $(GO) test -tags $(SRC_TAGS) -run '^$$' -bench VsLibwebrtc -benchtime 2000x -count 5 ./internal/webrtcoracle/; \
	fi
