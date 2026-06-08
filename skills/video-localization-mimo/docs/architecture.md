# 架构设计

## 系统架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        video-localization-mimo                      │
├─────────────────────────────────────────────────────────────────────┤
│  Entry Points                                                       │
│  ┌──────────────────────┐  ┌──────────────────────────────────┐    │
│  │  CLI (cobra)         │  │  Web Server (net/http + HTMX)    │    │
│  │  cmd/cli/            │  │  cmd/server/                     │    │
│  └──────────┬───────────┘  └──────────────┬───────────────────┘    │
├─────────────┼─────────────────────────────┼────────────────────────┤
│             │         Shared Core         │                        │
│  ┌──────────┴─────────────────────────────┴───────────┐            │
│  │              internal/workflow/                     │            │
│  │              (Pipeline Orchestration)               │            │
│  └────────────────────────┬────────────────────────────┘            │
├───────────────────────────┼────────────────────────────────────────┤
│  Core Layer               │                                        │
│  ┌────────────┐ ┌─────────┴───┐ ┌────────────┐ ┌────────────┐     │
│  │   ASR      │ │    LLM      │ │    TTS     │ │   FFmpeg   │     │
│  │  Handler   │ │   Handler   │ │   Handler  │ │   Wrapper  │     │
│  └─────┬──────┘ └──────┬──────┘ └─────┬──────┘ └─────┬──────┘     │
├────────┼───────────────┼──────────────┼──────────────┼─────────────┤
│  API Layer              │              │              │             │
│  ┌─────┴───────────────┴──────────────┴──────┐                      │
│  │            MiMo API Client                │                      │
│  │      (OpenAI-compatible HTTP API)         │                      │
│  └───────────────────────────────────────────┘                      │
├─────────────────────────────────────────────────────────────────────┤
│  Utility Layer                                                      │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │
│  │SRT Parser│ │WAV Mixer │ │Base64    │ │File Utils│ │WebSocket │ │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                          ┌─────────────────┐
                          │   MiMo API      │
                          │   (云端)        │
                          └─────────────────┘
```

## 目录结构

```
video-localization-mimo/
├── cmd/                            # 入口点
│   ├── cli/                        # CLI 入口
│   │   ├── root.go                 # 根命令
│   │   ├── transcribe.go           # 语音识别
│   │   ├── translate.go            # 翻译润色
│   │   ├── tts.go                  # 语音合成
│   │   ├── clone.go                # 语音克隆
│   │   ├── subtitle.go             # 字幕生成
│   │   ├── align.go                # 音频对齐
│   │   ├── compose.go              # 视频合成
│   │   └── localize.go             # 完整本地化流程
│   └── server/                     # Web 入口
│       ├── main.go                 # 服务器启动
│       └── router.go               # 路由配置
├── internal/                       # 内部包
│   ├── config/                     # 配置管理
│   │   └── config.go
│   ├── workflow/                   # 工作流编排（CLI 和 Web 共享）
│   │   ├── pipeline.go
│   │   ├── steps.go
│   │   └── state.go
│   ├── handler/                    # HTTP handlers（Web 专用）
│   │   ├── pages.go                # 页面渲染
│   │   ├── api.go                  # REST API
│   │   ├── websocket.go            # WebSocket 进度推送
│   │   └── upload.go               # 文件上传
│   ├── middleware/                  # HTTP 中间件
│   │   ├── auth.go                 # 认证
│   │   ├── cors.go                 # CORS
│   │   └── logger.go               # 请求日志
│   └── task/                       # 任务管理
│       ├── store.go                # 任务存储
│       └── types.go                # 任务类型
├── pkg/                            # 核心包（可复用）
│   ├── mimo/                       # MiMo API 客户端
│   │   ├── client.go
│   │   ├── asr.go
│   │   ├── tts.go
│   │   ├── llm.go
│   │   └── types.go
│   ├── srt/                        # SRT 处理
│   │   ├── parser.go
│   │   ├── generator.go
│   │   └── types.go
│   ├── audio/                      # 音频处理
│   │   ├── mixer.go
│   │   ├── converter.go
│   │   └── aligner.go
│   └── ffmpeg/                     # FFmpeg 封装
│       ├── runner.go
│       ├── probe.go
│       └── compose.go
├── web/                            # Web 前端资源
│   ├── templates/                  # Go 模板
│   │   ├── layouts/
│   │   │   └── base.html
│   │   ├── pages/
│   │   │   ├── home.html
│   │   │   ├── upload.html
│   │   │   ├── task.html
│   │   │   └── settings.html
│   │   └── partials/
│   │       ├── header.html
│   │       ├── footer.html
│   │       ├── progress.html
│   │       └── task-card.html
│   └── static/                     # 静态资源
│       ├── css/
│       │   └── style.css
│       ├── js/
│       │   ├── htmx.min.js
│       │   └── app.js
│       └── img/
├── docs/                           # 文档
├── go.mod
├── go.sum
└── main.go                         # 统一入口
```

## 核心模块

### 1. MiMo API Client (`pkg/mimo/`)

统一封装 MiMo API 调用，CLI 和 Web 共享：

```go
// 示例用法
client := mimo.NewClient(apiKey)

