//go:build linux && amd64

package embed

import _ "embed"

//go:embed bin/ffmpeg_linux_amd64
var FfmpegBinary []byte
