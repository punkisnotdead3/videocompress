package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// ── Data types ────────────────────────────────────────────────────────────────

// VideoInfo holds parsed video metadata.
type VideoInfo struct {
	FormatName    string
	Duration      float64
	FileSize      int64
	TotalBitrate  int64
	Width         int
	Height        int
	VideoCodec    string
	VideoBitrate  int64
	FPS           float64
	PixFmt        string
	AudioCodec    string
	AudioBitrate  int64
	AudioChannels int
	SampleRate    int
}

// ffprobe JSON structures
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecName  string `json:"codec_name"`
	CodecType  string `json:"codec_type"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	RFrameRate string `json:"r_frame_rate"`
	BitRate    string `json:"bit_rate"`
	PixFmt     string `json:"pix_fmt"`
	Channels   int    `json:"channels"`
	SampleRate string `json:"sample_rate"`
}

type ffprobeFormat struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
}

// ── FFmpeg lifecycle ──────────────────────────────────────────────────────────

var tempDir string

// extractFFmpeg writes the embedded ffmpeg/ffprobe binaries to a temp directory
// and returns their paths. Call cleanupFFmpeg() on exit.
func extractFFmpeg() (ffmpegPath, ffprobePath string, err error) {
	tempDir, err = os.MkdirTemp("", "videocompress-*")
	if err != nil {
		return "", "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// ffmpegAsset / ffprobeAsset are declared in the platform+arch embed file
	// (assets_darwin_arm64.go / assets_darwin_amd64.go / assets_windows.go).
	switch runtime.GOOS {
	case "darwin":
		ffmpegPath = filepath.Join(tempDir, "ffmpeg")
		ffprobePath = filepath.Join(tempDir, "ffprobe")
	case "windows":
		ffmpegPath = filepath.Join(tempDir, "ffmpeg.exe")
		ffprobePath = filepath.Join(tempDir, "ffprobe.exe")
	default:
		return "", "", fmt.Errorf("不支持的平台: %s", runtime.GOOS)
	}

	if err = writeAsset(ffmpegAsset, ffmpegPath); err != nil {
		return "", "", fmt.Errorf("释放 ffmpeg 失败: %w", err)
	}
	if err = writeAsset(ffprobeAsset, ffprobePath); err != nil {
		return "", "", fmt.Errorf("释放 ffprobe 失败: %w", err)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(ffmpegPath, 0755)
		_ = os.Chmod(ffprobePath, 0755)
	}
	return ffmpegPath, ffprobePath, nil
}

func writeAsset(assetName, destPath string) error {
	data, err := assetsFS.ReadFile(assetName)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0755)
}

func cleanupFFmpeg() {
	if tempDir != "" {
		_ = os.RemoveAll(tempDir)
	}
}

// ── Video analysis ────────────────────────────────────────────────────────────

// getVideoInfo runs ffprobe and returns parsed VideoInfo.
func getVideoInfo(ffprobePath, videoPath string) (*VideoInfo, error) {
	cmd := exec.Command(ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe 执行失败: %w", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, fmt.Errorf("解析 ffprobe 输出失败: %w", err)
	}

	info := &VideoInfo{
		FormatName: probe.Format.FormatName,
	}

	if d, e := strconv.ParseFloat(probe.Format.Duration, 64); e == nil {
		info.Duration = d
	}
	if s, e := strconv.ParseInt(probe.Format.Size, 10, 64); e == nil {
		info.FileSize = s
	}
	if b, e := strconv.ParseInt(probe.Format.BitRate, 10, 64); e == nil {
		info.TotalBitrate = b
	}

	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			info.VideoCodec = s.CodecName
			info.Width = s.Width
			info.Height = s.Height
			info.PixFmt = s.PixFmt
			info.FPS = parseFrameRate(s.RFrameRate)
			if b, e := strconv.ParseInt(s.BitRate, 10, 64); e == nil {
				info.VideoBitrate = b
			}
		case "audio":
			info.AudioCodec = s.CodecName
			info.AudioChannels = s.Channels
			if b, e := strconv.ParseInt(s.BitRate, 10, 64); e == nil {
				info.AudioBitrate = b
			}
			if r, e := strconv.Atoi(s.SampleRate); e == nil {
				info.SampleRate = r
			}
		}
	}

	// Estimate video bitrate from total if stream-level info missing
	if info.VideoBitrate == 0 && info.TotalBitrate > 0 {
		info.VideoBitrate = info.TotalBitrate - info.AudioBitrate
		if info.VideoBitrate < 0 {
			info.VideoBitrate = info.TotalBitrate
		}
	}

	return info, nil
}

func parseFrameRate(s string) float64 {
	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		num, e1 := strconv.ParseFloat(parts[0], 64)
		den, e2 := strconv.ParseFloat(parts[1], 64)
		if e1 == nil && e2 == nil && den != 0 {
			return math.Round(num/den*1000) / 1000
		}
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// ── Video compression ─────────────────────────────────────────────────────────

// progressFn receives (progress 0-1, logLine).
// progress < 0 means no progress update, just a log line.
type progressFn func(progress float64, logLine string)

// compressVideo transcodes inputPath → outputPath with the given bitrate settings.
// It parses ffmpeg stderr to report progress via progressFn.
func compressVideo(
	ffmpegPath, inputPath, outputPath string,
	targetKbps int64,
	preset string,
	audioBitrateKbps int64, // 0 = copy audio
	progress progressFn,
) error {
	maxRate := targetKbps * 3 / 2
	bufSize := targetKbps * 2

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-b:v", fmt.Sprintf("%dk", targetKbps),
		"-maxrate", fmt.Sprintf("%dk", maxRate),
		"-bufsize", fmt.Sprintf("%dk", bufSize),
		"-preset", preset,
	}

	if audioBitrateKbps == 0 {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "aac",
			"-b:a", fmt.Sprintf("%dk", audioBitrateKbps))
	}

	args = append(args, "-y", outputPath)

	cmd := exec.Command(ffmpegPath, args...)

	// FFmpeg writes progress info to stderr
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var totalDuration float64

	// Patterns to find in stderr
	durationRe := regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+\.?\d*)`)
	timeRe := regexp.MustCompile(`time=(\d+):(\d+):(\d+\.?\d*)`)

	scanner := bufio.NewScanner(stderr)
	// FFmpeg writes progress on the same line with \r; split on either
	scanner.Split(scanLinesOrCR)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Grab duration on first occurrence
		if totalDuration == 0 {
			if m := durationRe.FindStringSubmatch(line); len(m) == 4 {
				h, _ := strconv.ParseFloat(m[1], 64)
				min, _ := strconv.ParseFloat(m[2], 64)
				sec, _ := strconv.ParseFloat(m[3], 64)
				totalDuration = h*3600 + min*60 + sec
			}
		}

		// Parse current time for progress percentage
		if totalDuration > 0 {
			if m := timeRe.FindStringSubmatch(line); len(m) == 4 {
				h, _ := strconv.ParseFloat(m[1], 64)
				min, _ := strconv.ParseFloat(m[2], 64)
				sec, _ := strconv.ParseFloat(m[3], 64)
				cur := h*3600 + min*60 + sec
				pct := cur / totalDuration
				if pct > 1 {
					pct = 1
				}
				progress(pct, line)
				continue
			}
		}

		progress(-1, line)
	}

	return cmd.Wait()
}

// scanLinesOrCR is a bufio.SplitFunc that splits on \n or \r.
func scanLinesOrCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, b := range data {
		if b == '\n' || b == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// ── Formatting helpers ────────────────────────────────────────────────────────

func formatDuration(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - math.Floor(seconds)) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
