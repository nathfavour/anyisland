package visual

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type RecordedFrame struct {
	Content string
	Ticks   int
}

type RecordingSession struct {
	ID     string
	Frames []RecordedFrame
	Mu     sync.Mutex
}

func NewRecordingSession(id string) *RecordingSession {
	return &RecordingSession{
		ID: id,
	}
}

func (s *RecordingSession) AddFrame(content string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if len(s.Frames) > 0 && s.Frames[len(s.Frames)-1].Content == content {
		s.Frames[len(s.Frames)-1].Ticks++
		return
	}

	s.Frames = append(s.Frames, RecordedFrame{
		Content: content,
		Ticks:   1,
	})
}

func GetBestEncoder() string {
	if runtime.GOOS == "darwin" {
		return "h264_videotoolbox"
	}
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		cmd := exec.Command("ffmpeg", "-hide_banner", "-encoders")
		if out, err := cmd.Output(); err == nil && strings.Contains(string(out), "h264_nvenc") {
			return "h264_nvenc"
		}
	}
	return "libx264"
}

func ProcessRecording(frames []RecordedFrame, outDir string) (string, error) {
	if len(frames) == 0 {
		return "", fmt.Errorf("no frames recorded")
	}

	_ = os.MkdirAll(outDir, 0755)

	numFrames := len(frames)
	rgbDatas := make([][]byte, numFrames)
	var width, height int

	type renderResult struct {
		data []byte
		w, h int
		err  error
	}
	cache := make(map[string]*renderResult)
	var cacheMu sync.Mutex

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())
	var processedCount int32

	for i := range frames {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			ansi := frames[idx].Content
			cacheMu.Lock()
			res, ok := cache[ansi]
			cacheMu.Unlock()

			if !ok {
				data, w, h, err := RenderAnsiToRGB(ansi)
				res = &renderResult{data, w, h, err}
				cacheMu.Lock()
				cache[ansi] = res
				cacheMu.Unlock()
			}

			if res.err == nil {
				rgbDatas[idx] = res.data
				cacheMu.Lock()
				if width == 0 {
					width = res.w
					height = res.h
				}
				cacheMu.Unlock()
			}

			atomic.AddInt32(&processedCount, 1)
		}(i)
	}
	wg.Wait()

	timestamp := time.Now().Format("2006-01-02_150405")
	finalPath := filepath.Join(outDir, fmt.Sprintf("anyisland_rec_%s.mp4", timestamp))

	encoder := GetBestEncoder()
	args := []string{
		"-y",
		"-framerate", "24",
		"-f", "rawvideo",
		"-pixel_format", "rgb24",
		"-video_size", fmt.Sprintf("%dx%d", width, height),
		"-i", "-",
		"-c:v", encoder,
	}

	if encoder == "libx264" {
		args = append(args, "-preset", "slower", "-crf", "17", "-tune", "stillimage")
	} else if encoder == "h264_nvenc" {
		args = append(args, "-preset", "slow", "-cq", "17", "-rc", "vbr")
	} else if encoder == "h264_videotoolbox" {
		args = append(args, "-realtime", "false", "-q:v", "90")
	}

	args = append(args,
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-vf", "scale='min(1920,iw)':-2:force_original_aspect_ratio=decrease:flags=lanczos,pad=ceil(iw/2)*2:ceil(ih/2)*2",
		finalPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("ffmpeg pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ffmpeg start failed: %w", err)
	}

	// Feed the pipe with frames
	go func() {
		defer stdin.Close()
		for i, frame := range frames {
			data := rgbDatas[i]
			if data == nil {
				continue
			}
			for j := 0; j < frame.Ticks; j++ {
				_, _ = stdin.Write(data)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("ffmpeg execution failed: %w", err)
	}

	return finalPath, nil
}
