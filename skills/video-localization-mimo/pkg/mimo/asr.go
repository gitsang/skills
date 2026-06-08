package mimo

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gitsang/skills/video-localization-mimo/pkg/ffmpeg"
)

const (
	asrEndpoint = "/chat/completions"
	asrModel    = "mimo-v2.5-asr"
)

// Transcribe 将音频文件转换为文本。
func (c *Client) Transcribe(ctx context.Context, audioPath string, language string) (*TranscribeResult, error) {
	log.Printf("[ASR] 开始转录: %s (语言: %s)", audioPath, language)

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("读取音频文件失败: %w", err)
	}

	mimeType := detectMimeType(audioPath)
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)
	audioURI := fmt.Sprintf("data:%s;base64,%s", mimeType, audioBase64)

	req := &ASRRequest{
		Model: asrModel,
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "input_audio",
						InputAudio: &InputAudio{
							Data: audioURI,
						},
					},
				},
			},
		},
		ASROptions: &ASROptions{
			Language: language,
		},
	}

	var resp ASRResponse
	if err := c.doRequest(ctx, asrEndpoint, req, &resp); err != nil {
		return nil, fmt.Errorf("ASR 请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ASR 响应为空")
	}

	text, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("ASR 响应内容类型错误")
	}
	log.Printf("[ASR] 转录完成，文本长度: %d", len(text))

	return &TranscribeResult{
		Text: text,
	}, nil
}

// TranscribeBytes 将音频字节数据转换为文本。
func (c *Client) TranscribeBytes(ctx context.Context, audioData []byte, mimeType string, language string) (*TranscribeResult, error) {
	log.Printf("[ASR] 开始转录字节数据 (类型: %s, 语言: %s)", mimeType, language)

	audioBase64 := base64.StdEncoding.EncodeToString(audioData)
	audioURI := fmt.Sprintf("data:%s;base64,%s", mimeType, audioBase64)

	req := &ASRRequest{
		Model: asrModel,
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "input_audio",
						InputAudio: &InputAudio{
							Data: audioURI,
						},
					},
				},
			},
		},
		ASROptions: &ASROptions{
			Language: language,
		},
	}

	var resp ASRResponse
	if err := c.doRequest(ctx, asrEndpoint, req, &resp); err != nil {
		return nil, fmt.Errorf("ASR 请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ASR 响应为空")
	}

	text, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return nil, fmt.Errorf("ASR 响应内容类型错误")
	}
	log.Printf("[ASR] 转录完成，文本长度: %d", len(text))

	return &TranscribeResult{
		Text: text,
	}, nil
}

// detectMimeType 根据文件扩展名检测 MIME 类型。
func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a":
		return "audio/mp4"
	case ".ogg":
		return "audio/ogg"
	case ".flac":
		return "audio/flac"
	case ".webm":
		return "audio/webm"
	default:
		return "audio/wav"
	}
}

const (
	// MaxAudioSize 是 API 允许的最大音频大小（10MB）
	MaxAudioSize = 10 * 1024 * 1024
	// DefaultSegmentDuration 是默认的音频分段时长（秒）
	DefaultSegmentDuration = 300 // 5分钟
)

// TranscribeWithSplit 转录音频文件，如果文件超过大小限制则自动分段。
func (c *Client) TranscribeWithSplit(ctx context.Context, audioPath string, language string, ffmpegRunner *ffmpeg.Runner) (*TranscribeResult, error) {
	log.Printf("[ASR] 检查音频文件大小: %s", audioPath)

	// 检查文件大小
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("读取音频文件失败: %w", err)
	}

	// 如果文件大小在限制内，直接转录
	if len(audioData) <= MaxAudioSize {
		log.Printf("[ASR] 音频文件大小 %d 字节，在限制内，直接转录", len(audioData))
		return c.Transcribe(ctx, audioPath, language)
	}

	// 需要分段处理
	log.Printf("[ASR] 音频文件大小 %d 字节，超过 %d 字节限制，需要分段处理", len(audioData), MaxAudioSize)

	// 获取音频时长
	mediaInfo, err := ffmpegRunner.Probe(ctx, audioPath)
	if err != nil {
		return nil, fmt.Errorf("获取音频信息失败: %w", err)
	}

	// 计算分段数
	segmentDuration := DefaultSegmentDuration
	numSegments := int(mediaInfo.Duration)/segmentDuration + 1
	log.Printf("[ASR] 音频时长 %.1f 秒，分为 %d 段，每段 %d 秒", mediaInfo.Duration, numSegments, segmentDuration)

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "mimo-asr-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// 分段转录
	var allTexts []string
	for i := 0; i < numSegments; i++ {
		startTime := i * segmentDuration
		segmentPath := filepath.Join(tmpDir, fmt.Sprintf("segment_%03d.wav", i))

		log.Printf("[ASR] 处理分段 %d/%d: %d 秒开始", i+1, numSegments, startTime)

		// 切割音频
		if err := ffmpegRunner.CutAudio(ctx, audioPath, segmentPath, startTime, segmentDuration); err != nil {
			return nil, fmt.Errorf("切割音频分段 %d 失败: %w", i+1, err)
		}

		// 检查分段大小
		segmentData, err := os.ReadFile(segmentPath)
		if err != nil {
			return nil, fmt.Errorf("读取音频分段 %d 失败: %w", i+1, err)
		}

		if len(segmentData) > MaxAudioSize {
			log.Printf("[ASR] 警告：分段 %d 大小 %d 字节仍然超过限制", i+1, len(segmentData))
			// 继续尝试，可能会失败
		}

		// 转录分段
		result, err := c.Transcribe(ctx, segmentPath, language)
		if err != nil {
			return nil, fmt.Errorf("转录音频分段 %d 失败: %w", i+1, err)
		}

		allTexts = append(allTexts, result.Text)
		log.Printf("[ASR] 分段 %d/%d 转录完成，文本长度: %d", i+1, numSegments, len(result.Text))
	}

	// 合并结果
	combinedText := strings.Join(allTexts, "\n\n")
	log.Printf("[ASR] 所有分段转录完成，总文本长度: %d", len(combinedText))

	return &TranscribeResult{
		Text: combinedText,
	}, nil
}
