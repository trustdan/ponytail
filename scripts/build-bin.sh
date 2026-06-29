#!/bin/sh
# Cross-compile the committed ponytail binaries. Run on a Version bump (after
# `go run ./cmd/ponytail gen`); commit the result. CI rebuilds and diffs bin/ to
# guard drift, so the flags below must stay reproducible: -trimpath strips local
# paths, CGO_ENABLED=0 avoids the host C toolchain, -s -w drops debug tables.
set -eu
cd "$(dirname "$0")/.."

# os/arch pairs → committed filename. Windows arm64 runs amd64 via emulation, so
# ship one Windows build. ponytail: add windows-arm64 if native demand appears.
targets="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64"

for t in $targets; do
  os=${t%/*}
  arch=${t#*/}
  out="bin/ponytail-$os-$arch"
  [ "$os" = windows ] && out="$out.exe"
  echo "building $out"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags="-s -w" -o "$out" ./cmd/ponytail
done

chmod +x bin/ponytail
echo "done"
