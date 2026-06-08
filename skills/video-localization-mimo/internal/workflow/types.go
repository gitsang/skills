// Package workflow provides a pipeline-based workflow engine for video localization.
// It orchestrates the multi-step process of extracting audio, transcribing,
// translating, synthesizing speech, and composing the final video.
package workflow

import (
	"fmt"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/pkg/srt"
)

// StepStatus represents the execution status of a workflow step.
type StepStatus string

const (
	StepPending   StepStatus = "pending"
	StepRunning   StepStatus = "running"
	StepCompleted StepStatus = "completed"
	StepFailed    StepStatus = "failed"
)

// State holds the shared state passed between workflow steps.
// Each step reads from and writes to this state, enabling data flow
// through the pipeline without direct coupling between steps.
type State struct {
	// SourceVideo is the path to the input video file
	SourceVideo string

	// SourceLang is the source language code (e.g., "en", "zh", "auto")
	SourceLang string

	// TargetLang is the target language code (e.g., "zh", "en")
	TargetLang string

	// OutputDir is the base directory for all output files
	OutputDir string

	// Voice is the TTS voice name (e.g., "Chloe")
	Voice string

	// Speed is the TTS speech rate multiplier (0.5-2.0)
	Speed float64

	// CloneRef is the path to the reference audio for voice cloning (optional)
	CloneRef string

	// SourceAudio is the path to the extracted audio file
	SourceAudio string

	// SourceSRT contains the parsed source subtitle segments
	SourceSRT []srt.Segment

	// SourceText is the transcribed source text
	SourceText string

	// TargetText is the translated target text
	TargetText string

	// TargetSRT contains the translated subtitle segments
	TargetSRT []srt.Segment

	// TargetAudio is the path to the synthesized target audio
	TargetAudio string

	// BilingualSRT is the path to the generated bilingual subtitle file
	BilingualSRT string

	// FinalVideo is the path to the composed final video
	FinalVideo string
}

// Request contains the parameters for a video localization workflow.
type Request struct {
	// SourceVideo is the path to the input video file (required)
	SourceVideo string

	// SourceLang is the source language code (default: "auto")
	SourceLang string

	// TargetLang is the target language code (default: "zh")
	TargetLang string

	// OutputDir is the base directory for output files (default: "outputs/")
	OutputDir string

	// Voice is the TTS voice name (optional, uses config default)
	Voice string

	// Speed is the TTS speech rate (default: 1.0)
	Speed float64

	// CloneRef is the path to reference audio for voice cloning (optional)
	CloneRef string
}

// Validate checks if the request has valid parameters.
func (r *Request) Validate() error {
	if r.SourceVideo == "" {
		return fmt.Errorf("source video path is required")
	}
	if r.TargetLang == "" {
		return fmt.Errorf("target language is required")
	}
	if r.Speed < 0.5 || r.Speed > 2.0 {
		return fmt.Errorf("speed must be between 0.5 and 2.0, got %.1f", r.Speed)
	}
	return nil
}

// StepResult records the outcome of a single workflow step execution.
type StepResult struct {
	// Name is the display name of the step
	Name string `json:"name"`

	// Status is the final execution status
	Status StepStatus `json:"status"`

	// Error is the error message if the step failed (empty on success)
	Error string `json:"error,omitempty"`

	// Duration is how long the step took to execute
	Duration time.Duration `json:"duration"`
}
