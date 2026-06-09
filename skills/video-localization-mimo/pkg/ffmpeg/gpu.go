package ffmpeg

import (
	"context"
	"fmt"
	"strings"
)

type GPUType string

const (
	GPUTypeNVENC        GPUType = "nvenc"
	GPUTypeQSV          GPUType = "qsv"
	GPUTypeAMF          GPUType = "amf"
	GPUTypeVideoToolbox GPUType = "videotoolbox"
	GPUTypeVAAPI        GPUType = "vaapi"
	GPUTypeNone         GPUType = "none"
)

type GPUInfo struct {
	Type      GPUType
	Encoder   string
	Available bool
}

type GPUDetector struct {
	runner *Runner
}

func NewGPUDetector(runner *Runner) *GPUDetector {
	return &GPUDetector{runner: runner}
}

func (d *GPUDetector) DetectGPU(ctx context.Context) (*GPUInfo, error) {
	gpuChecks := []struct {
		name    string
		gpuType GPUType
		encoder string
	}{
		{"h264_nvenc", GPUTypeNVENC, "h264_nvenc"},
		{"h264_qsv", GPUTypeQSV, "h264_qsv"},
		{"h264_amf", GPUTypeAMF, "h264_amf"},
		{"h264_videotoolbox", GPUTypeVideoToolbox, "h264_videotoolbox"},
		{"h264_vaapi", GPUTypeVAAPI, "h264_vaapi"},
	}

	for _, check := range gpuChecks {
		available, err := d.isEncoderAvailable(ctx, check.name)
		if err != nil {
			continue
		}
		if available {
			return &GPUInfo{
				Type:      check.gpuType,
				Encoder:   check.encoder,
				Available: true,
			}, nil
		}
	}

	return &GPUInfo{
		Type:      GPUTypeNone,
		Encoder:   "libx264",
		Available: false,
	}, nil
}

func (d *GPUDetector) isEncoderAvailable(ctx context.Context, encoderName string) (bool, error) {
	output, err := d.runner.RunProbe(ctx, "-encoders")
	if err != nil {
		return false, fmt.Errorf("listing encoders: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, encoderName) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "V") || strings.HasPrefix(trimmed, "A") {
				parts := strings.Fields(trimmed)
				if len(parts) >= 2 && parts[1] == encoderName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func (g *GPUInfo) GetEncoderArgs() []string {
	if !g.Available {
		return []string{
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "23",
		}
	}

	switch g.Type {
	case GPUTypeNVENC:
		return []string{
			"-c:v", "h264_nvenc",
			"-preset", "medium",
			"-cq", "23",
		}
	case GPUTypeQSV:
		return []string{
			"-c:v", "h264_qsv",
			"-preset", "medium",
			"-global_quality", "23",
		}
	case GPUTypeAMF:
		return []string{
			"-c:v", "h264_amf",
			"-quality", "balanced",
			"-qp", "23",
		}
	case GPUTypeVideoToolbox:
		return []string{
			"-c:v", "h264_videotoolbox",
			"-q:v", "23",
		}
	case GPUTypeVAAPI:
		return []string{
			"-c:v", "h264_vaapi",
			"-qp", "23",
		}
	default:
		return []string{
			"-c:v", "libx264",
			"-preset", "medium",
			"-crf", "23",
		}
	}
}

func (g *GPUInfo) String() string {
	if !g.Available {
		return "No GPU encoder available (using software encoding)"
	}
	return fmt.Sprintf("GPU encoder: %s (%s)", g.Encoder, g.Type)
}