package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/workflow"
	"github.com/spf13/cobra"
)

var localizeCmd = &cobra.Command{
	Use:   "localize",
	Short: "完整视频本地化",
	Long:  "执行完整的视频本地化流程：提取音频 → 语音转录 → 文本翻译 → 语音合成 → 音频对齐 → 生成字幕 → 合成视频",
	RunE:  runLocalize,
}

func init() {
	localizeCmd.Flags().String("video", "", "源视频路径（必需）")
	localizeCmd.Flags().String("source-lang", "auto", "源语言（auto/zh/en）")
	localizeCmd.Flags().String("target-lang", "zh", "目标语言（zh/en/ja/ko）")
	localizeCmd.Flags().String("output-dir", "outputs/", "输出目录")
	localizeCmd.Flags().String("voice", "", "TTS 音色（可选，不指定则使用默认）")
	localizeCmd.Flags().String("clone-ref", "", "克隆参考音频路径（可选）")
	localizeCmd.Flags().Bool("clone", false, "从视频自动提取声音进行克隆")
	localizeCmd.Flags().Float64("speed", 1.0, "语速（0.5-2.0）")
	localizeCmd.Flags().Duration("timeout", 600*time.Second, "API 请求超时时间")
	_ = localizeCmd.MarkFlagRequired("video")
}

func runLocalize(cmd *cobra.Command, args []string) error {
	videoPath, _ := cmd.Flags().GetString("video")
	sourceLang, _ := cmd.Flags().GetString("source-lang")
	targetLang, _ := cmd.Flags().GetString("target-lang")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	voice, _ := cmd.Flags().GetString("voice")
	cloneRef, _ := cmd.Flags().GetString("clone-ref")
	clone, _ := cmd.Flags().GetBool("clone")
	speed, _ := cmd.Flags().GetFloat64("speed")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	if cloneRef != "" {
		if _, err := os.Stat(cloneRef); os.IsNotExist(err) {
			return fmt.Errorf("参考音频文件不存在: %s", cloneRef)
		}
	}

	if outputDir == "outputs/" {
		base := filepath.Base(videoPath)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		outputDir = filepath.Join("outputs", name)
	}

	// 如果指定 --clone，从视频提取前30秒音频作为参考
	if clone && cloneRef == "" {
		cloneRef = filepath.Join(outputDir, "clone_ref.wav")
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("创建输出目录失败: %w", err)
		}
		fmt.Printf("🎤 从视频提取前30秒音频作为克隆参考...\n")
		// 使用 ffmpeg 提取前30秒音频
		ffmpegPath := appConfig.FFmpeg.Path
		if ffmpegPath == "" {
			ffmpegPath = "ffmpeg"
		}
		cmd := exec.Command(ffmpegPath, "-i", videoPath, "-t", "30", "-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", "-y", cloneRef)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("提取参考音频失败: %w\n%s", err, output)
		}
		fmt.Printf("   参考音频已保存: %s\n", cloneRef)
	}

	req := &workflow.Request{
		SourceVideo: videoPath,
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
		OutputDir:   outputDir,
		Voice:       voice,
		Speed:       speed,
		CloneRef:    cloneRef,
	}

	pipeline := workflow.NewPipelineWithTimeout(appConfig, timeout)

	progressReporter := workflow.NewCLIProgressReporter()
	pipeline.SetProgressReporter(progressReporter)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("🎬 开始视频本地化\n")
	fmt.Printf("   源视频: %s\n", videoPath)
	fmt.Printf("   源语言: %s → 目标语言: %s\n", sourceLang, targetLang)
	fmt.Printf("   输出目录: %s\n", outputDir)
	fmt.Printf("   超时时间: %s\n", timeout)
	if cloneRef != "" {
		fmt.Printf("   克隆参考: %s\n", cloneRef)
	}
	fmt.Println()

	err := pipeline.Run(ctx, req)

	results := pipeline.GetResults()
	report := workflow.GenerateReport(results, err == nil)
	reportPath := filepath.Join(outputDir, "report.json")
	if saveErr := workflow.SaveReport(reportPath, report); saveErr != nil {
		fmt.Fprintf(os.Stderr, "警告: 保存报告失败: %v\n", saveErr)
	}

	if err != nil {
		return fmt.Errorf("视频本地化失败: %w", err)
	}

	fmt.Printf("\n📁 输出文件:\n")
	fmt.Printf("   字幕: %s\n", filepath.Join(outputDir, "bilingual.ass"))
	fmt.Printf("   视频: %s\n", filepath.Join(outputDir, "final.mp4"))
	fmt.Printf("   报告: %s\n", reportPath)

	return nil
}
