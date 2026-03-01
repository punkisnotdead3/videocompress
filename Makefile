.PHONY: setup-darwin-arm64 setup-darwin-amd64 setup-windows \
        build-darwin build-windows dist-darwin clean help

BINARY  := videocompress
DIST    := dist
ASSETS  := assets
FFMPEG_VER := 7.1

# ─────────────────────────────────────────────────────────────────────────────
help:
	@echo "Setup (run once to download ffmpeg):"
	@echo "  make setup-darwin-arm64   macOS Apple Silicon ffmpeg"
	@echo "  make setup-darwin-amd64   macOS Intel ffmpeg"
	@echo "  make setup-windows        Windows x64 ffmpeg"
	@echo ""
	@echo "Build:"
	@echo "  make build-darwin         Build macOS universal binary (needs both setups)"
	@echo "  make build-windows        Build Windows .exe (needs setup-windows)"
	@echo "  make dist-darwin          Bundle macOS .app (needs fyne CLI)"
	@echo "  make clean                Remove dist/"

# ── Download macOS arm64 ffmpeg (Apple Silicon) ───────────────────────────────
# Source: https://evermeet.cx — static builds for macOS arm64 and amd64
setup-darwin-arm64:
	@mkdir -p $(ASSETS)
	@echo ">>> Downloading ffmpeg $(FFMPEG_VER) for macOS arm64 (Apple Silicon)..."
	curl -fsSL "https://evermeet.cx/pub/ffmpeg/ffmpeg-$(FFMPEG_VER).zip" \
	     -o /tmp/vc-ffmpeg-darwin-arm64.zip
	unzip -o /tmp/vc-ffmpeg-darwin-arm64.zip ffmpeg -d /tmp/vc-ffmpeg-darwin-arm64/
	cp /tmp/vc-ffmpeg-darwin-arm64/ffmpeg $(ASSETS)/ffmpeg-darwin-arm64
	chmod +x $(ASSETS)/ffmpeg-darwin-arm64

	curl -fsSL "https://evermeet.cx/pub/ffprobe/ffprobe-$(FFMPEG_VER).zip" \
	     -o /tmp/vc-ffprobe-darwin-arm64.zip
	unzip -o /tmp/vc-ffprobe-darwin-arm64.zip ffprobe -d /tmp/vc-ffprobe-darwin-arm64/
	cp /tmp/vc-ffprobe-darwin-arm64/ffprobe $(ASSETS)/ffprobe-darwin-arm64
	chmod +x $(ASSETS)/ffprobe-darwin-arm64
	@echo "✓ arm64 binaries ready"; ls -lh $(ASSETS)/ffmpeg-darwin-arm64 $(ASSETS)/ffprobe-darwin-arm64

# ── Download macOS amd64 ffmpeg (Intel Mac) ───────────────────────────────────
setup-darwin-amd64:
	@mkdir -p $(ASSETS)
	@echo ">>> Downloading ffmpeg $(FFMPEG_VER) for macOS amd64 (Intel)..."
	curl -fsSL "https://evermeet.cx/pub/ffmpeg/ffmpeg-$(FFMPEG_VER)-x86_64.zip" \
	     -o /tmp/vc-ffmpeg-darwin-amd64.zip
	unzip -o /tmp/vc-ffmpeg-darwin-amd64.zip ffmpeg -d /tmp/vc-ffmpeg-darwin-amd64/
	cp /tmp/vc-ffmpeg-darwin-amd64/ffmpeg $(ASSETS)/ffmpeg-darwin-amd64
	chmod +x $(ASSETS)/ffmpeg-darwin-amd64

	curl -fsSL "https://evermeet.cx/pub/ffprobe/ffprobe-$(FFMPEG_VER)-x86_64.zip" \
	     -o /tmp/vc-ffprobe-darwin-amd64.zip
	unzip -o /tmp/vc-ffprobe-darwin-amd64.zip ffprobe -d /tmp/vc-ffprobe-darwin-amd64/
	cp /tmp/vc-ffprobe-darwin-amd64/ffprobe $(ASSETS)/ffprobe-darwin-amd64
	chmod +x $(ASSETS)/ffprobe-darwin-amd64
	@echo "✓ amd64 binaries ready"; ls -lh $(ASSETS)/ffmpeg-darwin-amd64 $(ASSETS)/ffprobe-darwin-amd64

