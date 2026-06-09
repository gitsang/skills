package workflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	if len(state.SourceSRT) == 0 {
		return fmt.Errorf("no source SRT segments available")
	}

	segments := make([]string, len(state.SourceSRT))
	for i, seg := range state.SourceSRT {
		translated, err := s.client.Translate(ctx, &mimo.TranslateRequest{
			Text:       seg.Text,
			SourceLang: state.SourceLang,
			TargetLang: state.TargetLang,
		})
		if err != nil {
			return fmt.Errorf("translating segment %d: %w", i+1, err)
		}

		segments[i] = translated
		log.Printf("[Translate] 片段 %d/%d 翻译完成: %s -> %s", i+1, len(state.SourceSRT), seg.Text, translated)
	}

	state.TargetSegments = segments

	// Also set the full translated text for backward compatibility
	state.TargetText = strings.Join(segments, "\n")

	txtPath := filepath.Join(state.OutputDir, "target_text.txt")
	if err := os.WriteFile(txtPath, []byte(state.TargetText), 0644); err != nil {
		return fmt.Errorf("writing target text: %w", err)
	}

	log.Printf("[Translate] 翻译完成，共 %d 个片段", len(segments))
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
	if len(state.TargetSegments) == 0 {
		return fmt.Errorf("no target segments available")
	}

	outputDir := filepath.Join(state.OutputDir, "segments")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating segments directory: %w", err)
	}

	audioSegments := make([]string, len(state.TargetSegments))
	for i, text := range state.TargetSegments {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("segment_%03d.wav", i+1))

		if state.CloneRef != "" {
			if err := s.cloneVoiceSegment(ctx, state, text, outputPath); err != nil {
				return fmt.Errorf("cloning voice for segment %d: %w", i+1, err)
			}
		} else {
			if err := s.synthesizeVoiceSegment(ctx, state, text, outputPath); err != nil {
				return fmt.Errorf("synthesizing segment %d: %w", i+1, err)
			}
		}

		audioSegments[i] = outputPath
		log.Printf("[Synthesize] 片段 %d/%d 合成完成: %s", i+1, len(state.TargetSegments), outputPath)
	}

	state.TargetAudioSegments = audioSegments

	// Also create a combined audio file for backward compatibility
	combinedPath := filepath.Join(state.OutputDir, "target_audio.wav")
	if err := s.combineAudioSegments(ctx, audioSegments, combinedPath); err != nil {
		return fmt.Errorf("combining audio segments: %w", err)
	}
	state.TargetAudio = combinedPath

	log.Printf("[Synthesize] 合成完成，共 %d 个片段", len(audioSegments))
	return nil
}

func (s *SynthesizeStep) synthesizeVoiceSegment(ctx context.Context, state *State, text, outputPath string) error {
	if err := ensureDir(outputPath); err != nil {
		return err
	}

	audioData, err := s.client.Synthesize(ctx, &mimo.SynthesizeRequest{
		Text:   text,
		Voice:  state.Voice,
		Format: "wav",
	})
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, audioData, 0644)
}

func (s *SynthesizeStep) cloneVoiceSegment(ctx context.Context, state *State, text, outputPath string) error {
	if err := ensureDir(outputPath); err != nil {
		return err
	}

	audioData, err := s.client.Clone(ctx, &mimo.CloneRequest{
		Text:           text,
		ReferenceAudio: state.CloneRef,
		Format:         "wav",
	})
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, audioData, 0644)
}

