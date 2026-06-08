package ffmpeg

import (
	"context"
	"fmt"
	"strconv"
)

// ExtractAudio extracts the audio track from a video file and saves it as a WAV file.
// The output audio is converted to 16-bit PCM, mono, 16kHz sample rate format,
// which is suitable for speech recognition and processing.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - input: path to the input video file
//   - output: path for the output WAV file
//
// Returns an error if the extraction fails.
func (r *Runner) ExtractAudio(ctx context.Context, input, output string) error {
	args := []string{
		"-i", input,
		"-vn",                    // No video
		"-acodec", "pcm_s16le",   // 16-bit PCM
		"-ar", "16000",           // 16kHz sample rate
		"-ac", "1",               // Mono
		"-y",                     // Overwrite output
		output,
	}

	if err := r.Run(ctx, args...); err != nil {
		return fmt.Errorf("extracting audio from %s: %w", input, err)
	}

	return nil
}

// ConvertAudio converts an audio file to the specified format with the given sample rate.
// The output format is determined by the output file extension (e.g., .wav, .mp3, .ogg).
// The output is always mono (single channel) for consistency.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - input: path to the input audio file
//   - output: path for the output audio file
//   - sampleRate: desired sample rate in Hz (e.g., 16000, 44100, 48000)
//
// Returns an error if the conversion fails.
func (r *Runner) ConvertAudio(ctx context.Context, input, output string, sampleRate int) error {
	args := []string{
		"-i", input,
		"-ar", strconv.Itoa(sampleRate),
		"-ac", "1", // Mono
		"-y", // Overwrite output
		output,
	}

	if err := r.Run(ctx, args...); err != nil {
		return fmt.Errorf("converting audio %s: %w", input, err)
	}

	return nil
}