# ── Download Windows x64 ffmpeg ───────────────────────────────────────────────
# Source: https://github.com/GyanD/codexffmpeg
setup-windows:
	@mkdir -p $(ASSETS)
	@echo ">>> Downloading ffmpeg $(FFMPEG_VER) for Windows x64..."
	curl -fsSL \
	  "https://github.com/GyanD/codexffmpeg/releases/download/$(FFMPEG_VER)/ffmpeg-$(FFMPEG_VER)-essentials_build.zip" \
	  -o /tmp/vc-ffmpeg-windows.zip
	unzip -o /tmp/vc-ffmpeg-windows.zip -d /tmp/vc-ffmpeg-windows/
	find /tmp/vc-ffmpeg-windows -name "ffmpeg.exe"  -exec cp {} $(ASSETS)/ffmpeg-windows.exe  \;
	find /tmp/vc-ffmpeg-windows -name "ffprobe.exe" -exec cp {} $(ASSETS)/ffprobe-windows.exe \;
	@echo "✓ Windows binaries ready"; ls -lh $(ASSETS)/ffmpeg-windows.exe $(ASSETS)/ffprobe-windows.exe

# ── Build macOS universal binary ──────────────────────────────────────────────
# Requires both setup-darwin-arm64 and setup-darwin-amd64 to have been run.
# Each arch slice embeds its own matching ffmpeg binary.
build-darwin: \
  $(ASSETS)/ffmpeg-darwin-arm64 $(ASSETS)/ffprobe-darwin-arm64 \
  $(ASSETS)/ffmpeg-darwin-amd64 $(ASSETS)/ffprobe-darwin-amd64
	@mkdir -p $(DIST)
	@echo ">>> Building arm64 slice..."
	GOOS=darwin GOARCH=arm64 go build \
	    -ldflags="-s -w" \
	    -o $(DIST)/$(BINARY)-darwin-arm64 .
	@echo ">>> Building amd64 slice..."
	GOOS=darwin GOARCH=amd64 go build \
	    -ldflags="-s -w" \
	    -o $(DIST)/$(BINARY)-darwin-amd64 .
	@echo ">>> Creating universal binary with lipo..."
	lipo -create -output $(DIST)/$(BINARY)-darwin-universal \
	    $(DIST)/$(BINARY)-darwin-arm64 \
	    $(DIST)/$(BINARY)-darwin-amd64
	@rm -f $(DIST)/$(BINARY)-darwin-arm64 $(DIST)/$(BINARY)-darwin-amd64
	@echo "✓ Done:"; ls -lh $(DIST)/$(BINARY)-darwin-universal

# ── Build Windows .exe ────────────────────────────────────────────────────────
# On Windows: just run  go build -ldflags="-s -w -H windowsgui" -o dist\videocompress.exe .
# Cross-compile from macOS requires mingw-w64: brew install mingw-w64
build-windows: $(ASSETS)/ffmpeg-windows.exe $(ASSETS)/ffprobe-windows.exe
	@mkdir -p $(DIST)
	CGO_ENABLED=1 \
	CC=x86_64-w64-mingw32-gcc \
	GOOS=windows GOARCH=amd64 \
	go build -ldflags="-s -w -H windowsgui" \
	    -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo "✓ Done:"; ls -lh $(DIST)/$(BINARY)-windows-amd64.exe

# ── macOS .app bundle ─────────────────────────────────────────────────────────
dist-darwin: build-darwin
	@command -v fyne >/dev/null 2>&1 || \
	    (echo "Install fyne CLI: go install fyne.io/fyne/v2/cmd/fyne@latest" && exit 1)
	fyne package -os darwin -name "Video Compressor"

# ── Clean ─────────────────────────────────────────────────────────────────────
clean:
	rm -rf $(DIST)

# ── Guard rules ───────────────────────────────────────────────────────────────
$(ASSETS)/ffmpeg-darwin-arm64 $(ASSETS)/ffprobe-darwin-arm64:
	@echo "ERROR: $@ missing — run: make setup-darwin-arm64" && exit 1

$(ASSETS)/ffmpeg-darwin-amd64 $(ASSETS)/ffprobe-darwin-amd64:
	@echo "ERROR: $@ missing — run: make setup-darwin-amd64" && exit 1

$(ASSETS)/ffmpeg-windows.exe $(ASSETS)/ffprobe-windows.exe:
	@echo "ERROR: $@ missing — run: make setup-windows" && exit 1
