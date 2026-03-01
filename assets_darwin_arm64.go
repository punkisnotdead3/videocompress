//go:build darwin && arm64

package main

import (
	"embed"
	"strings"
)

// assetsFS holds the macOS arm64 (Apple Silicon) ffmpeg static binaries.
// Run `make setup-darwin-arm64` to download before building.
//
//go:embed assets/ffmpeg-darwin-arm64 assets/ffprobe-darwin-arm64
var assetsFS embed.FS

const (
	ffmpegAsset  = "assets/ffmpeg-darwin-arm64"
	ffprobeAsset = "assets/ffprobe-darwin-arm64"
)

func cleanPath(p string) string { return strings.TrimPrefix(p, "/") }
