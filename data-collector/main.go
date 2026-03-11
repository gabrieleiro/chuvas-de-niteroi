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

type IssueList struct {
	issues []error
}

func (il IssueList) Add(e error) {
	if e != nil {
		il.issues = append(il.issues, e)
	}
}

func (il IssueList) IsEmpty() bool {
	return len(il.issues) == 0
}

func (il IssueList) String() string {
	var sb strings.Builder

	for _, i := range il.issues {
		sb.WriteString(i.Error())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

var (
	infoLogger   = log.New(os.Stdout, "INFO: ",  log.Ldate | log.Ltime)
	errorLogger  = log.New(os.Stderr, "ERROR: ", log.Ldate | log.Ltime)
)

func logInfo(msg string, v ...any) {
	infoLogger.Println(fmt.Sprintf(msg, v))
}

func logError(msg string, v ...any) {
	errorLogger.Println(fmt.Sprintf(msg, v))
}

func initFfmpeg() (success bool) {
	success = true

	platform := runtime.GOOS + "_" + runtime.GOARCH
	logInfo("Initializing ffmpeg for platform: %s\n", platform)

	tmpDir, err := os.MkdirTemp("", "ffmpeg-")
	if err != nil {
		success = false
		logError("creating temporary directory: %w", err)
	}

	ffmpegPath = filepath.Join(tmpDir, "ffmpeg")

	err = os.WriteFile(ffmpegPath, ffmpegBinary, 0755)
	if err != nil {
		success = false
		logError("extracting ffmpeg binary: %w", err)
	}

	return success
}

func snapshotFromCamera(cameraId string) {
	now := time.Now().UTC().Format(time.RFC3339)
	url := fmt.Sprintf("%s/%s/last_video.mp4", nittransBaseUrl, cameraId)
	videoFileName := fmt.Sprintf("%s_cam%s.mp4", now, cameraId)

	resp, err := http.Get(url)
	if err != nil {
		logError("Camera %s: Request failed: %w", cameraId, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logError("Camera %s: Failed with status %s", cameraId, resp.Status)
		return
	}

	out, err := os.Create(videoFileName)
	if err != nil {
		logError("Camera %s: Failed to create video file: %v", cameraId, err)
		return
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		logError("Camera %s: Failed to download video: %v", cameraId, err)
		return
	}

	logInfo("extracting frame %s", videoFileName)
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
		logError("Camera %s: Error extracting frame: %v\n%s", videoFileName, err, stderrBuf.String())
	} else {
		logInfo("Camera %s: Success -> %s", videoFileName, frameFileName)
	}

	os.Remove(videoFileName)
}

func main() {
	log.Println("Starting NitTrans camera downloads...")

	if !initFfmpeg() {
		log.Fatalf("Failed to initialize ffmpeg")
	}

	go func() {
		ticker := time.NewTicker(time.Minute)

		for range ticker.C {
			logInfo("new batch")
			for _, id := range cameraIds {
				snapshotFromCamera(id)
				time.Sleep(time.Second)
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logInfo("shutting down...")
}
