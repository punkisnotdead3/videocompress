# setup.ps1 — Download ffmpeg for Windows build (run once before building)
# Usage: .\setup.ps1

$ErrorActionPreference = "Stop"
$ffmpegVer = "7.1"
$assetsDir = "assets"

if (!(Test-Path $assetsDir)) { New-Item -ItemType Directory -Path $assetsDir | Out-Null }

$url = "https://github.com/GyanD/codexffmpeg/releases/download/$ffmpegVer/ffmpeg-$ffmpegVer-essentials_build.zip"
$tmpZip = "$env:TEMP\vc-ffmpeg-windows.zip"
$tmpDir = "$env:TEMP\vc-ffmpeg-windows"

Write-Host "Downloading ffmpeg $ffmpegVer for Windows..."
Invoke-WebRequest -Uri $url -OutFile $tmpZip -UseBasicParsing

Write-Host "Extracting..."
if (Test-Path $tmpDir) { Remove-Item $tmpDir -Recurse -Force }
Expand-Archive -Path $tmpZip -DestinationPath $tmpDir

$ffmpegExe  = Get-ChildItem -Path $tmpDir -Recurse -Filter "ffmpeg.exe"  | Select-Object -First 1
$ffprobeExe = Get-ChildItem -Path $tmpDir -Recurse -Filter "ffprobe.exe" | Select-Object -First 1

Copy-Item $ffmpegExe.FullName  -Destination "$assetsDir\ffmpeg-windows.exe"  -Force
Copy-Item $ffprobeExe.FullName -Destination "$assetsDir\ffprobe-windows.exe" -Force

Write-Host "Done! Binaries placed in $assetsDir\"
Get-ChildItem "$assetsDir\ffmpeg-windows.exe", "$assetsDir\ffprobe-windows.exe" |
    Format-Table Name, @{N='Size';E={"{0:N2} MB" -f ($_.Length/1MB)}}
