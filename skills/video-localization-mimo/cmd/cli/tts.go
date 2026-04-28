package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
	"github.com/spf13/cobra"
)

var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "文本转语音",
	Long:  "使用 MiMo TTS 将文本转换为语音音频文件",
	RunE:  runTTS,
}

func init() {
	ttsCmd.Flags().String("input", "", "输入文本文件路径（必需）")
	ttsCmd.Flags().String("output", "", "输出音频路径（默认: outputs/tts.wav）")
	ttsCmd.Flags().String("voice", "", "音色名称（默认: Chloe）")
	ttsCmd.Flags().Float64("speed", 1.0, "语速（0.5-2.0）")
	_ = ttsCmd.MarkFlagRequired("input")
}

func runTTS(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	voice, _ := cmd.Flags().GetString("voice")

	if outputPath == "" {
		outputPath = "outputs/tts.wav"
	}

	inputText, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	if voice == "" {
		voice = appConfig.Defaults.Voice
	}

	client := mimo.NewClient(&mimo.ClientConfig{
		APIKey:  appConfig.MiMo.APIKey,
		BaseURL: appConfig.MiMo.BaseURL,
	})

	progress := NewProgressBar(2, "合成中")
	progress.Start()

	ctx := context.Background()

	progress.Update(1)
	audioData, err := client.Synthesize(ctx, &mimo.SynthesizeRequest{
		Text:   string(inputText),
		Voice:  voice,
		Format: "wav",
	})
	if err != nil {
		return fmt.Errorf("语音合成失败: %w", err)
	}

	progress.Update(2)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("保存音频文件失败: %w", err)
	}

	progress.Finish()
	fmt.Printf("音频已保存: %s\n", outputPath)
	return nil
}
