package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
	"github.com/gitsang/skills/video-localization-mimo/pkg/srt"
	"github.com/spf13/cobra"
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe",
	Short: "音频转文本并生成 SRT 字幕",
	Long:  "使用 MiMo ASR 将音频文件转换为文本，并生成 SRT 字幕文件",
	RunE:  runTranscribe,
}

func init() {
	transcribeCmd.Flags().String("audio", "", "音频文件路径（必需）")
	transcribeCmd.Flags().String("lang", "auto", "语言（auto/zh/en）")
	transcribeCmd.Flags().String("txt-out", "", "文本输出路径（默认: outputs/transcribe.txt）")
	transcribeCmd.Flags().String("srt-out", "", "SRT 输出路径（默认: outputs/transcribe.srt）")
	_ = transcribeCmd.MarkFlagRequired("audio")
}

func runTranscribe(cmd *cobra.Command, args []string) error {
	audioPath, _ := cmd.Flags().GetString("audio")
	lang, _ := cmd.Flags().GetString("lang")
	txtOut, _ := cmd.Flags().GetString("txt-out")
	srtOut, _ := cmd.Flags().GetString("srt-out")

	if txtOut == "" {
		txtOut = "outputs/transcribe.txt"
	}
	if srtOut == "" {
		srtOut = "outputs/transcribe.srt"
	}

	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return fmt.Errorf("音频文件不存在: %s", audioPath)
	}

	client := mimo.NewClient(&mimo.ClientConfig{
		APIKey:  appConfig.MiMo.APIKey,
		BaseURL: appConfig.MiMo.BaseURL,
	})

	progress := NewProgressBar(3, "转录中")
	progress.Start()

	ctx := context.Background()

	progress.Update(1)
	result, err := client.Transcribe(ctx, audioPath, lang)
	if err != nil {
		return fmt.Errorf("转录失败: %w", err)
	}

	progress.Update(2)
	if err := os.MkdirAll(filepath.Dir(txtOut), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(txtOut, []byte(result.Text), 0644); err != nil {
		return fmt.Errorf("保存文本文件失败: %w", err)
	}

	progress.Update(3)
	if err := os.MkdirAll(filepath.Dir(srtOut), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	segments := buildSegments(result)
	if err := srt.GenerateFile(srtOut, segments); err != nil {
		return fmt.Errorf("保存 SRT 文件失败: %w", err)
	}

	progress.Finish()
	fmt.Printf("文本已保存: %s\n", txtOut)
	fmt.Printf("SRT 已保存: %s\n", srtOut)
	return nil
}

func buildSegments(result *mimo.TranscribeResult) []srt.Segment {
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
