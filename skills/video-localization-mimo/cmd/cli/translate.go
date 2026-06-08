package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
	"github.com/spf13/cobra"
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "文本翻译",
	Long:  "使用 MiMo LLM 将文本从源语言翻译到目标语言",
	RunE:  runTranslate,
}

func init() {
	translateCmd.Flags().String("input", "", "输入文本文件路径（必需）")
	translateCmd.Flags().String("output", "", "输出文件路径（默认: outputs/translate.txt）")
	translateCmd.Flags().String("target-lang", "en", "目标语言（zh/en/ja/ko）")
	translateCmd.Flags().String("source-lang", "", "源语言（默认: 自动检测）")
	translateCmd.Flags().String("system-prompt", "", "自定义系统提示词")
	_ = translateCmd.MarkFlagRequired("input")
}

func runTranslate(cmd *cobra.Command, args []string) error {
	inputPath, _ := cmd.Flags().GetString("input")
	outputPath, _ := cmd.Flags().GetString("output")
	targetLang, _ := cmd.Flags().GetString("target-lang")
	sourceLang, _ := cmd.Flags().GetString("source-lang")
	systemPrompt, _ := cmd.Flags().GetString("system-prompt")

	if outputPath == "" {
		outputPath = "outputs/translate.txt"
	}

	inputText, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}

	if sourceLang == "" {
		sourceLang = appConfig.Defaults.SourceLang
	}

	client := mimo.NewClient(&mimo.ClientConfig{
		APIKey:  appConfig.MiMo.APIKey,
		BaseURL: appConfig.MiMo.BaseURL,
	})

	progress := NewProgressBar(2, "翻译中")
	progress.Start()

	ctx := context.Background()

	progress.Update(1)
	translated, err := client.Translate(ctx, &mimo.TranslateRequest{
		Text:         string(inputText),
		SourceLang:   sourceLang,
		TargetLang:   targetLang,
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		return fmt.Errorf("翻译失败: %w", err)
	}

	progress.Update(2)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(translated), 0644); err != nil {
		return fmt.Errorf("保存翻译结果失败: %w", err)
	}

	progress.Finish()
	fmt.Printf("翻译结果已保存: %s\n", outputPath)
	return nil
}
