# Security

## Supported versions

goapm is pre-v1 and in development. Fixes go to `main` only.

## Reporting a vulnerability

Report it privately: on GitHub, open the repository's **Security** tab and
choose **Report a vulnerability**. Please don't open a public issue. Include
a config and input that reproduce the problem if you can.

## What goapm defends against

- **Untrusted audio is fine.** No frame makes a call panic. Every `float32`
  value is accepted.
- **NaN is not an attack, but it is permanent.** As in libwebrtc, a NaN
  sample can make every later output NaN. Replace NaN before the call, or
  create a new processor.
- **Memory is bounded per processor.** Nothing grows with the audio.
- **No `unsafe`, no cgo, no assembly** in the library, enforced in CI.

## What it does not defend against

- **Concurrent calls on one processor.** They are a data race. Nothing
  detects it, and Go's memory safety does not hold under a data race.
- **Timing.** Processing time depends on the audio and the config. goapm is
  not constant time and makes no attempt to be.

## How this is checked

Unit tests drive every module in every build, and parity tests compare each
with libwebrtc bit for bit. CI runs the race detector and `govulncheck`;
actions are pinned to commits.
