// Package ffmpeg provides a Go wrapper around FFmpeg and FFprobe command-line tools.
// It supports media file analysis, audio extraction, format conversion, and video composition
// with subtitle burning and audio replacement capabilities.
package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Runner encapsulates the execution of FFmpeg and FFprobe commands.
// It holds the filesystem paths to both binaries and provides methods
// to execute FFmpeg operations with proper context cancellation support.
type Runner struct {
	ffmpegPath  string
	ffprobePath string
	gpuDetector *GPUDetector
}

// NewRunner creates a new Runner instance with the specified paths to FFmpeg tools.
//
// Parameters:
//   - ffmpegPath: path to the ffmpeg binary (e.g., "/usr/bin/ffmpeg")
//   - ffprobePath: path to the ffprobe binary (e.g., "/usr/bin/ffprobe")
//
// Returns a pointer to the newly created Runner.
func NewRunner(ffmpegPath, ffprobePath string) *Runner {
	return &Runner{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
	}
}

func (r *Runner) WithGPU() *Runner {
	r.gpuDetector = NewGPUDetector(r)
	return r
}

func (r *Runner) GetGPUInfo(ctx context.Context) (*GPUInfo, error) {
	if r.gpuDetector == nil {
		return &GPUInfo{
			Type:      GPUTypeNone,
			Encoder:   "libx264",
			Available: false,
		}, nil
	}
	return r.gpuDetector.DetectGPU(ctx)
}

// Run executes an FFmpeg command with the given arguments.
// The command is executed within the provided context, which supports
// cancellation and timeout. If the context is cancelled, the underlying
// process will be terminated.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - args: command-line arguments to pass to ffmpeg (excluding the binary name itself)
//
// Returns an error if the command fails to start or exits with a non-zero status.
// On failure, the error includes both stderr and stdout output for debugging.
func (r *Runner) Run(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, r.ffmpegPath, args...)

	// Capture both stdout and stderr for error reporting
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return nil
}

// RunProbe executes an FFprobe command with the given arguments.
// This is similar to Run but uses the ffprobe binary instead.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - args: command-line arguments to pass to ffprobe
//
// Returns the combined stdout output and any error encountered.
func (r *Runner) RunProbe(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.ffprobePath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w\nstdout: %s\nstderr: %s",
			err, stdout.String(), stderr.String())
	}

	return stdout.Bytes(), nil
}
