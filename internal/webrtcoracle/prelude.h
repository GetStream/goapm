// Prelude of every translation unit of the oracle.
#ifndef GOAPM_PRELUDE_H
#define GOAPM_PRELUDE_H

// libwebrtc is built by clang with -ffp-contract=off. gcc ignores the pragma
// below and contracts by default, so the build insists on clang.
#if !defined(__clang__)
#error "the oracle needs clang (CC=clang CXX=clang++)"
#endif
#if !defined(NDEBUG)
#error "the oracle is a release (NDEBUG) build, like libwebrtc"
#endif

#pragma clang fp contract(off)
#pragma clang diagnostic ignored "-Wnullability-completeness"

#endif  // GOAPM_PRELUDE_H
