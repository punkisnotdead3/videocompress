//go:build darwin && amd64

package main

import (
	"embed"
	"strings"
)

// assetsFS holds the macOS amd64 (Intel) ffmpeg static binaries.
// Run `make setup-darwin-amd64` to download before building.
//
//go:embed assets/ffmpeg-darwin-amd64 assets/ffprobe-darwin-amd64
var assetsFS embed.FS

const (
	ffmpegAsset  = "assets/ffmpeg-darwin-amd64"
	ffprobeAsset = "assets/ffprobe-darwin-amd64"
)

func cleanPath(p string) string { return strings.TrimPrefix(p, "/") }
