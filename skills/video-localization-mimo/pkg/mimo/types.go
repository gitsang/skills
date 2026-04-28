// Package mimo 提供 MiMo API 的 Go 客户端实现。
//
// MiMo API 提供以下能力：
// - ASR（语音识别）：音频文件转文本
// - TTS（语音合成）：文本转音频
// - LLM（大语言模型）：文本翻译、润色等
//
// 所有 API 调用使用统一端点：POST https://api.xiaomimimo.com/v1/chat/completions
// 认证方式：api-key header
package mimo

import (
	"fmt"
	"time"
)

// ============================================================
// 通用类型
// ============================================================

// Message 表示对话消息，用于 ASR、TTS、LLM 请求。
type Message struct {
	Role    string      `json:"role"`    // "user", "assistant", "system"
	Content interface{} `json:"content"` // 文本内容（string）或多模态内容（[]ContentPart）
}

// ContentPart 表示多模态内容的一部分。
type ContentPart struct {
	Type       string      `json:"type"`                  // "input_audio", "text"
	InputAudio *InputAudio `json:"input_audio,omitempty"` // 音频数据
	Text       string      `json:"text,omitempty"`        // 文本数据
}

// InputAudio 表示输入音频数据。
type InputAudio struct {
	Data string `json:"data"` // "data:{MIME_TYPE};base64,{BASE64_AUDIO}"
}

// Usage 表示 API 调用的 token 使用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ============================================================
// ASR 相关类型
// ============================================================

// ASRRequest 表示语音识别请求。
type ASRRequest struct {
	Model       string       `json:"model"`                 // "mimo-v2.5-asr"
	Messages    []Message    `json:"messages"`
	ASROptions  *ASROptions  `json:"asr_options,omitempty"` // ASR 选项
	Stream      bool         `json:"stream,omitempty"`      // 是否流式输出
}

// ASROptions 表示 ASR 配置选项。
type ASROptions struct {
	Language string `json:"language,omitempty"` // "auto", "zh", "en"
}

// ASRResponse 表示语音识别响应。
type ASRResponse struct {
	ID      string      `json:"id"`
	Choices []ASRChoice `json:"choices"`
	Usage   Usage       `json:"usage"`
}

// ASRChoice 表示 ASR 响应中的一个选项。
type ASRChoice struct {
	Index   int        `json:"index"`
	Message ASRMessage `json:"message"`
}

// ASRMessage 表示 ASR 响应中的消息。
type ASRMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // 识别出的文本
}

// TranscribeResult 表示语音识别的结果。
type TranscribeResult struct {
	Text     string    `json:"text"`     // 识别出的完整文本
	Segments []Segment `json:"segments"` // 时间片段（如果有）
}

// Segment 表示一个时间片段。
type Segment struct {
	Start float64 `json:"start"` // 开始时间（秒）
	End   float64 `json:"end"`   // 结束时间（秒）
	Text  string  `json:"text"`  // 片段文本
}

// ============================================================
// TTS 相关类型
// ============================================================

// TTSRequest 表示语音合成请求。
type TTSRequest struct {
	Model    string    `json:"model"`    // "mimo-v2.5-tts", "mimo-v2.5-tts-voiceclone"
	Messages []Message `json:"messages"`
	Audio    Audio     `json:"audio"`    // 音频配置
	Stream   bool      `json:"stream,omitempty"`
}

// Audio 表示音频配置。
type Audio struct {
	Format string `json:"format"` // "wav", "mp3", "pcm16"
	Voice  string `json:"voice"`  // 音色名称或 Base64 音频（用于克隆）
}

// TTSResponse 表示语音合成响应。
type TTSResponse struct {
	ID      string      `json:"id"`
	Choices []TTSChoice `json:"choices"`
	Usage   Usage       `json:"usage"`
}

// TTSChoice 表示 TTS 响应中的一个选项。
type TTSChoice struct {
	Index   int        `json:"index"`
	Message TTSMessage `json:"message"`
}

