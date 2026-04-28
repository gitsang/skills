package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/pkg/srt"
	"github.com/spf13/cobra"
)

var subtitleCmd = &cobra.Command{
	Use:   "subtitle",
	Short: "生成双语字幕",
	Long:  "根据源 SRT 和目标翻译文本，生成双语 ASS 字幕文件",
	RunE:  runSubtitle,
}

func init() {
	subtitleCmd.Flags().String("source-srt", "", "源 SRT 文件路径（必需）")
	subtitleCmd.Flags().String("target-text", "", "目标翻译文本路径（必需）")
	subtitleCmd.Flags().String("output", "", "输出 ASS 文件路径（默认: outputs/bilingual.ass）")
	subtitleCmd.Flags().String("style", "default", "字幕样式（default/modern/classic）")
	_ = subtitleCmd.MarkFlagRequired("source-srt")
	_ = subtitleCmd.MarkFlagRequired("target-text")
}

func runSubtitle(cmd *cobra.Command, args []string) error {
	sourceSRT, _ := cmd.Flags().GetString("source-srt")
	targetTextPath, _ := cmd.Flags().GetString("target-text")
	outputPath, _ := cmd.Flags().GetString("output")
	style, _ := cmd.Flags().GetString("style")

	if outputPath == "" {
		outputPath = "outputs/bilingual.ass"
	}

	segments, err := srt.ParseFile(sourceSRT)
	if err != nil {
		return fmt.Errorf("解析源 SRT 文件失败: %w", err)
	}

	targetText, err := os.ReadFile(targetTextPath)
	if err != nil {
		return fmt.Errorf("读取目标文本失败: %w", err)
	}

	targetLines := splitTargetText(string(targetText), len(segments))

	assContent := generateASS(segments, targetLines, style)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(assContent), 0644); err != nil {
		return fmt.Errorf("保存 ASS 文件失败: %w", err)
	}

	fmt.Printf("双语字幕已保存: %s\n", outputPath)
	return nil
}

func splitTargetText(text string, expectedCount int) []string {
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

func generateASS(segments []srt.Segment, targetLines []string, style string) string {
	var b strings.Builder

	b.WriteString("[Script Info]\n")
	b.WriteString("Title: Bilingual Subtitles\n")
	b.WriteString("ScriptType: v4.00+\n")
	b.WriteString("PlayResX: 1920\n")
	b.WriteString("PlayResY: 1080\n")
	b.WriteString("WrapStyle: 0\n")
	b.WriteString("\n")

	b.WriteString("[V4+ Styles]\n")
	b.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")

	switch style {
	case "modern":
		b.WriteString("Style: Source,Arial,48,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,1,2,10,10,60,1\n")
		b.WriteString("Style: Target,Arial,42,&H0000FFFF,&H000000FF,&H00000000,&H80000000,0,-1,0,0,100,100,0,0,1,2,1,8,10,10,20,1\n")
	case "classic":
		b.WriteString("Style: Source,SimSun,40,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,1,2,10,10,60,1\n")
		b.WriteString("Style: Target,SimSun,36,&H0000FFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,1,8,10,10,20,1\n")
	default:
		b.WriteString("Style: Source,Arial,46,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,0,0,100,100,0,0,1,2,1,2,10,10,60,1\n")
		b.WriteString("Style: Target,Arial,40,&H0000FFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,1,8,10,10,20,1\n")
	}

	b.WriteString("\n[Events]\n")
	b.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	for i, seg := range segments {
		start := formatASSTime(seg.Start)
		end := formatASSTime(seg.End)
		sourceText := strings.ReplaceAll(seg.Text, "\n", "\\N")

		b.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Source,,0,0,0,,%s\n", start, end, sourceText))

		if i < len(targetLines) && targetLines[i] != "" {
			target := strings.ReplaceAll(targetLines[i], "\n", "\\N")
			b.WriteString(fmt.Sprintf("Dialogue: 0,%s,%s,Target,,0,0,0,,%s\n", start, end, target))
		}
	}

	return b.String()
}

func formatASSTime(t time.Time) string {
	h := t.Hour()
	m := t.Minute()
	s := t.Second()
	cs := t.Nanosecond() / 10000000
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
}
