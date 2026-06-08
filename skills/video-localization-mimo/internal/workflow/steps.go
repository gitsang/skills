package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/pkg/ffmpeg"
	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
	"github.com/gitsang/skills/video-localization-mimo/pkg/srt"
)

type Step interface {
	Name() string
	Run(ctx context.Context, state *State) error
}

type ExtractAudioStep struct {
	runner *ffmpeg.Runner
}

func NewExtractAudioStep(runner *ffmpeg.Runner) *ExtractAudioStep {
	return &ExtractAudioStep{runner: runner}
}

func (s *ExtractAudioStep) Name() string { return "提取音频" }

func (s *ExtractAudioStep) Run(ctx context.Context, state *State) error {
	outputPath := filepath.Join(state.OutputDir, "source_audio.wav")

	if err := ensureDir(outputPath); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	if err := s.runner.ExtractAudio(ctx, state.SourceVideo, outputPath); err != nil {
		return fmt.Errorf("extracting audio: %w", err)
	}

	state.SourceAudio = outputPath
	log.Printf("[ExtractAudio] 音频已提取: %s", outputPath)
	return nil
}

type TranscribeStep struct {
	client       *mimo.Client
	ffmpegRunner *ffmpeg.Runner
}

func NewTranscribeStep(client *mimo.Client, ffmpegRunner *ffmpeg.Runner) *TranscribeStep {
	return &TranscribeStep{client: client, ffmpegRunner: ffmpegRunner}
}

func (s *TranscribeStep) Name() string { return "语音转录" }

func (s *TranscribeStep) Run(ctx context.Context, state *State) error {
	result, err := s.client.TranscribeWithSplit(ctx, state.SourceAudio, state.SourceLang, s.ffmpegRunner)
	if err != nil {
		return fmt.Errorf("transcribing audio: %w", err)
	}

	state.SourceText = result.Text

	segments := buildSegmentsFromResult(result, state)
	state.SourceSRT = segments

	txtPath := filepath.Join(state.OutputDir, "source_text.txt")
	if err := os.WriteFile(txtPath, []byte(result.Text), 0644); err != nil {
		return fmt.Errorf("writing source text: %w", err)
	}

	srtPath := filepath.Join(state.OutputDir, "source.srt")
	if err := ensureDir(srtPath); err != nil {
		return err
	}
	if err := srt.GenerateFile(srtPath, segments); err != nil {
		return fmt.Errorf("writing source SRT: %w", err)
	}

	log.Printf("[Transcribe] 转录完成，文本长度: %d, 片段数: %d", len(result.Text), len(segments))
	return nil
}

func buildSegmentsFromResult(result *mimo.TranscribeResult, state *State) []srt.Segment {
	if len(result.Segments) > 0 {
		segments := make([]srt.Segment, len(result.Segments))
		for i, seg := range result.Segments {
			segments[i] = srt.Segment{
				Index: i + 1,
				Start: secondsToTime(seg.Start),
				End:   secondsToTime(seg.End),
				Text:  seg.Text,
			}
		}
		return segments
	}

	lines := splitTextToLines(result.Text, 80)
	segments := make([]srt.Segment, len(lines))
	for i, line := range lines {
		segments[i] = srt.Segment{
			Index: i + 1,
			Start: secondsToTime(float64(i) * 5),
			End:   secondsToTime(float64(i+1) * 5),
			Text:  line,
		}
	}
	return segments
}

func secondsToTime(s float64) time.Time {
	sec := int(s)
	ms := int((s - float64(sec)) * 1000)
	return time.Date(0, 1, 1, sec/3600, (sec%3600)/60, sec%60, ms*1000000, time.UTC)
}

func splitTextToLines(text string, maxLen int) []string {
	runes := []rune(text)
	var lines []string
	for len(runes) > 0 {
		if len(runes) <= maxLen {
			lines = append(lines, string(runes))
			break
		}
		cut := maxLen
		for cut > 0 && runes[cut] != ' ' && runes[cut] != '，' && runes[cut] != '。' {
			cut--
		}
		if cut == 0 {
			cut = maxLen
		}
		lines = append(lines, strings.TrimSpace(string(runes[:cut])))
		runes = runes[cut:]
	}
	return lines
}

