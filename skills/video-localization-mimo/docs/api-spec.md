# MiMo API 规范

## API 端点

所有 MiMo API 调用使用统一端点：

```
POST https://api.xiaomimimo.com/v1/chat/completions
```

## 认证

支持两种认证方式：

```go
// 方式一：api-key header
req.Header.Set("api-key", apiKey)

// 方式二：Bearer token
req.Header.Set("Authorization", "Bearer "+apiKey)
```

## 语音识别 (ASR)

### 模型

- `mimo-v2.5-asr`

### 请求格式

```go
type ASRRequest struct {
    Model    string    `json:"model"`    // "mimo-v2.5-asr"
    Messages []Message `json:"messages"`
    ASROptions *ASROptions `json:"asr_options,omitempty"`
    Stream   bool      `json:"stream,omitempty"`
}

type Message struct {
    Role    string        `json:"role"`    // "user"
    Content []ContentPart `json:"content"`
}

type ContentPart struct {
    Type      string      `json:"type"`       // "input_audio"
    InputAudio *InputAudio `json:"input_audio,omitempty"`
}

type InputAudio struct {
    Data string `json:"data"` // "data:{MIME_TYPE};base64,{BASE64_AUDIO}"
}

type ASROptions struct {
    Language string `json:"language,omitempty"` // "auto", "zh", "en"
}
```

### 响应格式

```go
type ASRResponse struct {
    ID      string   `json:"id"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

type Choice struct {
    Index   int     `json:"index"`
    Message Message `json:"message"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

### 使用示例

```go
// 构建请求
audioBase64, _ := os.ReadFile("audio.wav")
audioData := fmt.Sprintf("data:audio/wav;base64,%s", base64.StdEncoding.EncodeToString(audioBase64))

req := &ASRRequest{
    Model: "mimo-v2.5-asr",
    Messages: []Message{
        {
            Role: "user",
            Content: []ContentPart{
                {
                    Type: "input_audio",
                    InputAudio: &InputAudio{
                        Data: audioData,
                    },
                },
            },
        },
    },
    ASROptions: &ASROptions{
        Language: "zh",
    },
}

// 发送请求
resp, err := client.Do(ctx, req)
```

### 限制

- 音频格式：WAV, MP3
- 最大大小：10MB（Base64 编码后）
- 语言：中文、英文、自动检测

---

## 语音合成 (TTS)

### 模型

| 模型 | 功能 | 音色来源 |
|------|------|----------|
| `mimo-v2.5-tts` | 预置音色合成 | 内置音色列表 |
| `mimo-v2.5-tts-voicedesign` | 文本描述生成音色 | 用户描述 |
| `mimo-v2.5-tts-voiceclone` | 音频样本克隆音色 | 参考音频 |

### 请求格式

```go
type TTSRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Audio    Audio     `json:"audio"`
    Stream   bool      `json:"stream,omitempty"`
}

type Audio struct {
    Format string `json:"format"` // "wav", "mp3", "pcm16"
    Voice  string `json:"voice"`  // 音色名称或 Base64 音频
}
```

### 预置音色列表

| 音色 ID | 语言 | 描述 |
|---------|------|------|
| `mimo_default` | 多语言 | 默认音色 |
| `default_zh` | 中文 | 中文女声 |
| `default_en` | 英文 | 英文女声 |
| `Chloe` | 英文 | 年轻女声 |
| `Mia` | 中文 | 温柔女声 |
| `Milo` | 英文 | 男声 |
| `Dean` | 英文 | 成熟男声 |

### 使用示例

#### 预置音色合成

```go
req := &TTSRequest{
    Model: "mimo-v2.5-tts",
    Messages: []Message{
        {
            Role:    "assistant",
            Content: "Hello, this is a test.",
        },
    },
    Audio: Audio{
        Format: "wav",
        Voice:  "Chloe",
    },
}
```

#### 语音克隆

```go
// 读取参考音频
refAudio, _ := os.ReadFile("reference.wav")
refBase64 := base64.StdEncoding.EncodeToString(refAudio)

