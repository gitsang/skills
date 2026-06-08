package mimo

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
)

const (
	ttsEndpoint      = "/chat/completions"
	ttsModel         = "mimo-v2.5-tts"
	ttsCloneModel    = "mimo-v2.5-tts-voiceclone"
	defaultVoice     = "Chloe"
	defaultFormat    = "wav"
)

// Synthesize 将文本转换为音频。
func (c *Client) Synthesize(ctx context.Context, req *SynthesizeRequest) ([]byte, error) {
	log.Printf("[TTS] 开始合成: %s (音色: %s)", truncateText(req.Text, 50), req.Voice)

	if req.Voice == "" {
		req.Voice = defaultVoice
	}
	if req.Format == "" {
		req.Format = defaultFormat
	}

	messages := []Message{
		{
			Role:    "assistant",
			Content: req.Text,
		},
	}

	if req.Style != "" {
		messages = append([]Message{
			{
				Role:    "user",
				Content: req.Style,
			},
		}, messages...)
	}

	ttsReq := &TTSRequest{
		Model:    ttsModel,
		Messages: messages,
		Audio: Audio{
			Format: req.Format,
			Voice:  req.Voice,
		},
	}

	var resp TTSResponse
	if err := c.doRequest(ctx, ttsEndpoint, ttsReq, &resp); err != nil {
		return nil, fmt.Errorf("TTS 请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("TTS 响应为空")
	}

	audioBase64 := resp.Choices[0].Message.Audio.Data
	audioData, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		return nil, fmt.Errorf("解码音频数据失败: %w", err)
	}

	log.Printf("[TTS] 合成完成，音频大小: %d bytes", len(audioData))
	return audioData, nil
}

// Clone 使用参考音频克隆音色并合成文本。
func (c *Client) Clone(ctx context.Context, req *CloneRequest) ([]byte, error) {
	log.Printf("[TTS] 开始克隆合成: %s (参考音频: %s)", truncateText(req.Text, 50), req.ReferenceAudio)

	if req.Format == "" {
		req.Format = defaultFormat
	}

	refData, err := os.ReadFile(req.ReferenceAudio)
	if err != nil {
		return nil, fmt.Errorf("读取参考音频失败: %w", err)
	}

	refBase64 := base64.StdEncoding.EncodeToString(refData)
	mimeType := detectMimeType(req.ReferenceAudio)
	refURI := fmt.Sprintf("data:%s;base64,%s", mimeType, refBase64)

	ttsReq := &TTSRequest{
		Model: ttsCloneModel,
		Messages: []Message{
			{
				Role:    "assistant",
				Content: req.Text,
			},
		},
		Audio: Audio{
			Format: req.Format,
			Voice:  refURI,
		},
	}

	var resp TTSResponse
	if err := c.doRequest(ctx, ttsEndpoint, ttsReq, &resp); err != nil {
		return nil, fmt.Errorf("TTS 克隆请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("TTS 克隆响应为空")
	}

	audioBase64 := resp.Choices[0].Message.Audio.Data
	audioData, err := base64.StdEncoding.DecodeString(audioBase64)
	if err != nil {
		return nil, fmt.Errorf("解码音频数据失败: %w", err)
	}

	log.Printf("[TTS] 克隆合成完成，音频大小: %d bytes", len(audioData))
	return audioData, nil
}

// SynthesizeToFile 将文本转换为音频并保存到文件。
func (c *Client) SynthesizeToFile(ctx context.Context, req *SynthesizeRequest, outputPath string) error {
	audioData, err := c.Synthesize(ctx, req)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("保存音频文件失败: %w", err)
	}

	log.Printf("[TTS] 音频已保存: %s", outputPath)
	return nil
}

// CloneToFile 使用参考音频克隆音色并保存到文件。
func (c *Client) CloneToFile(ctx context.Context, req *CloneRequest, outputPath string) error {
	audioData, err := c.Clone(ctx, req)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, audioData, 0644); err != nil {
		return fmt.Errorf("保存音频文件失败: %w", err)
	}

	log.Printf("[TTS] 克隆音频已保存: %s", outputPath)
	return nil
}

// truncateText 截断文本用于日志显示。
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
