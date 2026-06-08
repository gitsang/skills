# 实现计划

## 阶段划分

### 阶段一：基础框架（1-2 天）

**目标**：搭建项目骨架，实现 CLI 框架和配置管理

**任务清单**：
- [ ] 初始化 Go 模块
- [ ] 配置 cobra CLI 框架
- [ ] 实现配置文件管理（`~/.config/video-localization-mimo/config.yaml`）
- [ ] 实现环境变量支持（`MIMO_API_KEY`）
- [ ] 创建基础目录结构
- [ ] 添加日志框架（zerolog 或 slog）

**产出**：
```
video-localization-mimo/
├── main.go
├── go.mod
├── cmd/
│   └── root.go
├── internal/
│   └── config/
│       └── config.go
└── pkg/
    └── logger/
        └── logger.go
```

---

### 阶段二：MiMo API 客户端（2-3 天）

**目标**：实现 MiMo API 的统一封装

**任务清单**：
- [ ] 实现 HTTP 客户端基类
  - 请求构建
  - 认证头注入
  - 响应解析
  - 错误处理
  - 重试逻辑（指数退避）
  - 限流控制（令牌桶）
- [ ] 实现 ASR 模块
  - 音频文件 Base64 编码
  - 请求构建和发送
  - 响应解析
- [ ] 实现 TTS 模块
  - 预置音色合成
  - 语音克隆
  - 音频数据解码和保存
- [ ] 实现 LLM 模块
  - 翻译请求
  - 流式响应支持（可选）
- [ ] 编写单元测试

**产出**：
```
pkg/mimo/
├── client.go       # HTTP 客户端
├── asr.go          # 语音识别
├── tts.go          # 语音合成
├── llm.go          # 大语言模型
├── types.go        # 请求/响应类型
├── retry.go        # 重试逻辑
├── ratelimit.go    # 限流控制
└── client_test.go  # 测试
```

**关键代码**：

```go
// client.go
type Client struct {
    httpClient *http.Client
    apiKey     string
    baseURL    string
    retryConf  RetryConfig
    limiter    *rate.Limiter
}

func NewClient(apiKey string, opts ...Option) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: 30 * time.Second},
        apiKey:     apiKey,
        baseURL:    "https://api.xiaomimimo.com/v1",
        retryConf:  DefaultRetryConfig(),
        limiter:    rate.NewLimiter(rate.Limit(100), 10), // 100 RPM
    }
}

func (c *Client) Do(ctx context.Context, req interface{}, resp interface{}) error {
    // 限流
    if err := c.limiter.Wait(ctx); err != nil {
        return err
    }
    
    // 重试
    for i := 0; i <= c.retryConf.MaxRetries; i++ {
        err := c.doOnce(ctx, req, resp)
        if err == nil {
            return nil
        }
        if !isRetryable(err) || i == c.retryConf.MaxRetries {
            return err
        }
        time.Sleep(c.retryConf.Backoff(i))
    }
    return nil
}
```

---

### 阶段三：SRT 处理（1 天）

**目标**：实现 SRT 字幕文件的解析和生成

**任务清单**：
- [ ] 实现 SRT 解析器
  - 解析时间戳
  - 解析字幕文本
  - 处理多行字幕
- [ ] 实现 SRT 生成器
  - 格式化时间戳
  - 输出标准 SRT 格式
- [ ] 实现 SRT 验证
  - 检查时间戳顺序
  - 检查段落连续性
- [ ] 编写单元测试

**产出**：
```
pkg/srt/
├── parser.go       # SRT 解析
├── generator.go    # SRT 生成
├── validator.go    # SRT 验证
├── types.go        # 数据类型
└── parser_test.go  # 测试
```

**关键代码**：

```go
// types.go
type Segment struct {
    Index int
    Start time.Duration
    End   time.Duration
    Text  string
}

// parser.go
func Parse(reader io.Reader) ([]Segment, error) {
    scanner := bufio.NewScanner(reader)
    var segments []Segment
    var current Segment
    
    for scanner.Scan() {
        line := scanner.Text()
        // 解析逻辑...
    }
    return segments, nil
}
```

---

### 阶段四：音频处理（2 天）

**目标**：实现音频格式转换和混音

**任务清单**：
- [ ] 实现 FFmpeg 封装
  - 命令构建
  - 进度解析
  - 错误处理
- [ ] 实现音频探测
  - 时长
  - 采样率
  - 声道数
- [ ] 实现音频转换
  - 格式转换
  - 采样率转换
  - 声道转换
- [ ] 实现音频混音
  - 多轨混音
  - 音量调节
  - 时间偏移
- [ ] 编写单元测试

**产出**：
```
pkg/ffmpeg/
├── runner.go       # 命令执行
├── probe.go        # 媒体探测
├── convert.go      # 格式转换
└── compose.go      # 视频合成

pkg/audio/
├── mixer.go        # WAV 混音
├── converter.go    # 格式转换
└── aligner.go      # 时间对齐
```

**关键代码**：

