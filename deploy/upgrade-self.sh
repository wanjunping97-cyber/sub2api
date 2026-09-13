#!/usr/bin/env bash
# Replace a systemd install of this fork with a prebuilt linux package.
# Do not run official `sub2api upgrade` / install.sh upgrade — that downloads
# Wei-Shaw/sub2api and wipes the self-hosted changes.
set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/sub2api}"
SERVICE_NAME="${SERVICE_NAME:-sub2api}"
VERSION="${VERSION:-0.2.4-self.2}"
REPO="${REPO:-wanjunping97-cyber/sub2api}"
REF="${REF:-cursor/self-package-current-fec0}"

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) pkg_arch=amd64 ;;
  aarch64|arm64) pkg_arch=arm64 ;;
  *)
    echo "unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

archive="sub2api_${VERSION}_linux_${pkg_arch}.tar.gz"
download_url="${DOWNLOAD_URL:-https://github.com/${REPO}/raw/${REF}/self-packages/${archive}}"
checksum_url="${CHECKSUM_URL:-https://github.com/${REPO}/raw/${REF}/self-packages/checksums.txt}"

if [ "$(id -u)" -ne 0 ]; then
  echo "run as root: sudo $0" >&2
  exit 1
fi

if [ ! -x "$INSTALL_DIR/sub2api" ]; then
  echo "did not find $INSTALL_DIR/sub2api; set INSTALL_DIR if this is not a default install" >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "downloading $download_url"
curl -fL "$download_url" -o "$tmpdir/$archive"
if curl -fsL "$checksum_url" -o "$tmpdir/checksums.txt"; then
  expected="$(grep -E "[[:space:]]${archive}\$" "$tmpdir/checksums.txt" | awk '{print $1}')"
  actual="$(sha256sum "$tmpdir/$archive" | awk '{print $1}')"
  if [ -n "$expected" ] && [ "$expected" != "$actual" ]; then
    echo "checksum mismatch for $archive" >&2
    echo "expected: $expected" >&2
    echo "actual:   $actual" >&2
    exit 1
  fi
fi

tar -xzf "$tmpdir/$archive" -C "$tmpdir"
if [ ! -x "$tmpdir/sub2api" ]; then
  echo "archive is missing the sub2api binary" >&2
  exit 1
fi

stamp="$(date +%Y%m%d%H%M%S)"
cp -a "$INSTALL_DIR/sub2api" "$INSTALL_DIR/sub2api.bak.$stamp"
systemctl stop "$SERVICE_NAME"
install -m 0755 "$tmpdir/sub2api" "$INSTALL_DIR/sub2api"
systemctl start "$SERVICE_NAME"
systemctl --no-pager --full status "$SERVICE_NAME" || true
echo "updated $INSTALL_DIR/sub2api to $VERSION (backup: $INSTALL_DIR/sub2api.bak.$stamp)"
"$INSTALL_DIR/sub2api" --version || true