type TranslateStep struct {
	client *mimo.Client
}

func NewTranslateStep(client *mimo.Client) *TranslateStep {
	return &TranslateStep{client: client}
}

func (s *TranslateStep) Name() string { return "文本翻译" }

func (s *TranslateStep) Run(ctx context.Context, state *State) error {
	translated, err := s.client.Translate(ctx, &mimo.TranslateRequest{
		Text:       state.SourceText,
		SourceLang: state.SourceLang,
		TargetLang: state.TargetLang,
	})
	if err != nil {
		return fmt.Errorf("translating text: %w", err)
	}

	state.TargetText = translated

	txtPath := filepath.Join(state.OutputDir, "target_text.txt")
	if err := os.WriteFile(txtPath, []byte(translated), 0644); err != nil {
		return fmt.Errorf("writing target text: %w", err)
	}

	log.Printf("[Translate] 翻译完成，文本长度: %d", len(translated))
	return nil
}

type SynthesizeStep struct {
	client *mimo.Client
}

func NewSynthesizeStep(client *mimo.Client) *SynthesizeStep {
	return &SynthesizeStep{client: client}
}

func (s *SynthesizeStep) Name() string { return "语音合成" }

func (s *SynthesizeStep) Run(ctx context.Context, state *State) error {
	outputPath := filepath.Join(state.OutputDir, "target_audio.wav")

	if state.CloneRef != "" {
		return s.cloneVoice(ctx, state, outputPath)
	}
	return s.synthesizeVoice(ctx, state, outputPath)
}

func (s *SynthesizeStep) synthesizeVoice(ctx context.Context, state *State, outputPath string) error {
	if err := ensureDir(outputPath); err != nil {
		return err
	}

	audioData, err := s.client.Synthesize(ctx, &mimo.SynthesizeRequest{
		Text:   state.TargetText,
		Voice:  state.Voice,
		Format: "wav",
	})
	if err != nil {
		return fmt.Errorf("synthesizing speech: %w", err)
	}

	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("writing audio file: %w", err)
	}

	state.TargetAudio = outputPath
	log.Printf("[Synthesize] 合成完成: %s (%d bytes)", outputPath, len(audioData))
	return nil
}

func (s *SynthesizeStep) cloneVoice(ctx context.Context, state *State, outputPath string) error {
	if err := ensureDir(outputPath); err != nil {
		return err
	}

	audioData, err := s.client.Clone(ctx, &mimo.CloneRequest{
		Text:           state.TargetText,
		ReferenceAudio: state.CloneRef,
		Format:         "wav",
	})
	if err != nil {
		return fmt.Errorf("cloning voice: %w", err)
	}

	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("writing cloned audio: %w", err)
	}

	state.TargetAudio = outputPath
	log.Printf("[Synthesize] 克隆合成完成: %s (%d bytes)", outputPath, len(audioData))
	return nil
}

type AlignAudioStep struct {
	runner *ffmpeg.Runner
}

func NewAlignAudioStep(runner *ffmpeg.Runner) *AlignAudioStep {
	return &AlignAudioStep{runner: runner}
}

func (s *AlignAudioStep) Name() string { return "音频对齐" }

func (s *AlignAudioStep) Run(ctx context.Context, state *State) error {
	if len(state.SourceSRT) == 0 {
		log.Printf("[AlignAudio] 无 SRT 片段，跳过对齐")
		return nil
	}

	probe, err := s.runner.Probe(ctx, state.TargetAudio)
	if err != nil {
		return fmt.Errorf("probing target audio: %w", err)
	}

	lastSeg := state.SourceSRT[len(state.SourceSRT)-1]
	requiredDuration := timeToSeconds(lastSeg.End)
	if probe.Duration < requiredDuration {
		log.Printf("[AlignAudio] 目音频时长 (%.1fs) 短于 SRT 时长 (%.1fs)，跳过对齐", probe.Duration, requiredDuration)
		return nil
	}

	log.Printf("[AlignAudio] 音频时长: %.1fs, SRT 时长: %.1fs", probe.Duration, requiredDuration)
	return nil
}