// 语音识别
transcript, err := client.Transcribe(ctx, &mimo.TranscribeRequest{
    AudioPath: "source.mp3",
    Language:  "zh",
})

// 语音合成
audioPath, err := client.Synthesize(ctx, &mimo.SynthesizeRequest{
    Text:   "Hello, world!",
    Voice:  "Chloe",
    Format: "wav",
})

// 语音克隆
audioPath, err := client.CloneVoice(ctx, &mimo.CloneRequest{
    Text:           "Hello, world!",
    ReferenceAudio: "reference.wav",
    Format:         "wav",
})
```

### 2. 工作流编排 (`internal/workflow/`)

CLI 和 Web 共享的工作流引擎：

```go
// 定义步骤
type Step interface {
    Name() string
    Run(ctx context.Context, state *State) error
    Progress() float64
}

// Pipeline 编排
pipeline := workflow.NewPipeline(config, client)

// CLI 使用
err := pipeline.Run(ctx, &workflow.Request{...})

// Web 使用（异步）
taskID := pipeline.StartAsync(ctx, &workflow.Request{...})
status := pipeline.GetStatus(taskID)
```

### 3. 任务管理 (`internal/task/`)

Web 模式的任务状态管理：

```go
type Task struct {
    ID        string
    Status    TaskStatus  // pending, running, completed, failed
    Progress  float64     // 0.0 - 1.0
    Step      string      // 当前步骤
    Result    *Result
    Error     error
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Store interface {
    Create(task *Task) error
    Get(id string) (*Task, error)
    Update(task *Task) error
    List() ([]*Task, error)
}
```

### 4. HTTP Handlers (`internal/handler/`)

Web 请求处理：

```go
// 页面渲染
func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
    h.tmpl.ExecuteTemplate(w, "home.html", nil)
}

// 文件上传
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
    file, header, _ := r.FormFile("video")
    taskID := h.taskStore.Create(file, header.Filename)
    
    // HTMX 重定向
    w.Header().Set("HX-Redirect", fmt.Sprintf("/task/%s", taskID))
}

// 任务状态 API（HTMX 轮询）
func (h *Handler) TaskStatusHandler(w http.ResponseWriter, r *http.Request) {
    taskID := mux.Vars(r)["id"]
    task, _ := h.taskStore.Get(taskID)
    
    // 返回 HTML 片段
    h.tmpl.ExecuteTemplate(w, "partials/task-status.html", task)
}
```

### 5. WebSocket 进度推送

实时进度更新：

```go
func (h *Handler) TaskProgressWS(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    taskID := mux.Vars(r)["id"]
    
    // 订阅任务进度
    ch := h.progress.Subscribe(taskID)
    defer h.progress.Unsubscribe(taskID, ch)
    
    for progress := range ch {
        conn.WriteJSON(progress)
    }
}
```

## 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                        Web / CLI 入口                           │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   Workflow Pipeline   │
                    └───────────┬───────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│ 1. Extract   │      │ 2. Transcribe│      │ 3. Translate │
│    Audio     │ ───→ │    (ASR)     │ ───→ │    (LLM)     │
│  (ffmpeg)    │      │  (MiMo API)  │      │  (MiMo API)  │
└──────────────┘      └──────────────┘      └──────┬───────┘
                                                    │
        ┌───────────────────────┬───────────────────┘
        │                       │
        ▼                       ▼
┌──────────────┐      ┌──────────────┐
│ 4. Synthesize│      │ 5. Align     │
│    (TTS)     │ ───→ │    Audio     │
│  (MiMo API)  │      │  (local)     │
└──────────────┘      └──────┬───────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│ 6. Generate  │    │ 7. Compose   │    │ 8. Package   │
│   Subtitles  │    │    Video     │ ─→ │   & Report   │
│   (local)    │    │   (ffmpeg)   │    │              │
└──────────────┘    └──────────────┘    └──────────────┘
```

## 配置管理

配置文件位置：`~/.config/video-localization-mimo/config.yaml`

```yaml
mimo:
  api_key: "your-api-key"  # 或通过 MIMO_API_KEY 环境变量
  base_url: "https://api.xiaomimimo.com/v1"
  timeout: 30s

ffmpeg:
  path: "ffmpeg"
  ffprobe_path: "ffprobe"

server:
  host: "0.0.0.0"
  port: 8080
  upload_limit: 500MB    # 上传文件大小限制
  task_ttl: 24h          # 任务保留时间

defaults:
  source_lang: "zh"
  target_lang: "en"
  voice: "Chloe"
  audio_format: "wav"
  video_codec: "libx264"
  audio_codec: "aac"
```

## 错误处理

- API 调用失败：重试 3 次，指数退避
- 音频格式不支持：自动转换为 WAV
- 字幕数量不匹配：报错并提示用户检查
- ffmpeg 执行失败：捕获 stderr 并报错
- Web 上传失败：返回详细错误信息
- WebSocket 断开：自动重连

## 并发处理

- 多个 TTS 段落可并发合成（控制并发数）
- 使用 goroutine pool 管理并发
- Web 模式支持多任务并行执行
- 进度条/进度卡片实时更新
