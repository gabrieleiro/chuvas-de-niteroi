//go:build linux && amd64

package main

import _ "embed"

//go:embed bin/ffmpeg_linux_amd64
var ffmpegBinary []byte
