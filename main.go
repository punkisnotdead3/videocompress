package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	currentVideoPath string
	ffmpegBinPath    string
	ffprobeBinPath   string
)

func main() {
	// Initialize app first so we can show errors in GUI
	a := app.NewWithID("com.videocompress.app")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Video Compressor")
	w.Resize(fyne.NewSize(860, 680))
	w.SetFixedSize(false)

	// Extract embedded ffmpeg binaries
	var err error
	ffmpegBinPath, ffprobeBinPath, err = extractFFmpeg()
	if err != nil {
		w.SetContent(container.NewCenter(
			widget.NewLabel("初始化失败: " + err.Error()),
		))
		w.ShowAndRun()
		return
	}
	defer cleanupFFmpeg()

	// ── UI State ──────────────────────────────────────────────────────────────

	// File section
	selectedFileLabel := widget.NewLabel("未选择文件")
	selectedFileLabel.Wrapping = fyne.TextWrapWord

	// Info grid (key-value pairs)
	infoGrid := container.New(layout.NewFormLayout())
	infoCard := widget.NewCard("视频信息", "", infoGrid)
	infoCard.Hide()

	// Compression section
	bitrateEntry := widget.NewEntry()
	bitrateEntry.SetPlaceHolder("目标视频码率 (kbps)，如: 1000")

	presetSelect := widget.NewSelect(
		[]string{"ultrafast", "superfast", "veryfast", "faster", "fast", "medium", "slow", "slower", "veryslow"},
		nil,
	)
	presetSelect.SetSelected("medium")

	audioSelect := widget.NewSelect(
		[]string{"复制原始音频", "重新编码 128k", "重新编码 192k", "重新编码 256k"},
		nil,
	)
	audioSelect.SetSelected("复制原始音频")

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	logEntry := widget.NewMultiLineEntry()
	logEntry.SetMinRowsVisible(6)
	logEntry.Disable()
	logScroll := container.NewVScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(0, 130))

	compressBtn := widget.NewButton("开始压缩", nil)
	compressBtn.Importance = widget.HighImportance
	compressBtn.Disable()

	selectBtn := widget.NewButton("选择视频文件", nil)
	selectBtn.Importance = widget.MediumImportance

	// ── Button Handlers ───────────────────────────────────────────────────────

	// loadVideo is shared by both the file picker and the drag-and-drop handler.
	loadVideo := func(path string) {
		currentVideoPath = path
		selectedFileLabel.SetText(filepath.Base(path))
		statusLabel.SetText("正在分析视频...")
		infoCard.Hide()
		bitrateEntry.SetText("")
		compressBtn.Disable()
		logEntry.SetText("")

		go func() {
			info, err := getVideoInfo(ffprobeBinPath, path)
			if err != nil {
				statusLabel.SetText("分析失败: " + err.Error())
				return
			}
			updateInfoCard(infoGrid, infoCard, info)
			// Suggest 50% of current video bitrate as default target
			if info.VideoBitrate > 0 {
				suggested := info.VideoBitrate / 2 / 1000
				if suggested < 200 {
					suggested = 200
				}
				bitrateEntry.SetText(strconv.FormatInt(suggested, 10))
			}
			compressBtn.Enable()
			statusLabel.SetText("分析完成，可以开始压缩")
		}()
	}

	selectBtn.OnTapped = func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			reader.Close()
			loadVideo(cleanPath(reader.URI().Path()))
		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{
			".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".ts", ".mts",
		}))
		fd.Show()
	}

	compressBtn.OnTapped = func() {
		if currentVideoPath == "" {
			dialog.ShowInformation("提示", "请先选择视频文件", w)
			return
		}

		targetStr := strings.TrimSpace(bitrateEntry.Text)
		if targetStr == "" {
			dialog.ShowInformation("提示", "请输入目标码率", w)
			return
		}
		targetKbps, err := strconv.ParseInt(targetStr, 10, 64)
		if err != nil || targetKbps <= 0 {
			dialog.ShowInformation("提示", "请输入有效的正整数码率 (kbps)", w)
			return
		}

		outputPath := buildOutputPath(currentVideoPath)
		preset := presetSelect.Selected

		var audioBitrate int64 = 0 // 0 = copy
		switch audioSelect.Selected {
		case "重新编码 128k":
			audioBitrate = 128
		case "重新编码 192k":
			audioBitrate = 192
		case "重新编码 256k":
			audioBitrate = 256
		}

		progressBar.SetValue(0)
		progressBar.Show()
		statusLabel.SetText("压缩中...")
		compressBtn.Disable()
		selectBtn.Disable()
		logEntry.SetText("")

		go func() {
			compressErr := compressVideo(
				ffmpegBinPath, currentVideoPath, outputPath,
				targetKbps, preset, audioBitrate,
				func(progress float64, logLine string) {
					if progress >= 0 {
						progressBar.SetValue(progress)
					}
					if logLine != "" {
						current := logEntry.Text
						if len(current) > 8000 {
							// Keep last 6000 chars to avoid huge string
							current = current[len(current)-6000:]
						}
						logEntry.SetText(current + logLine + "\n")
						logScroll.ScrollToBottom()
					}
				},
			)

			progressBar.SetValue(1.0)
			compressBtn.Enable()
			selectBtn.Enable()

			if compressErr != nil {
				statusLabel.SetText("压缩失败: " + compressErr.Error())
				dialog.ShowError(compressErr, w)
			} else {
				statusLabel.SetText("压缩完成！输出: " + filepath.Base(outputPath))
				dialog.ShowInformation("完成", fmt.Sprintf(
					"视频压缩成功！\n\n输出文件:\n%s", outputPath,
				), w)
			}
		}()
	}

	// ── Layout ────────────────────────────────────────────────────────────────

	fileSection := widget.NewCard("选择视频", "支持从文件管理器拖拽视频文件到此窗口",
		container.NewVBox(
			selectedFileLabel,
			selectBtn,
		),
	)

	compressionForm := container.New(layout.NewFormLayout(),
		widget.NewLabel("目标视频码率 (kbps)"), bitrateEntry,
		widget.NewLabel("编码速度预设"), presetSelect,
		widget.NewLabel("音频处理"), audioSelect,
	)

	compressionSection := widget.NewCard("压缩设置", "",
		container.NewVBox(
			compressionForm,
			compressBtn,
		),
	)

	progressSection := container.NewVBox(
		progressBar,
		statusLabel,
		widget.NewLabel("FFmpeg 输出日志:"),
		logScroll,
	)

	content := container.NewVBox(
		fileSection,
		infoCard,
		compressionSection,
		progressSection,
	)

	w.SetContent(container.NewPadded(container.NewVScroll(content)))

	// Accept video files dragged from the OS file manager onto the window.
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}
		path := cleanPath(uris[0].Path())
		switch strings.ToLower(filepath.Ext(path)) {
		case ".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".ts", ".mts":
			loadVideo(path)
		default:
			statusLabel.SetText("不支持的文件格式: " + filepath.Ext(path))
		}
	})

	w.ShowAndRun()
}

