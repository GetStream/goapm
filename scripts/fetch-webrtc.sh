#!/usr/bin/env bash
# Fetches the pinned WebRTC m141 sources (and the abseil-cpp headers they
# include) that the from-source C++ oracle (build tag webrtc_src_oracle)
# compiles: modules/audio_processing (ns, agc2, the high-pass filter, the
# band splitting, AudioBuffer, AudioProcessingImpl) and what it includes.
#
#   scripts/fetch-webrtc.sh [dest]      (dest defaults to tmp_check/webrtc-m141)
#
# WebRTC is a blobless partial clone with a non-cone sparse checkout, so only
# the listed paths are downloaded. abseil-cpp comes from chromium's
# third_party repository at the revision WebRTC's DEPS pins, as do rnnoise
# (the weights of the AGC2 voice activity detector) and pffft (its FFT); only
# those subtrees are downloaded (gitiles archives: the repository's tree
# objects alone are over 600 MB) and each is verified against its git tree
# id. Rerunning is cheap and idempotent.
#
# The checkout is the one goaec3 uses (same commit, same paths), so
# tmp_check may be a symlink to goaec3's.
set -euo pipefail

WEBRTC_URL=${WEBRTC_URL:-https://webrtc.googlesource.com/src}
# branch-heads/7390 (m141).
WEBRTC_COMMIT=d2eaa5570fc9959f8dbde32912a16366b8ee75f4
THIRD_PARTY_URL=${THIRD_PARTY_URL:-https://chromium.googlesource.com/chromium/src/third_party}
# src/third_party in WebRTC's DEPS at WEBRTC_COMMIT.
THIRD_PARTY_COMMIT=75d337208257a84e4e2fafff284af52179174fc1
# The git tree ids of the subtrees at THIRD_PARTY_COMMIT.
ABSEIL_TREE=6548fbffe996d930f8bb11a6e91257374b444a46
RNNOISE_TREE=3d26a763561906606001c508753badfab582aa55
PFFFT_TREE=d4934c1704a1fb572c073a69fafc6fc01fc53387

root=$(cd "$(dirname "$0")/.." && pwd)
dest=${1:-$root/tmp_check/webrtc-m141}

# Paths of the WebRTC tree the oracles compile or include.
webrtc_paths=(
	/DEPS
	/api/
	/common_audio/
	/modules/audio_processing/
	/rtc_base/
	/system_wrappers/
	/experiments/
)

if [ ! -d "$dest/.git" ]; then
	mkdir -p "$dest"
	git -C "$dest" init -q
	git -C "$dest" remote add origin "$WEBRTC_URL"
	git -C "$dest" config extensions.partialClone origin
	git -C "$dest" config remote.origin.promisor true
	git -C "$dest" config remote.origin.partialclonefilter blob:none
fi
printf '%s\n' "${webrtc_paths[@]}" | git -C "$dest" sparse-checkout set --no-cone --stdin
if [ "$(git -C "$dest" rev-parse -q --verify HEAD 2>/dev/null || true)" != "$WEBRTC_COMMIT" ]; then
	if ! git -C "$dest" cat-file -e "$WEBRTC_COMMIT^{commit}" 2>/dev/null; then
		git -C "$dest" fetch -q --depth 1 --filter=blob:none origin "$WEBRTC_COMMIT"
	fi
	git -C "$dest" -c advice.detachedHead=false checkout -q --detach "$WEBRTC_COMMIT"
fi
if [ "$(git -C "$dest" rev-parse HEAD)" != "$WEBRTC_COMMIT" ]; then
	echo "fetch-webrtc: $dest is not at $WEBRTC_COMMIT" >&2
	exit 1
fi
if [ -n "$(git -C "$dest" status --porcelain --untracked-files=no)" ]; then
	echo "fetch-webrtc: $dest has local modifications" >&2
	exit 1
fi

deps_rev=$(awk "/'src\/third_party':/ { getline; print; exit }" "$dest/DEPS" |
	sed -n 's/.*@\([0-9a-f]\{40\}\).*/\1/p')
if [ "$deps_rev" != "$THIRD_PARTY_COMMIT" ]; then
	echo "fetch-webrtc: DEPS pins src/third_party at '$deps_rev', the script at $THIRD_PARTY_COMMIT" >&2
	exit 1
fi

# tree_id DIR prints the git tree id of the files under DIR.
tree_id() {
	local tmp
	tmp=$(mktemp -d)
	GIT_DIR=$tmp GIT_WORK_TREE=$1 git init -q
	GIT_DIR=$tmp GIT_WORK_TREE=$1 git add -A -f
	GIT_DIR=$tmp GIT_WORK_TREE=$1 git write-tree
	rm -rf "$tmp"
}

# fetch_third_party NAME TREE downloads third_party/NAME unless it is there
# with tree id TREE.
fetch_third_party() {
	local name=$1 tree=$2
	local dir=$dest/third_party/$name
	local stamp=$dir.tree
	if [ -d "$dir" ] && [ -f "$stamp" ] && [ "$(cat "$stamp")" = "$tree" ]; then
		return
	fi
	rm -rf "$dir" "$dir.tmp" "$stamp"
	mkdir -p "$dir.tmp"
	curl -fsSL --retry 3 "$THIRD_PARTY_URL/+archive/$THIRD_PARTY_COMMIT/$name.tar.gz" |
		tar -xzf - -C "$dir.tmp"
	local got
	got=$(tree_id "$dir.tmp")
	if [ "$got" != "$tree" ]; then
		rm -rf "$dir.tmp"
		echo "fetch-webrtc: $name tree is $got, want $tree" >&2
		exit 1
	fi
	mv "$dir.tmp" "$dir"
	echo "$tree" >"$stamp"
}

fetch_third_party abseil-cpp "$ABSEIL_TREE"
fetch_third_party rnnoise "$RNNOISE_TREE"
fetch_third_party pffft "$PFFFT_TREE"

echo "fetch-webrtc: $dest at $WEBRTC_COMMIT, third_party at $THIRD_PARTY_COMMIT"