```go
// ffmpeg/runner.go
type Runner struct {
    ffmpegPath  string
    ffprobePath string
}

func (r *Runner) Probe(ctx context.Context, path string) (*MediaInfo, error) {
    cmd := exec.CommandContext(ctx, r.ffprobePath,
        "-v", "error",
        "-show_entries", "format=duration,size:stream=codec_type,codec_name,sample_rate,channels",
        "-of", "json",
        path,
    )
    // 解析输出...
}

// audio/mixer.go
func MixTracks(tracks []Track, output string, sampleRate int) error {
    // 使用 ffmpeg amix filter
    // 或直接读取 WAV 文件进行混音
}
```

---

### 阶段五：CLI 命令实现（2-3 天）

**目标**：实现各个子命令

**任务清单**：
- [ ] `transcribe` 命令
  - 输入：音频文件
  - 输出：文本文件 + SRT 文件
- [ ] `translate` 命令
  - 输入：文本文件
  - 输出：翻译后的文本文件
- [ ] `tts` 命令
  - 输入：文本文件
  - 输出：音频文件
- [ ] `clone` 命令
  - 输入：文本文件 + 参考音频
  - 输出：克隆后的音频文件
- [ ] `subtitle` 命令
  - 输入：源 SRT + 目标文本
  - 输出：双语 ASS 字幕
- [ ] `align` 命令
  - 输入：源 SRT + 目标音频
  - 输出：对齐后的音频
- [ ] `compose` 命令
  - 输入：视频 + 音频 + 字幕
  - 输出：最终视频
- [ ] 添加进度条显示

**产出**：
```
cmd/
├── root.go
├── transcribe.go
├── translate.go
├── tts.go
├── clone.go
├── subtitle.go
├── align.go
├── compose.go
└── progress.go     # 进度条
```

**命令设计**：

```go
// transcribe.go
var transcribeCmd = &cobra.Command{
    Use:   "transcribe",
    Short: "Transcribe audio to text and SRT",
    RunE: func(cmd *cobra.Command, args []string) error {
        audioPath, _ := cmd.Flags().GetString("audio")
        lang, _ := cmd.Flags().GetString("lang")
        txtOut, _ := cmd.Flags().GetString("txt-out")
        srtOut, _ := cmd.Flags().GetString("srt-out")
        
        // 1. 读取音频文件
        // 2. Base64 编码
        // 3. 调用 MiMo ASR API
        // 4. 保存文本和 SRT
        return nil
    },
}

func init() {
    transcribeCmd.Flags().StringP("audio", "a", "", "Input audio file (required)")
    transcribeCmd.Flags().StringP("lang", "l", "auto", "Source language (auto/zh/en)")
    transcribeCmd.Flags().String("txt-out", "", "Text output path")
    transcribeCmd.Flags().String("srt-out", "", "SRT output path")
    transcribeCmd.MarkFlagRequired("audio")
}
```

---

### 阶段六：工作流编排（2 天）

**目标**：实现完整的本地化流程

**任务清单**：
- [ ] 实现 Pipeline 编排
  - 步骤依赖管理
  - 错误恢复
  - 断点续传
- [ ] 实现 `localize` 命令
  - 参数解析
  - 工作目录创建
  - 全流程执行
- [ ] 实现进度报告
  - 整体进度
  - 当前步骤
  - 预计剩余时间
- [ ] 实现报告生成
  - 命令记录
  - 问题记录
  - 最终报告

**产出**：
```
internal/
└── workflow/
    ├── pipeline.go     # 流程编排
    ├── steps.go        # 步骤定义
    ├── report.go       # 报告生成
    └── pipeline_test.go

cmd/
└── localize.go         # 完整本地化命令
```

**关键代码**：

```go
// workflow/pipeline.go
type Pipeline struct {
    config *config.Config
    client *mimo.Client
    steps  []Step
}

type Step interface {
    Name() string
    Run(ctx context.Context, state *State) error
}

func (p *Pipeline) Run(ctx context.Context, req *Request) error {
    state := &State{
        SourceVideo: req.SourceVideo,
        SourceLang:  req.SourceLang,
        TargetLang:  req.TargetLang,
        OutputDir:   req.OutputDir,
    }
    
    for _, step := range p.steps {
        if err := step.Run(ctx, state); err != nil {
            return fmt.Errorf("step %s failed: %w", step.Name(), err)
        }
    }
    return nil
}
```

---

### 阶段七：Web 应用（2-3 天）

**目标**：实现 Web UI，支持浏览器操作

**任务清单**：
- [ ] 实现 HTTP 服务器
  - 路由配置（gorilla/mux）
  - 中间件（日志、CORS）
  - 静态文件服务（embed.FS）
- [ ] 实现页面模板
  - 基础布局（base.html）
  - 首页（home.html）
  - 任务详情页（task.html）
  - 设置页（settings.html）
  - 部分组件（partials）
- [ ] 实现文件上传
  - multipart 处理
  - 大小限制
  - 类型验证
