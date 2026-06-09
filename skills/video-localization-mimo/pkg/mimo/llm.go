package mimo

import (
	"context"
	"fmt"
	"log"
)

const (
	llmEndpoint = "/chat/completions"
	llmModel    = "mimo-v2.5-pro"
)

// Translate 将文本从源语言翻译到目标语言。
func (c *Client) Translate(ctx context.Context, req *TranslateRequest) (string, error) {
	log.Printf("[LLM] 开始翻译: %s -> %s", req.SourceLang, req.TargetLang)

	systemPrompt := req.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = fmt.Sprintf(
			"You are a professional translator. Translate the following text from %s to %s. "+
				"Only return the translated text, no explanations.",
			req.SourceLang, req.TargetLang,
		)
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = 0.3
	}

	llmReq := &LLMRequest{
		Model: llmModel,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: req.Text,
			},
		},
		Temperature: temperature,
	}

	var resp LLMResponse
	if err := c.doRequest(ctx, llmEndpoint, llmReq, &resp); err != nil {
		return "", fmt.Errorf("LLM 请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM 响应为空")
	}

	translatedText, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return "", fmt.Errorf("LLM 响应内容类型错误")
	}
	log.Printf("[LLM] 翻译完成，文本长度: %d", len(translatedText))

	return translatedText, nil
}

// Chat 发送聊天请求到 LLM。
func (c *Client) Chat(ctx context.Context, model string, messages []Message, temperature float64) (string, error) {
	log.Printf("[LLM] 发送聊天请求 (模型: %s)", model)

	if model == "" {
		model = llmModel
	}

	llmReq := &LLMRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	var resp LLMResponse
	if err := c.doRequest(ctx, llmEndpoint, llmReq, &resp); err != nil {
		return "", fmt.Errorf("LLM 请求失败: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM 响应为空")
	}

	content, ok := resp.Choices[0].Message.Content.(string)
	if !ok {
		return "", fmt.Errorf("LLM 响应内容类型错误")
	}
	log.Printf("[LLM] 聊天完成，响应长度: %d", len(content))

	return content, nil
}
