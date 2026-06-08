package ffmpeg

import (
	"context"
	"fmt"
)

// ComposeOptions contains the configuration for composing a final video.
// It supports replacing the audio track, burning in subtitles, and
// adjusting audio volume.
type ComposeOptions struct {
	// VideoPath is the path to the input video file
	VideoPath string

	// AudioPath is the path to the replacement audio file.
	// If empty, the original audio is kept.
	AudioPath string

	// SubtitlePath is the path to the SRT subtitle file to burn in.
	// If empty, no subtitles are added.
	SubtitlePath string

	// OutputPath is the path for the output video file
	OutputPath string

	// AudioVolume is the volume multiplier for the audio track.
	// 1.0 = original volume, 0.5 = half volume, 2.0 = double volume.
	// If 0, defaults to 1.0 (original volume).
	AudioVolume float64
}

// Compose creates a final video by combining video, audio, and subtitle tracks.
// It supports three main operations:
//   - Audio replacement: replaces the original audio with a new audio track
//   - Subtitle burning: embeds SRT subtitles directly into the video frames
//   - Volume adjustment: adjusts the audio volume by a multiplier
//
// The operations can be combined. For example, you can replace the audio
// AND burn in subtitles in a single pass.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - opts: composition options specifying input files and settings
//
// Returns an error if the composition fails.
func (r *Runner) Compose(ctx context.Context, opts ComposeOptions) error {
	if opts.VideoPath == "" {
		return fmt.Errorf("video path is required")
	}
	if opts.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	// Default volume to 1.0 if not set
	volume := opts.AudioVolume
	if volume == 0 {
		volume = 1.0
	}

	args := []string{}

	// Add input video
	args = append(args, "-i", opts.VideoPath)

	// Add input audio if replacing
	hasAudioReplace := opts.AudioPath != ""
	if hasAudioReplace {
		args = append(args, "-i", opts.AudioPath)
	}

	// Build video filter chain
	videoFilters := []string{}

	// Add subtitle filter if specified
	if opts.SubtitlePath != "" {
		videoFilters = append(videoFilters, fmt.Sprintf("subtitles=%s", opts.SubtitlePath))
	}

	// Apply video filters if any
	if len(videoFilters) > 0 {
		filterStr := ""
		for i, f := range videoFilters {
			if i > 0 {
				filterStr += ","
			}
			filterStr += f
		}
		args = append(args, "-vf", filterStr)
	}

	// Build audio filter for volume adjustment
	if volume != 1.0 {
		args = append(args, "-af", fmt.Sprintf("volume=%.2f", volume))
	}

	// Configure stream mapping
	if hasAudioReplace {
		// Map video from first input, audio from second input
		args = append(args,
			"-c:v", "copy", // Copy video codec (no re-encoding unless filters applied)
			"-c:a", "aac", // Encode audio as AAC
			"-map", "0:v:0", // Use video from first input
			"-map", "1:a:0", // Use audio from second input
		)
	} else {
		// Keep original audio, just apply volume if needed
		args = append(args,
			"-c:v", "copy",
			"-c:a", "aac",
		)
	}

	// Overwrite output and use shortest stream duration
	args = append(args, "-y", opts.OutputPath)

	if err := r.Run(ctx, args...); err != nil {
		return fmt.Errorf("composing video: %w", err)
	}

	return nil
}

// ComposeWithAudioReplace creates a video with a replaced audio track.
// This is a convenience method for the common case of replacing audio.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - videoPath: path to the input video
//   - audioPath: path to the replacement audio
//   - outputPath: path for the output video
//
// Returns an error if the composition fails.
func (r *Runner) ComposeWithAudioReplace(ctx context.Context, videoPath, audioPath, outputPath string) error {
	return r.Compose(ctx, ComposeOptions{
		VideoPath:   videoPath,
		AudioPath:   audioPath,
		OutputPath:  outputPath,
		AudioVolume: 1.0,
	})
}

// ComposeWithSubtitles creates a video with burned-in subtitles.
// This is a convenience method for the common case of adding subtitles.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - videoPath: path to the input video
//   - subtitlePath: path to the SRT subtitle file
//   - outputPath: path for the output video
//
// Returns an error if the composition fails.
func (r *Runner) ComposeWithSubtitles(ctx context.Context, videoPath, subtitlePath, outputPath string) error {
	return r.Compose(ctx, ComposeOptions{
		VideoPath:    videoPath,
		SubtitlePath: subtitlePath,
		OutputPath:   outputPath,
		AudioVolume:  1.0,
	})
}
