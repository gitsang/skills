package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "语音克隆",
	Long:  "使用参考音频克隆音色并合成文本",
	RunE:  runClone,
}

func init() {
	cloneCmd.Flags().String("input", "", "输入文本文件路径（必需）")
	cloneCmd.Flags().String("reference", "", "参考音频文件路径（必需）")
	cloneCmd.Flags().String("output", "", "输出音频路径（默认: outputs/clone.wav）")
	_ = cloneCmd.MarkFlagRequired("input")
	_ = cloneCmd.MarkFlagRequired("reference")
}

func runClone(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	referencePath, _ := cmd.Flags().GetString("reference")
	outputPath, _ := cmd.Flags().GetString("output")

	if outputPath == "" {
		outputPath = "outputs/clone.wav"
	}

	inputText, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	if _, err := os.Stat(referencePath); os.IsNotExist(err) {
		return fmt.Errorf("参考音频文件不存在: %s", referencePath)
	}

	client := mimo.NewClient(&mimo.ClientConfig{
		APIKey:  appConfig.MiMo.APIKey,
		BaseURL: appConfig.MiMo.BaseURL,
	})

	progress := NewProgressBar(2, "克隆中")
	progress.Start()

	ctx := context.Background()

	progress.Update(1)
	audioData, err := client.Clone(ctx, &mimo.CloneRequest{
		Text:           string(inputText),
		ReferenceAudio: referencePath,
		Format:         "wav",
	})
	if err != nil {
		return fmt.Errorf("语音克隆失败: %w", err)
	}

	progress.Update(2)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("保存克隆音频失败: %w", err)
	}

	progress.Finish()
	fmt.Printf("克隆音频已保存: %s\n", outputPath)
	return nil
}
