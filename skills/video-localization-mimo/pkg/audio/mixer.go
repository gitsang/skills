// Package audio provides audio processing utilities for video localization.
// It includes functionality for mixing multiple audio tracks and aligning
// audio segments to SRT subtitle timestamps.
package audio

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

// MixTracks mixes multiple audio files into a single output file.
// Each track can have its own volume level. The mixing uses FFmpeg's
// amix filter, which combines audio streams by summing them together.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - tracks: list of paths to input audio files
//   - output: path for the mixed output audio file
//   - volumes: volume multiplier for each track (must match tracks length).
//     Use 1.0 for original volume, 0.5 for half volume, etc.
//
// Returns an error if:
//   - tracks and volumes have different lengths
//   - no tracks are provided
//   - the mixing operation fails
func MixTracks(ctx context.Context, tracks []string, output string, volumes []float64) error {
	if len(tracks) == 0 {
		return fmt.Errorf("at least one track is required")
	}
	if len(tracks) != len(volumes) {
		return fmt.Errorf("tracks count (%d) must match volumes count (%d)", len(tracks), len(volumes))
	}

	// Build FFmpeg command
	args := []string{}

	// Add input files
	for _, track := range tracks {
		args = append(args, "-i", track)
	}

	// Build filter complex for mixing
	// Format: [0:a]volume=1.0[a0];[1:a]volume=0.5[a1];[a0][a1]amix=inputs=2:duration=longest
	filterComplex := ""
	for i := range tracks {
		filterComplex += fmt.Sprintf("[%d:a]volume=%.1f[a%d];", i, volumes[i], i)
	}

	// Add amix filter
	mixInputs := ""
	for i := range tracks {
		mixInputs += fmt.Sprintf("[a%d]", i)
	}
	filterComplex += fmt.Sprintf("%samix=inputs=%d:duration=longest[out]", mixInputs, len(tracks))

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[out]",
		"-y", // Overwrite output
		output,
	)

	// Execute FFmpeg
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mixing tracks: %w\noutput: %s", err, string(output))
	}

	return nil
}

// MixTracksSimple mixes multiple audio files with equal volume.
// This is a convenience function that sets all track volumes to 1.0.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - tracks: list of paths to input audio files
//   - output: path for the mixed output audio file
//
// Returns an error if the mixing fails.
func MixTracksSimple(ctx context.Context, tracks []string, output string) error {
	volumes := make([]float64, len(tracks))
	for i := range volumes {
		volumes[i] = 1.0
	}
	return MixTracks(ctx, tracks, output, volumes)
}

// formatTimestamp formats a duration in seconds to FFmpeg timestamp format (HH:MM:SS.mmm).
func formatTimestamp(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := seconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%02d:%02d:%s", hours, minutes, formatSeconds(secs))
}

// formatSeconds formats seconds to SS.mmm format.
func formatSeconds(seconds float64) string {
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}
