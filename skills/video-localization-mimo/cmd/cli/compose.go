package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/gitsang/skills/video-localization-mimo/pkg/ffmpeg"
	"github.com/spf13/cobra"
)

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "合成最终视频",
	Long:  "将视频、音频和字幕合成为最终视频文件",
	RunE:  runCompose,
}

func init() {
	composeCmd.Flags().String("video", "", "输入视频路径（必需）")
	composeCmd.Flags().String("audio", "", "替换音频路径")
	composeCmd.Flags().String("subtitle", "", "字幕文件路径（SRT/ASS）")
	composeCmd.Flags().String("output", "", "输出视频路径（默认: outputs/composed.mp4）")
	composeCmd.Flags().Float64("audio-volume", 1.0, "音频音量（0.0-2.0）")
	_ = composeCmd.MarkFlagRequired("video")
}

func runCompose(cmd *cobra.Command, args []string) error {
	videoPath, _ := cmd.Flags().GetString("video")
	audioPath, _ := cmd.Flags().GetString("audio")
	subtitlePath, _ := cmd.Flags().GetString("subtitle")
	outputPath, _ := cmd.Flags().GetString("output")
	audioVolume, _ := cmd.Flags().GetFloat64("audio-volume")

	if outputPath == "" {
		outputPath = "outputs/composed.mp4"
	}

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	ffmpegPath := appConfig.FFmpeg.Path
	ffprobePath := appConfig.FFmpeg.FFprobePath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if ffprobePath == "" {
		ffprobePath = "ffprobe"
	}

	runner := ffmpeg.NewRunner(ffmpegPath, ffprobePath)

	progress := NewProgressBar(2, "合成中")
	progress.Start()

	ctx := context.Background()

	progress.Update(1)
	if err := os.MkdirAll("outputs", 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	opts := ffmpeg.ComposeOptions{
		VideoPath:    videoPath,
		AudioPath:    audioPath,
		SubtitlePath: subtitlePath,
		OutputPath:   outputPath,
		AudioVolume:  audioVolume,
	}

	if err := runner.Compose(ctx, opts); err != nil {
		return fmt.Errorf("合成视频失败: %w", err)
	}

	progress.Finish()
	fmt.Printf("视频已保存: %s\n", outputPath)
	return nil
}
