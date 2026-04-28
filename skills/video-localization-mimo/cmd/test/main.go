package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/config"
	"github.com/gitsang/skills/video-localization-mimo/pkg/mimo"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== MiMo API 客户端测试 ===")

	// 1. 加载配置
	log.Println("\n[1] 加载配置...")
	cfg, err := config.Load("configs/mimo.yml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("✓ 配置加载成功")
	log.Printf("  API Key: %s...", cfg.MiMo.APIKey[:10])
	log.Printf("  Base URL: %s", cfg.MiMo.BaseURL)
	log.Printf("  Timeout: %v", cfg.MiMo.Timeout)

	// 2. 初始化客户端
	log.Println("\n[2] 初始化 MiMo 客户端...")
	clientConfig := &mimo.ClientConfig{
		APIKey:  cfg.MiMo.APIKey,
		BaseURL: cfg.MiMo.BaseURL,
		Timeout: cfg.MiMo.Timeout,
	}
	client := mimo.NewClient(clientConfig)
	log.Printf("✓ 客户端初始化成功")
	log.Printf("  配置: %+v", client.GetConfig())

	// 3. 测试 ASR API
	log.Println("\n[3] 测试 ASR API...")

	// 查找音频文件
	audioPath := findAudioFile()
	if audioPath == "" {
		log.Println("⚠ 未找到音频文件，跳过 ASR 测试")
		log.Println("  提示: 使用 ffmpeg 从视频中提取音频:")
		log.Println("  ffmpeg -i inputs/video.mkv -vn -acodec pcm_s16le -ar 16000 -ac 1 outputs/audio.wav")
	} else {
		log.Printf("  使用音频文件: %s", audioPath)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		result, err := client.Transcribe(ctx, audioPath, "auto")
		if err != nil {
			log.Printf("✗ ASR 请求失败: %v", err)
		} else {
			log.Printf("✓ ASR 请求成功")
			log.Printf("  识别文本: %s", truncate(result.Text, 200))
		}
	}

	// 4. 测试 LLM API (简单翻译)
	log.Println("\n[4] 测试 LLM API...")
	testLLM(client)

	log.Println("\n=== 测试完成 ===")
}

func findAudioFile() string {
	// 检查 outputs 目录下的音频文件
	outputDir := "outputs"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return ""
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".wav" || ext == ".mp3" {
			return filepath.Join(outputDir, entry.Name())
		}
	}

	return ""
}

func extractAudio(videoPath string) (string, error) {
	outputDir := "outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	baseName := filepath.Base(videoPath)
	ext := filepath.Ext(baseName)
	audioName := baseName[:len(baseName)-len(ext)] + ".wav"
	audioPath := filepath.Join(outputDir, audioName)

	if _, err := os.Stat(audioPath); err == nil {
		return audioPath, nil
	}

	log.Printf("  提取音频: %s -> %s", videoPath, audioPath)
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		audioPath,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg 执行失败: %w", err)
	}

	return audioPath, nil
}

func testLLM(client *mimo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	translated, err := client.Translate(ctx, &mimo.TranslateRequest{
		Text:       "你好世界，这是一个测试。",
		SourceLang: "zh",
		TargetLang: "en",
	})
	if err != nil {
		log.Printf("✗ LLM 翻译失败: %v", err)
		return
	}

	log.Printf("✓ LLM 翻译成功")
	log.Printf("  原文: 你好世界，这是一个测试。")
	log.Printf("  译文: %s", translated)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