// TTSMessage 表示 TTS 响应中的消息。
type TTSMessage struct {
	Role  string    `json:"role"`
	Audio AudioData `json:"audio"`
}

// AudioData 表示音频数据。
type AudioData struct {
	Data string `json:"data"` // Base64 编码的音频
}

// SynthesizeRequest 表示语音合成的简化请求。
type SynthesizeRequest struct {
	Text   string `json:"text"`   // 要合成的文本
	Voice  string `json:"voice"`  // 音色名称
	Format string `json:"format"` // 输出格式："wav", "mp3"
	Style  string `json:"style,omitempty"` // 风格描述（可选）
}

// CloneRequest 表示语音克隆请求。
type CloneRequest struct {
	Text           string `json:"text"`            // 要合成的文本
	ReferenceAudio string `json:"reference_audio"` // 参考音频文件路径
	Format         string `json:"format"`          // 输出格式："wav", "mp3"
}

// ============================================================
// LLM 相关类型
// ============================================================

// LLMRequest 表示大语言模型请求。
type LLMRequest struct {
	Model       string    `json:"model"`                           // "mimo-v2.5-pro", "mimo-v2.5", "mimo-v2-flash"
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_completion_tokens,omitempty"` // 最大生成 token 数
	Temperature float64   `json:"temperature,omitempty"`           // 温度参数
	Stream      bool      `json:"stream,omitempty"`
}

// LLMResponse 表示大语言模型响应。
type LLMResponse struct {
	ID      string      `json:"id"`
	Choices []LLMChoice `json:"choices"`
	Usage   Usage       `json:"usage"`
}

// LLMChoice 表示 LLM 响应中的一个选项。
type LLMChoice struct {
	Index   int        `json:"index"`
	Message LLMMessage `json:"message"`
}

// LLMMessage 表示 LLM 响应中的消息。
type LLMMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// TranslateRequest 表示翻译请求。
type TranslateRequest struct {
	Text       string  `json:"text"`        // 要翻译的文本
	SourceLang string  `json:"source_lang"` // 源语言
	TargetLang string  `json:"target_lang"` // 目标语言
	SystemPrompt string `json:"system_prompt,omitempty"` // 系统提示词
	Temperature  float64 `json:"temperature,omitempty"`  // 温度参数
}

// ============================================================
// 错误类型
// ============================================================

// ErrorResponse 表示 MiMo API 的错误响应。
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 表示错误详情。
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// APIError 表示 MiMo API 调用错误。
type APIError struct {
	StatusCode int       `json:"status_code"` // HTTP 状态码
	Type       string    `json:"type"`        // 错误类型
	Code       string    `json:"code"`        // 错误代码
	Message    string    `json:"message"`     // 错误消息
	RequestID  string    `json:"request_id"`  // 请求 ID（用于排查）
	RetryAfter time.Duration `json:"-"`       // 建议重试等待时间
}

func (e *APIError) Error() string {
	return fmt.Sprintf("MiMo API 错误 [%d] %s: %s (请求ID: %s)",
		e.StatusCode, e.Type, e.Message, e.RequestID)
}

// IsRetryable 判断错误是否可重试。
func (e *APIError) IsRetryable() bool {
	switch e.StatusCode {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

// ============================================================
// 配置类型
// ============================================================

// ClientConfig 表示 MiMo 客户端配置。
type ClientConfig struct {
	APIKey  string        // API 密钥
	BaseURL string        // API 基础 URL
	Timeout time.Duration // 请求超时时间

	// 重试配置
	MaxRetries  int           // 最大重试次数，默认 3
	InitialWait time.Duration // 初始等待时间，默认 1s
	MaxWait     time.Duration // 最大等待时间，默认 30s
	Multiplier  float64       // 退避倍数，默认 2.0

	// 限流配置
	RateLimit int // 每分钟请求数限制，默认 100
}

// DefaultConfig 返回默认配置。
func DefaultConfig(apiKey string) *ClientConfig {
	return &ClientConfig{
		APIKey:      apiKey,
		BaseURL:     "https://api.xiaomimimo.com/v1",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
		RateLimit:   100,
	}
}
