#!/usr/bin/env bash
# setup-darwin.sh — Download static ffmpeg/ffprobe for macOS (arm64 by default)
# Usage:  bash setup-darwin.sh [arm64|amd64]

set -euo pipefail

ARCH="${1:-arm64}"
VER="7.1"
ASSETS="assets"

mkdir -p "$ASSETS"

download_and_extract() {
  local tool="$1"      # ffmpeg or ffprobe
  local url="https://evermeet.cx/pub/${tool}/${tool}-${VER}.zip"
  local tmp="/tmp/vc-${tool}-darwin.zip"
  local tmpdir="/tmp/vc-${tool}-darwin"

  echo ">>> Downloading ${tool} ${VER} for macOS..."
  curl -fsSL "$url" -o "$tmp"
  mkdir -p "$tmpdir"
  unzip -o "$tmp" "$tool" -d "$tmpdir"
  cp "${tmpdir}/${tool}" "${ASSETS}/${tool}-darwin"
  chmod +x "${ASSETS}/${tool}-darwin"
  echo "    Saved: ${ASSETS}/${tool}-darwin ($(du -sh "${ASSETS}/${tool}-darwin" | cut -f1))"
}

download_and_extract ffmpeg
download_and_extract ffprobe

echo ""
echo "✓ macOS ffmpeg binaries ready in ${ASSETS}/"
ls -lh "${ASSETS}/ffmpeg-darwin" "${ASSETS}/ffprobe-darwin"
echo ""
echo "Next: go build -ldflags='-s -w' -o dist/videocompress-darwin ."
