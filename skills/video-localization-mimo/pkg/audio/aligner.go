package audio

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/pkg/srt"
)

// AlignToSRT cuts an audio file into segments based on SRT subtitle timestamps.
// Each segment corresponds to a subtitle entry, allowing the audio to be
// aligned with the subtitle timing for processing (e.g., translation,
// re-recording with TTS).
//
// The function extracts audio segments using FFmpeg's -ss (start) and -t (duration)
// options for precise cutting. Output files are named sequentially as
// "segment_001.wav", "segment_002.wav", etc.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - audioPath: path to the input audio file
//   - segments: list of SRT segments defining the cut points
//   - outputDir: directory where segment files will be created
//
// Returns a list of paths to the created segment files, ordered by segment index.
// Returns an error if:
//   - the output directory cannot be created
//   - any segment extraction fails
func AlignToSRT(ctx context.Context, audioPath string, segments []srt.Segment, outputDir string) ([]string, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("at least one segment is required")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	// Extract each segment
	outputFiles := make([]string, 0, len(segments))
	for i, segment := range segments {
		// Calculate start time and duration from the segment
		// SRT timestamps are stored as time.Time with a reference date (2000-01-01)
		// We need to extract the time-of-day portion as seconds
		startSeconds := timeToSeconds(segment.Start)
		durationSeconds := segment.Duration().Seconds()

		// Generate output filename with zero-padded index
		outputFile := filepath.Join(outputDir, fmt.Sprintf("segment_%03d.wav", i+1))

		// Extract segment using FFmpeg
		if err := extractSegment(ctx, audioPath, outputFile, startSeconds, durationSeconds); err != nil {
			return nil, fmt.Errorf("extracting segment %d: %w", i+1, err)
		}

		outputFiles = append(outputFiles, outputFile)
	}

	return outputFiles, nil
}

// extractSegment extracts a single audio segment from the input file.
// It uses FFmpeg with -ss for seeking and -t for duration.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - input: path to the input audio file
//   - output: path for the output segment file
//   - startSeconds: start time in seconds from the beginning
//   - durationSeconds: duration of the segment in seconds
//
// Returns an error if the extraction fails.
func extractSegment(ctx context.Context, input, output string, startSeconds, durationSeconds float64) error {
	args := []string{
		"-i", input,
		"-ss", formatTimestamp(startSeconds),
		"-t", formatTimestamp(durationSeconds),
		"-acodec", "pcm_s16le", // 16-bit PCM
		"-ar", "16000", // 16kHz sample rate
		"-ac", "1", // Mono
		"-y", // Overwrite output
		output,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extracting segment: %w\noutput: %s", err, string(output))
	}

	return nil
}

// timeToSeconds converts a time.Time to seconds from midnight.
// SRT timestamps are typically stored as time.Time values where the
// date portion is irrelevant - only the time-of-day matters.
//
// Parameters:
//   - t: the time.Time value to convert
//
// Returns the number of seconds from midnight (00:00:00).
func timeToSeconds(t time.Time) float64 {
	return float64(t.Hour()*3600+t.Minute()*60+t.Second()) + float64(t.Nanosecond())/1e9
}