// updateInfoCard fills the info grid and shows the card on the main goroutine.
func updateInfoCard(grid *fyne.Container, card *widget.Card, info *VideoInfo) {
	grid.Objects = nil

	addRow := func(label, value string) {
		lbl := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		val := widget.NewLabel(value)
		val.Wrapping = fyne.TextWrapWord
		grid.Add(lbl)
		grid.Add(val)
	}

	// addRedRow displays the value in error/red color to draw attention to bitrate info.
	addRedRow := func(label, value string) {
		lbl := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		val := widget.NewRichText(&widget.TextSegment{
			Style: widget.RichTextStyle{ColorName: theme.ColorNameError},
			Text:  value,
		})
		grid.Add(lbl)
		grid.Add(val)
	}

	addRow("格式", info.FormatName)
	addRow("时长", formatDuration(info.Duration))
	addRow("文件大小", formatSize(info.FileSize))
	addRedRow("总码率", fmt.Sprintf("%.2f Mbps", float64(info.TotalBitrate)/1_000_000))
	addRow("分辨率", fmt.Sprintf("%d × %d", info.Width, info.Height))
	addRow("视频编码", info.VideoCodec)
	addRedRow("视频码率", fmt.Sprintf("%.2f Mbps", float64(info.VideoBitrate)/1_000_000))
	addRow("帧率", fmt.Sprintf("%.3g fps", info.FPS))
	addRow("像素格式", info.PixFmt)
	if info.AudioCodec != "" {
		addRow("音频编码", info.AudioCodec)
		addRedRow("音频码率", fmt.Sprintf("%.2f Mbps", float64(info.AudioBitrate)/1_000_000))
		addRow("声道数", strconv.Itoa(info.AudioChannels))
		addRow("采样率", fmt.Sprintf("%d Hz", info.SampleRate))
	}

	grid.Refresh()
	card.Show()
}

// buildOutputPath derives a "_compressed" output path next to the input file.
func buildOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	return filepath.Join(dir, base+"_compressed"+ext)
}