- [ ] 实现任务管理
  - 任务存储（内存/文件）
  - 异步执行
  - 状态查询
- [ ] 实现进度推送
  - WebSocket 升级
  - 实时进度更新
  - HTMX 集成
- [ ] 实现 REST API
  - 任务 CRUD
  - 文件下载
  - 设置管理
- [ ] 添加 HTMX 交互
  - 表单提交
  - 轮询更新
  - 动态加载

**产出**：
```
cmd/
└── server/
    ├── main.go         # 服务器入口
    └── router.go       # 路由配置

internal/
├── handler/            # HTTP handlers
│   ├── pages.go        # 页面渲染
│   ├── api.go          # REST API
│   ├── upload.go       # 文件上传
│   └── websocket.go    # WebSocket
├── middleware/          # 中间件
│   ├── auth.go
│   ├── cors.go
│   └── logger.go
└── task/               # 任务管理
    ├── store.go        # 任务存储
    └── types.go        # 任务类型

web/
├── templates/          # Go 模板
│   ├── layouts/
│   │   └── base.html
│   ├── pages/
│   │   ├── home.html
│   │   ├── task.html
│   │   └── settings.html
│   └── partials/
│       ├── header.html
│       ├── progress.html
│       └── task-card.html
└── static/             # 静态资源
    ├── css/
    │   ├── pico.min.css
    │   └── style.css
    └── js/
        └── htmx.min.js
```

**关键代码**：

```go
// cmd/server/main.go
func main() {
    // 加载配置
    cfg := config.Load()
    
    // 初始化依赖
    client := mimo.NewClient(cfg.MiMo.APIKey)
    taskStore := task.NewMemoryStore()
    pipeline := workflow.NewPipeline(cfg, client)
    
    // 创建 handler
    handler := handler.New(cfg, taskStore, pipeline)
    
    // 配置路由
    router := router.New(handler)
    
    // 启动服务器
    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
        Handler: router,
    }
    
    log.Printf("Server starting on http://localhost%s", srv.Addr)
    srv.ListenAndServe()
}

// internal/handler/upload.go
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
    // 限制上传大小
    r.Body = http.MaxBytesReader(w, r.Body, h.config.Server.UploadLimit)
    
    // 解析表单
    file, header, err := r.FormFile("video")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // 保存文件并创建任务
    taskID := h.saveFile(file, header)
    
    // 启动异步处理
    go h.pipeline.StartAsync(r.Context(), taskID)
    
    // HTMX 重定向
    w.Header().Set("HX-Redirect", fmt.Sprintf("/task/%s", taskID))
}
```

---

### 阶段八：测试和文档（2 天）

**目标**：完善测试覆盖和用户文档

**任务清单**：
- [ ] 编写集成测试
  - MiMo API 调用测试
  - 端到端流程测试
  - Web API 测试
- [ ] 编写使用示例
- [ ] 编写 README.md
- [ ] 编写安装指南
- [ ] 编写故障排除指南
- [ ] 配置 GitHub Actions CI

**产出**：
```
├── README.md
├── docs/
│   ├── installation.md
│   ├── usage.md
│   ├── examples.md
│   └── troubleshooting.md
├── .github/
│   └── workflows/
│       └── ci.yml
└── testdata/
    ├── audio/
    └── srt/
```

---

## 里程碑

| 阶段 | 任务 | 预计时间 | 产出 |
|------|------|----------|------|
| 一 | 基础框架 | 1-2 天 | CLI 骨架 + 配置管理 |
| 二 | MiMo API | 2-3 天 | API 客户端 |
| 三 | SRT 处理 | 1 天 | SRT 解析/生成 |
| 四 | 音频处理 | 2 天 | FFmpeg 封装 |
| 五 | CLI 命令 | 2-3 天 | 所有子命令 |
| 六 | 工作流 | 2 天 | 完整流程 |
| **七** | **Web 应用** | **2-3 天** | **Web UI** |
| 八 | 测试文档 | 2 天 | 测试 + 文档 |
| **总计** | | **14-18 天** | 完整工具 + Web UI |

---

## 风险和应对

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| MiMo API 限流 | TTS 合成变慢 | 实现限流控制，支持并发 |
| 音频格式兼容 | 无法处理某些格式 | 统一转换为 WAV |
| 长文本超限 | TTS 失败 | 分段处理，拼接音频 |
| 网络不稳定 | API 调用失败 | 重试机制，断点续传 |
| API 变更 | 代码需要更新 | 抽象接口，易于适配 |
| 大文件上传 | 内存溢出 | 流式处理，分块上传 |
| WebSocket 断开 | 进度丢失 | 自动重连，轮询降级 |

---

## 未来扩展

- [ ] 支持更多 TTS 音色
- [ ] 支持批量处理
- [ ] 支持更多字幕格式（ASS, VTT）
- [ ] 支持视频字幕烧录
- [ ] 支持 GPU 加速的本地模型
- [ ] 支持用户认证
- [ ] 支持多租户
- [ ] 支持插件扩展
