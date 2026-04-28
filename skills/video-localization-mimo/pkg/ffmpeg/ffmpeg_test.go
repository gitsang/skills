package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

const testVideoPath = "../../inputs/USENIX ATC '19 - Zanzibar： Google\u2019s Consistent, Global Authorization System [mstZT431AeQ].mkv"

func newTestRunner(t *testing.T) *Runner {
	t.Helper()
	return NewRunner("/usr/bin/ffmpeg", "/usr/bin/ffprobe")
}

func TestProbe(t *testing.T) {
	if _, err := os.Stat(testVideoPath); os.IsNotExist(err) {
		t.Skipf("test video not found: %s", testVideoPath)
	}

	runner := newTestRunner(t)
	ctx := context.Background()

	info, err := runner.Probe(ctx, testVideoPath)
	if err != nil {
		t.Fatalf("Probe failed: %v", err)
	}

	if info.Duration <= 0 {
		t.Errorf("expected positive duration, got %f", info.Duration)
	}
	if info.Size <= 0 {
		t.Errorf("expected positive size, got %d", info.Size)
	}
	if info.AudioCodec == "" {
		t.Error("expected non-empty audio codec")
	}
	if info.AudioSampleRate <= 0 {
		t.Errorf("expected positive sample rate, got %d", info.AudioSampleRate)
	}
	if info.AudioChannels <= 0 {
		t.Errorf("expected positive channels, got %d", info.AudioChannels)
	}
	if info.VideoCodec == "" {
		t.Error("expected non-empty video codec")
	}
	if info.Width <= 0 || info.Height <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", info.Width, info.Height)
	}

	t.Logf("Media info: duration=%.2fs, size=%d, audio=%s/%dHz/%dch, video=%s/%dx%d",
		info.Duration, info.Size,
		info.AudioCodec, info.AudioSampleRate, info.AudioChannels,
		info.VideoCodec, info.Width, info.Height)
}

func TestExtractAudio(t *testing.T) {
	if _, err := os.Stat(testVideoPath); os.IsNotExist(err) {
		t.Skipf("test video not found: %s", testVideoPath)
	}

	runner := newTestRunner(t)
	ctx := context.Background()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "extracted_audio.wav")

	if err := runner.ExtractAudio(ctx, testVideoPath, outputPath); err != nil {
		t.Fatalf("ExtractAudio failed: %v", err)
	}

	stat, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if stat.Size() == 0 {
		t.Error("output file is empty")
	}

	t.Logf("Extracted audio: %s (%d bytes)", outputPath, stat.Size())
}
