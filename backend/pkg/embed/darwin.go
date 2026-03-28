//go:build darwin && arm64

package embed

import _ "embed"

//go:embed bin/ffmpeg_darwin_arm64
var FfmpegBinary []byte
