package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
)

// MediaInfo contains metadata extracted from a media file using FFprobe.
// It includes both format-level information (duration, size) and
// stream-specific information for audio and video tracks.
type MediaInfo struct {
	// Duration is the total duration of the media file in seconds
	Duration float64 `json:"duration"`

	// Size is the file size in bytes
	Size int64 `json:"size"`

	// AudioCodec is the name of the audio codec (e.g., "aac", "mp3", "opus")
	AudioCodec string `json:"audio_codec"`

	// AudioSampleRate is the audio sampling rate in Hz (e.g., 44100, 48000)
	AudioSampleRate int `json:"audio_sample_rate"`

	// AudioChannels is the number of audio channels (e.g., 1 for mono, 2 for stereo)
	AudioChannels int `json:"audio_channels"`

	// VideoCodec is the name of the video codec (e.g., "h264", "vp9")
	VideoCodec string `json:"video_codec"`

	// Width is the video width in pixels
	Width int `json:"width"`

	// Height is the video height in pixels
	Height int `json:"height"`
}

// ffprobeFormat represents the format section of ffprobe JSON output.
type ffprobeFormat struct {
	Duration string `json:"duration"`
	Size     string `json:"size"`
}

// ffprobeStream represents a single stream in ffprobe JSON output.
type ffprobeStream struct {
	CodecType  string `json:"codec_type"`
	CodecName  string `json:"codec_name"`
	SampleRate string `json:"sample_rate"`
	Channels   int    `json:"channels"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
}

// ffprobeOutput represents the complete ffprobe JSON output structure.
type ffprobeOutput struct {
	Format  ffprobeFormat   `json:"format"`
	Streams []ffprobeStream `json:"streams"`
}

// Probe extracts media information from the specified file using FFprobe.
// It retrieves format-level metadata (duration, file size) and stream-level
// metadata (codecs, resolution, sample rate, channels) for both audio
// and video streams.
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - path: absolute or relative path to the media file to analyze
//
// Returns a MediaInfo struct populated with the extracted metadata,
// or an error if the file cannot be read or parsed.
func (r *Runner) Probe(ctx context.Context, path string) (*MediaInfo, error) {
	// Use ffprobe with JSON output for reliable parsing
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration,size:stream=codec_type,codec_name,sample_rate,channels,width,height",
		"-of", "json",
		path,
	}

	output, err := r.RunProbe(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("probing %s: %w", path, err)
	}

	// Parse the JSON output
	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("parsing ffprobe output: %w", err)
	}

	// Build MediaInfo from the parsed output
	info := &MediaInfo{}

	// Parse format-level metadata
	if probe.Format.Duration != "" {
		if _, err := fmt.Sscanf(probe.Format.Duration, "%f", &info.Duration); err != nil {
			return nil, fmt.Errorf("parsing duration %q: %w", probe.Format.Duration, err)
		}
	}
	if probe.Format.Size != "" {
		if _, err := fmt.Sscanf(probe.Format.Size, "%d", &info.Size); err != nil {
			return nil, fmt.Errorf("parsing size %q: %w", probe.Format.Size, err)
		}
	}

	// Extract stream-specific metadata
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "audio":
			info.AudioCodec = stream.CodecName
			if stream.SampleRate != "" {
				if _, err := fmt.Sscanf(stream.SampleRate, "%d", &info.AudioSampleRate); err != nil {
					return nil, fmt.Errorf("parsing sample rate %q: %w", stream.SampleRate, err)
				}
			}
			info.AudioChannels = stream.Channels
		case "video":
			info.VideoCodec = stream.CodecName
			info.Width = stream.Width
			info.Height = stream.Height
		}
	}

	return info, nil
}