func timeToSeconds(t time.Time) float64 {
	return float64(t.Hour()*3600+t.Minute()*60+t.Second()) + float64(t.Nanosecond())/1e9
}

type GenerateSubtitleStep struct{}

func NewGenerateSubtitleStep() *GenerateSubtitleStep {
	return &GenerateSubtitleStep{}
}

func (s *GenerateSubtitleStep) Name() string { return "生成字幕" }

func (s *GenerateSubtitleStep) Run(ctx context.Context, state *State) error {
	if len(state.SourceSRT) == 0 {
		return fmt.Errorf("no source SRT segments available")
	}

	targetLines := splitTranslationToLines(state.TargetText, len(state.SourceSRT))

	assContent := generateBilingualASS(state.SourceSRT, targetLines)

	assPath := filepath.Join(state.OutputDir, "bilingual.ass")
	if err := ensureDir(assPath); err != nil {
		return err
	}
	if err := os.WriteFile(assPath, []byte(assContent), 0644); err != nil {
		return fmt.Errorf("writing ASS file: %w", err)
	}

	state.BilingualSRT = assPath
	log.Printf("[Subtitle] 双语字幕已生成: %s", assPath)
	return nil
}

func splitTranslationToLines(text string, expectedCount int) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	result := make([]string, 0, expectedCount)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	for len(result) < expectedCount {
		result = append(result, "")
	}
	return result[:expectedCount]
}

func generateBilingualASS(segments []srt.Segment, targetLines []string) string {
	var b strings.Builder

	b.WriteString("[Script Info]\n")
	b.WriteString("Title: Bilingual Subtitles\n")
	b.WriteString("ScriptType: v4.00+\n")
	b.WriteString("PlayResX: 1920\n")
	b.WriteString("PlayResY: 1080\n")
	b.WriteString("WrapStyle: 0\n\n")

	b.WriteString("[V4+ Styles]\n")
	b.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	b.WriteString("Style: Source,Arial,46,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,1,2,10,10,60,1\n")
	b.WriteString("Style: Target,Arial,40,&H0000FFFF,&H000000FF,&H00000000,&H80000000,0,-1,0,0,100,100,0,0,1,2,1,8,10,10,20,1\n\n")

	b.WriteString("[Events]\n")
	b.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	for i, seg := range segments {
		start := formatASS(seg.Start)
		end := formatASS(seg.End)
		sourceText := strings.ReplaceAll(seg.Text, "\n", "\\N")

		b.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Source,,0,0,0,,%s\n", start, end, sourceText))

		if i < len(targetLines) && targetLines[i] != "" {
			target := strings.ReplaceAll(targetLines[i], "\n", "\\N")
			b.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Target,,0,0,0,,%s\n", start, end, target))
		}
	}

	return b.String()
}

func formatASS(t time.Time) string {
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	cs := t.Nanosecond() / 10000000
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
}

type ComposeStep struct {
	runner *ffmpeg.Runner
}

func NewComposeStep(runner *ffmpeg.Runner) *ComposeStep {
	return &ComposeStep{runner: runner}
}

func (s *ComposeStep) Name() string { return "合成视频" }

func (s *ComposeStep) Run(ctx context.Context, state *State) error {
	outputPath := filepath.Join(state.OutputDir, "final.mp4")

	if err := ensureDir(outputPath); err != nil {
		return err
	}

	opts := ffmpeg.ComposeOptions{
		VideoPath:    state.SourceVideo,
		AudioPath:    state.TargetAudio,
		SubtitlePath: state.BilingualSRT,
		OutputPath:   outputPath,
		AudioVolume:  1.0,
	}

	if err := s.runner.Compose(ctx, opts); err != nil {
		return fmt.Errorf("composing video: %w", err)
	}

	state.FinalVideo = outputPath
	log.Printf("[Compose] 最终视频已生成: %s", outputPath)
	return nil
}
