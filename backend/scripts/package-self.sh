#!/usr/bin/env bash
# Build linux amd64/arm64 release tarballs with the embedded admin UI.
# Intended for a machine with enough RAM (roughly 4G+ free). Do not run this
# on a small production VPS.
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
BACKEND_DIR="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"
REPO_DIR="$(CDPATH= cd -- "$BACKEND_DIR/.." && pwd)"
OUT_DIR="${OUT_DIR:-$REPO_DIR/self-packages}"
VERSION="${VERSION:-0.2.4-self.1}"
COMMIT="${COMMIT:-$(git -C "$REPO_DIR" rev-parse --short HEAD)}"
DATE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
ARCHES="${ARCHES:-amd64 arm64}"

export NODE_OPTIONS="${NODE_OPTIONS:---max-old-space-size=3072}"
export GOMAXPROCS="${GOMAXPROCS:-2}"
export CGO_ENABLED=0

mkdir -p "$OUT_DIR"
WORKDIR="$(mktemp -d)"
cleanup() { rm -rf "$WORKDIR"; }
trap cleanup EXIT

echo "==> frontend build (embedded admin UI)"
if [ ! -d "$REPO_DIR/frontend/node_modules" ]; then
  (cd "$REPO_DIR/frontend" && pnpm install --frozen-lockfile)
fi
(cd "$REPO_DIR/frontend" && pnpm run build)

if [ ! -f "$BACKEND_DIR/internal/web/dist/index.html" ]; then
  echo "frontend build did not produce backend/internal/web/dist/index.html" >&2
  exit 1
fi

LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.Date=${DATE} -X main.BuildType=release"

build_one() {
  local arch="$1"
  local stage="$WORKDIR/linux_${arch}"
  mkdir -p "$stage/deploy"
  echo "==> go build linux/${arch} -tags embed"
  (
    cd "$BACKEND_DIR"
    GOOS=linux GOARCH="$arch" go build -trimpath -tags embed \
      -ldflags "$LDFLAGS" \
      -o "$stage/sub2api" \
      ./cmd/server
  )
  chmod +x "$stage/sub2api"
  cp "$REPO_DIR/LICENSE" "$stage/LICENSE"
  cp "$REPO_DIR/README.md" "$stage/README.md"
  cp -R "$REPO_DIR/deploy/." "$stage/deploy/"
  local archive="sub2api_${VERSION}_linux_${arch}.tar.gz"
  tar -C "$stage" -czf "$OUT_DIR/$archive" sub2api LICENSE README.md deploy
  echo "wrote $OUT_DIR/$archive ($(du -h "$OUT_DIR/$archive" | awk '{print $1}'))"
}

for arch in $ARCHES; do
  build_one "$arch"
done

(
  cd "$OUT_DIR"
  sha256sum sub2api_${VERSION}_linux_*.tar.gz > checksums.txt
)

echo "==> packages"
ls -lh "$OUT_DIR"/sub2api_${VERSION}_linux_*.tar.gz "$OUT_DIR/checksums.txt"
echo "version=${VERSION} commit=${COMMIT}"
