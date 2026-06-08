package mimo

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
