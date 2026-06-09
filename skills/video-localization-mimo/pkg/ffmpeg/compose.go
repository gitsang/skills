package ffmpeg

import (
	"context"
	"fmt"
	"strings"
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

	volume := opts.AudioVolume
	if volume == 0 {
		volume = 1.0
	}

	args := []string{}
	args = append(args, "-i", opts.VideoPath)

	hasAudioReplace := opts.AudioPath != ""
	if hasAudioReplace {
		args = append(args, "-i", opts.AudioPath)
	}

	videoFilters := []string{}
	if opts.SubtitlePath != "" {
		escapedPath := escapeFFmpegPath(opts.SubtitlePath)
		videoFilters = append(videoFilters, fmt.Sprintf("subtitles=%s", escapedPath))
	}

	if len(videoFilters) > 0 {
		filterStr := ""
		for i, f := range videoFilters {
			if i > 0 {
				filterStr += ","
			}
			filterStr += f
		}
		args = append(args, "-vf", filterStr)
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "23")
	} else {
		args = append(args, "-c:v", "copy")
	}

	if volume != 1.0 {
		args = append(args, "-af", fmt.Sprintf("volume=%.2f", volume))
	}

	if hasAudioReplace {
		args = append(args,
			"-c:a", "aac",
			"-map", "0:v:0",
			"-map", "1:a:0",
		)
	} else {
		args = append(args, "-c:a", "aac")
	}

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

func escapeFFmpegPath(path string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`:`, `\:`,
		`'`, `\'`,
		`[`, `\[`,
		`]`, `\]`,
		`,`, `\,`,
		`;`, `\;`,
	)
	return replacer.Replace(path)
}