func (s *SynthesizeStep) combineAudioSegments(ctx context.Context, segments []string, outputPath string) error {
	if len(segments) == 0 {
		return fmt.Errorf("no audio segments to combine")
	}

	// Create a file list for FFmpeg concat
	listPath := outputPath + ".txt"
	var listContent strings.Builder
	for _, seg := range segments {
		listContent.WriteString(fmt.Sprintf("file '%s'\n", seg))
	}
	if err := os.WriteFile(listPath, []byte(listContent.String()), 0644); err != nil {
		return fmt.Errorf("writing concat list: %w", err)
	}
	defer os.Remove(listPath)

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("combining audio: %w\noutput: %s", err, string(output))
	}

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
	if len(state.SourceSRT) == 0 || len(state.TargetAudioSegments) == 0 {
		log.Printf("[AlignAudio] 无 SRT 片段或音频片段，跳过对齐")
		return nil
	}

	alignedDir := filepath.Join(state.OutputDir, "aligned")
	if err := os.MkdirAll(alignedDir, 0755); err != nil {
		return fmt.Errorf("creating aligned directory: %w", err)
	}

	alignedSegments := make([]string, len(state.TargetAudioSegments))

	for i, audioPath := range state.TargetAudioSegments {
		sourceSeg := state.SourceSRT[i]
		sourceDuration := timeToSeconds(sourceSeg.End) - timeToSeconds(sourceSeg.Start)

		probe, err := s.runner.Probe(ctx, audioPath)
		if err != nil {
			return fmt.Errorf("probing audio segment %d: %w", i+1, err)
		}
		targetDuration := probe.Duration

		alignedPath := filepath.Join(alignedDir, fmt.Sprintf("aligned_%03d.wav", i+1))

		if targetDuration <= sourceDuration {
			// Case 1: Target is shorter than source - pad with silence
			if err := s.padAudio(ctx, audioPath, alignedPath, sourceDuration); err != nil {
				return fmt.Errorf("padding audio segment %d: %w", i+1, err)
			}
			log.Printf("[AlignAudio] 片段 %d: 目标 (%.1fs) < 源 (%.1fs)，填充静音", i+1, targetDuration, sourceDuration)
		} else if targetDuration <= sourceDuration*1.5 {
			// Case 2: Target is longer but within 1.5x - speed up
			speedFactor := targetDuration / sourceDuration
			if err := s.speedUpAudio(ctx, audioPath, alignedPath, speedFactor); err != nil {
				return fmt.Errorf("speeding up audio segment %d: %w", i+1, err)
			}
			log.Printf("[AlignAudio] 片段 %d: 目标 (%.1fs) > 源 (%.1fs)，加速 %.2fx", i+1, targetDuration, sourceDuration, speedFactor)
		} else {
			// Case 3: Target is much longer - speed up to 1.5x and adjust timing
			speedFactor := 1.5
			newDuration := targetDuration / speedFactor
			if err := s.speedUpAudio(ctx, audioPath, alignedPath, speedFactor); err != nil {
				return fmt.Errorf("speeding up audio segment %d: %w", i+1, err)
			}
			log.Printf("[AlignAudio] 片段 %d: 目标 (%.1fs) >> 源 (%.1fs)，加速 1.5x，新时长 %.1fs", i+1, targetDuration, sourceDuration, newDuration)
		}

		alignedSegments[i] = alignedPath
	}

	// Combine aligned segments into final audio
	combinedPath := filepath.Join(state.OutputDir, "aligned_audio.wav")
	if err := s.combineAlignedSegments(ctx, alignedSegments, combinedPath); err != nil {
		return fmt.Errorf("combining aligned segments: %w", err)
	}

	state.TargetAudio = combinedPath
	log.Printf("[AlignAudio] 音频对齐完成: %s", combinedPath)
	return nil
}

func (s *AlignAudioStep) padAudio(ctx context.Context, input, output string, targetDuration float64) error {
	args := []string{
		"-i", input,
		"-af", fmt.Sprintf("apad=pad_dur=%.3f", targetDuration),
		"-t", fmt.Sprintf("%.3f", targetDuration),
		"-y",
		output,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("padding audio: %w\noutput: %s", err, string(output))
	}

	return nil
}

func (s *AlignAudioStep) speedUpAudio(ctx context.Context, input, output string, speedFactor float64) error {
	args := []string{
		"-i", input,
		"-af", fmt.Sprintf("atempo=%.3f", speedFactor),
		"-y",
		output,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("speeding up audio: %w\noutput: %s", err, string(output))
	}

	return nil
}

func (s *AlignAudioStep) combineAlignedSegments(ctx context.Context, segments []string, outputPath string) error {
	if len(segments) == 0 {
		return fmt.Errorf("no aligned segments to combine")
	}

	listPath := outputPath + ".txt"
	var listContent strings.Builder
	for _, seg := range segments {
		listContent.WriteString(fmt.Sprintf("file '%s'\n", seg))
	}
	if err := os.WriteFile(listPath, []byte(listContent.String()), 0644); err != nil {
		return fmt.Errorf("writing concat list: %w", err)
	}
	defer os.Remove(listPath)

	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", listPath,
		"-c", "copy",
		"-y",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("combining aligned segments: %w\noutput: %s", err, string(output))
	}

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

	if len(state.TargetSegments) == 0 {
		return fmt.Errorf("no target segments available")
	}

	assContent := generateBilingualASS(state.SourceSRT, state.TargetSegments)

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
