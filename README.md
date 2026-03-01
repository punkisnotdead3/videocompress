# Video Compressor / 视频压缩工具

[English](#english) | [中文](#中文)

---

## English

A cross-platform desktop video compressor with a clean dark-theme GUI.
Powered by **Go + Fyne** and ships with **embedded FFmpeg** — no external dependencies needed.

### Features

- **Drag & Drop** — drag video files directly onto the window
- **Video analysis** — displays format, duration, resolution, codec, FPS, and bitrates (Mbps)
- **H.264 compression** — target bitrate in Mbps, adjustable encoding preset and audio handling
- **Progress tracking** — real-time progress bar and FFmpeg log output
- **Self-contained** — FFmpeg/FFprobe binaries are embedded inside the executable

### Supported Platforms

| Platform | Architecture | Binary |
|----------|-------------|--------|
| macOS | Apple Silicon (arm64) | `videocompress-darwin-arm64` |
| Windows | x86-64 | `videocompress-windows-amd64.exe` |

### Download

Pre-built binaries are available on the [Releases](https://github.com/punkisnotdead3/videocompress/releases) page.
Download the binary for your platform and run it directly — no installation required.

### Usage

1. Launch the application.
2. **Select a video** by clicking "选择视频文件" or **drag and drop** a video file onto the window.
3. The app analyses the file and shows detailed metadata. Bitrates are highlighted in red for quick reference.
4. Set the **target bitrate** in Mbps (recommended: 8 Mbps for high quality).
   The app pre-fills a suggested value at 50% of the source bitrate.
5. Choose an **encoding preset** (slower = smaller file, faster = quicker encode).
6. Select **audio handling** (copy original or re-encode at 128 / 192 / 256 kbps).
7. Click **"开始压缩"** to start. The output file is saved next to the source with a `_compressed` suffix.

### Supported Input Formats

`.mp4` `.mov` `.avi` `.mkv` `.wmv` `.flv` `.webm` `.m4v` `.ts` `.mts`

### Build from Source

**Requirements**

- Go 1.21+
- GCC (Linux/Windows) or Xcode Command Line Tools (macOS)
- FFmpeg static binaries (downloaded by the setup scripts below)

**macOS**

```bash
bash setup-darwin.sh   # downloads static ffmpeg/ffprobe into assets/
make build-darwin
# output: dist/videocompress-darwin-arm64
```

**Windows** (PowerShell)

```powershell
.\setup.ps1            # downloads static ffmpeg/ffprobe into assets\
go build -ldflags="-s -w -H windowsgui" -o dist\videocompress.exe .
```

### CI / CD

Pushing a version tag (e.g. `v1.2.0`) triggers GitHub Actions to build both platforms and publish a GitHub Release with the binaries attached.

---

## 中文

跨平台桌面视频压缩工具，深色主题 GUI，基于 **Go + Fyne** 构建，内嵌 **FFmpeg 静态二进制**，无需任何外部依赖。

### 功能特性

- **拖拽支持** — 直接将视频文件拖入窗口即可加载
- **视频分析** — 展示格式、时长、分辨率、编码、帧率、码率（Mbps，红色高亮）
- **H.264 压缩** — 以 Mbps 为单位设置目标码率，可调编码速度预设和音频处理方式
- **实时进度** — 进度条 + FFmpeg 日志实时输出
- **开箱即用** — FFmpeg/FFprobe 已内嵌到可执行文件中，无需单独安装

### 支持平台

| 平台 | 架构 | 可执行文件 |
|------|------|-----------|
| macOS | Apple Silicon (arm64) | `videocompress-darwin-arm64` |
| Windows | x86-64 | `videocompress-windows-amd64.exe` |

### 下载

在 [Releases](https://github.com/punkisnotdead3/videocompress/releases) 页面下载对应平台的二进制文件，直接运行，无需安装。

### 使用方法

1. 启动程序。
2. 点击 **"选择视频文件"** 按钮，或直接将视频文件**拖拽到窗口**中。
3. 程序自动分析视频，展示详细信息。码率字段以红色显示，便于快速识别。
4. 设置**目标码率**（单位：Mbps，建议 8 Mbps）。
   程序会自动预填源视频码率的 50% 作为建议值。
5. 选择**编码速度预设**（越慢文件越小，越快编码越快）。
6. 选择**音频处理**方式（复制原始音频，或重新编码为 128 / 192 / 256 kbps）。
7. 点击 **"开始压缩"**，输出文件保存在原视频同目录下，文件名后缀为 `_compressed`。

### 支持的输入格式

`.mp4` `.mov` `.avi` `.mkv` `.wmv` `.flv` `.webm` `.m4v` `.ts` `.mts`

### 从源码构建

**环境要求**

- Go 1.21+
- GCC（Linux/Windows）或 Xcode Command Line Tools（macOS）
- FFmpeg 静态二进制（由下方脚本自动下载）

**macOS**

```bash
bash setup-darwin.sh   # 下载 ffmpeg/ffprobe 到 assets/ 目录
make build-darwin
# 输出: dist/videocompress-darwin-arm64
```

**Windows**（PowerShell）

```powershell
.\setup.ps1            # 下载 ffmpeg/ffprobe 到 assets\ 目录
go build -ldflags="-s -w -H windowsgui" -o dist\videocompress.exe .
```

### CI / CD

推送版本 tag（如 `v1.2.0`）后，GitHub Actions 自动为两个平台构建可执行文件并发布 GitHub Release。

---

## License

MIT
