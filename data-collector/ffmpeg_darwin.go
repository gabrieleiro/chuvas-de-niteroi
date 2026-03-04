//go:build darwin && arm64

package main

import _ "embed"

//go:embed bin/ffmpeg_darwin_arm64
var ffmpegBinary []byte
