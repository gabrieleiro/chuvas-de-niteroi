package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const nittransBaseUrl = "https://appnittrans.niteroi.rj.gov.br:8888"

var cameraIds = []string{
	"000001", "000002", "000003", "000004", "000005",
	"000006", "000007", "000009", "000010", "000012",
}

var ffmpegPath string

var videoDownloadChannel = make(chan string, 100)
var videoFilesChannel = make(chan string, 100)

func initFfmpeg() error {
	platform := runtime.GOOS + "_" + runtime.GOARCH
	log.Printf("Initializing ffmpeg for platform: %s\n", platform)

	tmpDir, err := os.MkdirTemp("", "ffmpeg-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	ffmpegPath = filepath.Join(tmpDir, "ffmpeg")

	err = os.WriteFile(ffmpegPath, ffmpegBinary, 0755)
	if err != nil {
		return fmt.Errorf("failed to write ffmpeg binary: %w", err)
	}

	log.Printf("FFmpeg extracted to: %s\n", ffmpegPath)
	return nil
}

func snapshotFromCamera(cameraId string) {
	now := time.Now().UTC().Format(time.RFC3339)
	url := fmt.Sprintf("%s/%s/last_video.mp4", nittransBaseUrl, cameraId)
	videoFileName := fmt.Sprintf("%s_cam%s.mp4", now, cameraId)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Camera %s: Request failed: %v\n", cameraId, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Camera %s: Failed with status %s\n", cameraId, resp.Status)
		return
	}

	out, err := os.Create(videoFileName)
	if err != nil {
		log.Printf("Camera %s: Failed to create video file: %v\n", cameraId, err)
		return
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		log.Printf("Camera %s: Failed to download video: %v\n", cameraId, err)
		return
	}

	log.Printf("Camera %s: Successfully downloaded video -> %s\n", cameraId, videoFileName)

	log.Printf("extracting frame %s\n", videoFileName)
	frameFileName := fmt.Sprintf("%s.jpg", strings.TrimSuffix(videoFileName, ".mp4"))

	cmd := exec.Command(ffmpegPath,
		"-i", "file:"+videoFileName,
		"-vframes", "1",
		"-f", "image2",
		"-update", "1",
		"-y", "file:"+frameFileName,
	)

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		log.Printf("Camera %s: Error extracting frame: %v\n%s\n", videoFileName, err, stderrBuf.String())
		return
	}

	os.Remove(videoFileName)

	log.Printf("Camera %s: Success -> %s\n", videoFileName, frameFileName)
}

func main() {
	log.Println("Starting NitTrans camera downloads...")

	err := initFfmpeg()
	if err != nil {
		log.Fatalf("Failed to initialize ffmpeg: %v\n", err)
	}

	go func() {
		ticker := time.NewTicker(time.Minute)

		for range ticker.C {
			log.Printf("new batch\n")
			for _, id := range cameraIds {
				snapshotFromCamera(id)
				time.Sleep(time.Second)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down...")
}
