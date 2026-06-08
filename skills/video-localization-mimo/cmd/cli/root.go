package cli

import (
	"fmt"
	"os"

	"github.com/gitsang/skills/video-localization-mimo/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	verbose  bool
	appConfig *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "video-localization-mimo",
	Short: "视频本地化工具",
	Long:  "基于 MiMo API 的视频本地化工具，支持语音识别、翻译、语音合成等功能",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}
		appConfig = cfg
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")

	rootCmd.AddCommand(transcribeCmd)
	rootCmd.AddCommand(translateCmd)
	rootCmd.AddCommand(ttsCmd)
	rootCmd.AddCommand(cloneCmd)
	rootCmd.AddCommand(subtitleCmd)
	rootCmd.AddCommand(composeCmd)
	rootCmd.AddCommand(localizeCmd)
}