req := &TTSRequest{
    Model: "mimo-v2.5-tts-voiceclone",
    Messages: []Message{
        {
            Role:    "assistant",
            Content: "Hello, this is cloned voice.",
        },
    },
    Audio: Audio{
        Format: "wav",
        Voice:  fmt.Sprintf("data:audio/wav;base64,%s", refBase64),
    },
}
```

### 响应格式

```go
type TTSResponse struct {
    ID      string       `json:"id"`
    Choices []TTSChoice  `json:"choices"`
    Usage   Usage        `json:"usage"`
}

type TTSChoice struct {
    Index   int         `json:"index"`
    Message TTSMessage  `json:"message"`
}

type TTSMessage struct {
    Role   string    `json:"role"`
    Audio  AudioData `json:"audio"`
}

type AudioData struct {
    Data string `json:"data"` // Base64 编码的音频
}
```

### 风格控制

通过 `user` 消息控制合成风格：

```go
req := &TTSRequest{
    Model: "mimo-v2.5-tts",
    Messages: []Message{
        {
            Role:    "user",
            Content: "Speak with excitement and fast pace",
        },
        {
            Role:    "assistant",
            Content: "I just won the lottery!",
        },
    },
    Audio: Audio{
        Format: "wav",
        Voice:  "Chloe",
    },
}
```

### 限制

- 文本长度：最大 8K tokens
- 参考音频大小：最大 10MB（Base64 编码后）
- 参考音频格式：WAV, MP3

---

## 大语言模型 (LLM)

### 模型

| 模型 | 功能 | 适用场景 |
|------|------|----------|
| `mimo-v2.5-pro` | 复杂推理 | 长文档、深度分析 |
| `mimo-v2.5` | 通用 | 多模态理解 |
| `mimo-v2-flash` | 快速响应 | 简单任务 |

### 请求格式

```go
type LLMRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_completion_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
}
```

### 使用示例

```go
req := &LLMRequest{
    Model: "mimo-v2-flash",
    Messages: []Message{
        {
            Role:    "system",
            Content: "You are a professional translator.",
        },
        {
            Role:    "user",
            Content: "Translate the following Chinese text to English:\n\n你好世界",
        },
    },
    MaxTokens: 1024,
}
```

### 响应格式

```go
type LLMResponse struct {
    ID      string     `json:"id"`
    Choices []LLMChoice `json:"choices"`
    Usage   Usage      `json:"usage"`
}

type LLMChoice struct {
    Index   int       `json:"index"`
    Message LLMMessage `json:"message"`
}

type LLMMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

---

## 通用类型

```go
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## 错误处理

### 错误响应

```go
type ErrorResponse struct {
    Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Message string `json:"message"`
    Type    string `json:"type"`
    Code    string `json:"code"`
}
```

### 常见错误

| HTTP 状态码 | 错误类型 | 处理方式 |
|-------------|----------|----------|
| 401 | 认证失败 | 检查 API Key |
| 429 | 请求过快 | 重试（指数退避） |
| 500 | 服务器错误 | 重试（有限次数） |
| 503 | 服务不可用 | 等待后重试 |

### 重试策略

```go
type RetryConfig struct {
    MaxRetries  int           // 最大重试次数，默认 3
    InitialWait time.Duration // 初始等待时间，默认 1s
    MaxWait     time.Duration // 最大等待时间，默认 30s
    Multiplier  float64       // 退避倍数，默认 2.0
}
```

## 限流

| 模型 | RPM（每分钟请求数） | TPM（每分钟 Token 数） |
|------|---------------------|------------------------|
| mimo-v2.5-asr | 100 | 10K |
| mimo-v2.5-tts | 100 | 10M |
| mimo-v2-flash | 100 | 100K |

## 最佳实践

1. **音频预处理**
   - 转换为 WAV 格式
   - 采样率 16kHz 或 24kHz
   - 单声道
   - 控制文件大小 < 10MB

2. **文本预处理**
   - 去除多余空白
   - 分段处理长文本
   - 保留标点符号以控制停顿

3. **并发控制**
   - 遵守 RPM 限制
   - 使用令牌桶限流
   - 批量请求时添加延迟

4. **错误处理**
   - 记录请求 ID 用于排查
   - 实现指数退避重试
   - 保存中间结果支持断点续传
