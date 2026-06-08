package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/config"
	"github.com/gitsang/skills/video-localization-mimo/pkg/ffmpeg"
	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
)

type Pipeline struct {
	config       *config.Config
	mimoClient   *mimo.Client
	ffmpegRunner *ffmpeg.Runner
	steps        []Step
	progress     ProgressReporter
	results      []StepResult
}

func NewPipeline(cfg *config.Config) *Pipeline {
	mimoClient := mimo.NewClient(&mimo.ClientConfig{
		APIKey:  cfg.MiMo.APIKey,
		BaseURL: cfg.MiMo.BaseURL,
	})

	ffmpegRunner := ffmpeg.NewRunner(cfg.FFmpeg.Path, cfg.FFmpeg.FFprobePath)

	p := &Pipeline{
		config:       cfg,
		mimoClient:   mimoClient,
		ffmpegRunner: ffmpegRunner,
	}

	p.steps = []Step{
		NewExtractAudioStep(ffmpegRunner),
		NewTranscribeStep(mimoClient),
		NewTranslateStep(mimoClient),
		NewSynthesizeStep(mimoClient),
		NewAlignAudioStep(ffmpegRunner),
		NewGenerateSubtitleStep(),
		NewComposeStep(ffmpegRunner),
	}

	return p
}

func (p *Pipeline) SetProgressReporter(reporter ProgressReporter) {
	p.progress = reporter
}

func (p *Pipeline) GetResults() []StepResult {
	return p.results
}

func (p *Pipeline) Run(ctx context.Context, req *Request) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if err := os.MkdirAll(req.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	state := &State{
		SourceVideo: req.SourceVideo,
		SourceLang:  req.SourceLang,
		TargetLang:  req.TargetLang,
		OutputDir:   req.OutputDir,
		Voice:       req.Voice,
		Speed:       req.Speed,
		CloneRef:    req.CloneRef,
	}

	if state.SourceLang == "" {
		state.SourceLang = "auto"
	}
	if state.Voice == "" {
		state.Voice = p.config.Defaults.Voice
	}
	if state.Speed == 0 {
		state.Speed = 1.0
	}

	reporter := p.progress
	if reporter == nil {
		reporter = &NoOpProgressReporter{}
	}
	reporter.OnStart(len(p.steps))

	var results []StepResult
	for i, step := range p.steps {
		select {
		case <-ctx.Done():
			return fmt.Errorf("pipeline cancelled: %w", ctx.Err())
		default:
		}

		reporter.OnStepStart(i, step.Name())

		result := p.executeStep(ctx, step, state)
		results = append(results, result)

		reporter.OnStepComplete(i, result)

		if result.Status == StepFailed {
			log.Printf("[Pipeline] 步骤 %q 失败: %v", step.Name(), result.Error)
			p.results = results
			reporter.OnFinish(results, false)
			return fmt.Errorf("step %q failed: %s", step.Name(), result.Error)
		}

		log.Printf("[Pipeline] 步骤 %q 完成 (耗时: %v)", step.Name(), result.Duration)
	}

	reporter.OnFinish(results, true)
	p.results = results
	return nil
}

func (p *Pipeline) executeStep(ctx context.Context, step Step, state *State) StepResult {
	start := time.Now()
	err := step.Run(ctx, state)
	duration := time.Since(start)

	result := StepResult{
		Name:     step.Name(),
		Duration: duration,
	}

	if err != nil {
		result.Status = StepFailed
		result.Error = err.Error()
	} else {
		result.Status = StepCompleted
	}

	return result
}

func ensureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
