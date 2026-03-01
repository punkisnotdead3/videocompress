//go:build windows

package main

import (
	"embed"
	"strings"
)

// assetsFS holds the Windows amd64 ffmpeg static binaries.
// Run `.\setup.ps1` to download before building.
//
//go:embed assets/ffmpeg-windows.exe assets/ffprobe-windows.exe
var assetsFS embed.FS

const (
	ffmpegAsset  = "assets/ffmpeg-windows.exe"
	ffprobeAsset = "assets/ffprobe-windows.exe"
)

// cleanPath converts fyne's "/C:/Users/..." URI path to a Windows path.
func cleanPath(p string) string {
	p = strings.TrimPrefix(p, "/")
	return strings.ReplaceAll(p, "/", "\\")
}
