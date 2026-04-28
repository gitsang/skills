# Web 应用设计

## 概述

使用 Go 标准库 `net/http` + `html/template` + HTMX 实现轻量级 Web UI，无需前端构建工具。

## 技术栈

| 组件 | 技术选择 | 理由 |
|------|----------|------|
| HTTP 路由 | `net/http` + `gorilla/mux` | Go 标准库 + 灵活路由 |
| 模板引擎 | `html/template` | Go 原生，安全 |
| 前端交互 | HTMX | 无需 JS 框架，声明式 AJAX |
| 样式 | Pico CSS / Simple.css | 无类 CSS 框架，快速开发 |
| 实时更新 | WebSocket | 进度推送 |
| 文件上传 | multipart/form-data | 标准上传 |

## 页面设计

### 1. 首页 (`/`)

```
┌─────────────────────────────────────────────────────────────┐
│  🎬 Video Localization                            [Settings] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                                                     │   │
│  │           Drag & drop your video here               │   │
│  │                  or click to upload                  │   │
│  │                                                     │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  📁 Choose File                              │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                                                     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Recent Tasks                                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ✅ video-en.mp4           Completed    2 min ago    │   │
│  │ ⏳ tutorial-zh.mp4        Processing   50%          │   │
│  │ ❌ demo.mp4               Failed       10 min ago   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**模板**: `web/templates/pages/home.html`

```html
{{template "base" .}}
{{define "content"}}
<main class="container">
    <section>
        <h1>Video Localization</h1>
        <p>Translate and dub videos using MiMo AI</p>
    </section>

    <!-- Upload Area -->
    <section>
        <form hx-post="/api/upload"
              hx-encoding="multipart/form-data"
              hx-trigger="submit"
              hx-indicator="#upload-spinner">
            
            <div id="drop-zone" 
                 class="drop-zone"
                 _="on dragover add .drag-over to me
                    on dragleave remove .drag-over from me
                    on drop get the transfer item and trigger upload">
                
                <label for="video-upload">
                    📁 Drag & drop or click to upload
                </label>
                <input type="file" 
                       id="video-upload"
                       name="video" 
                       accept="video/*,audio/*"
                       required>
            </div>

            <!-- Settings -->
            <details>
                <summary>Advanced Settings</summary>
                <div class="grid">
                    <label>
                        Source Language
                        <select name="source_lang">
                            <option value="auto">Auto Detect</option>
                            <option value="zh">Chinese</option>
                            <option value="en">English</option>
                        </select>
                    </label>
                    <label>
                        Target Language
                        <select name="target_lang">
                            <option value="en">English</option>
                            <option value="zh">Chinese</option>
                            <option value="ja">Japanese</option>
                        </select>
                    </label>
                    <label>
                        Voice
                        <select name="voice">
                            <option value="Chloe">Chloe (EN)</option>
                            <option value="Mia">Mia (ZH)</option>
                            <option value="mimo_default">Default</option>
                        </select>
                    </label>
                </div>
                <label>
                    <input type="checkbox" name="use_voice_clone" role="switch">
                    Use Voice Cloning
                </label>
            </details>

            <button type="submit">
                <span id="upload-spinner" class="htmx-indicator">⏳</span>
                Start Localization
            </button>
        </form>
    </section>

    <!-- Recent Tasks -->
    <section>
        <h2>Recent Tasks</h2>
        <div hx-get="/api/tasks" hx-trigger="every 5s" hx-swap="innerHTML">
            {{template "task-list" .Tasks}}
        </div>
    </section>
</main>
{{end}}
```

### 2. 任务详情页 (`/task/:id`)

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back                              task-abc123            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Source: tutorial.mp4                    [▶ Preview]        │
│  Status: Processing                                        │
│                                                             │
│  Progress                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │████████████████████████████░░░░░░░░░░░░░░░░░░  65%  │   │
│  └─────────────────────────────────────────────────────┘   │
│  Current Step: Synthesizing speech (TTS)                   │
│                                                             │
│  Steps                                                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ ✅ Extract Audio                                    │   │
│  │ ✅ Transcribe (ASR)                                 │   │
│  │ ✅ Translate                                        │   │
│  │ ⏳ Synthesize Speech (TTS)    3/10 segments         │   │
│  │ ⬜ Align Audio                                        │   │
│  │ ⬜ Generate Subtitles                                 │   │
│  │ ⬜ Compose Video                                      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Artifacts                                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 📄 transcript.zh.txt           [Download]           │   │
│  │ 📄 transcript.zh.srt           [Download]           │   │
│  │ 📄 script.en.txt               [Edit] [Download]    │   │
│  │ 🎵 narration.en.wav            [▶ Play] [Download]  │   │
│  │ 📄 subtitles.en-zh.ass         [Download]           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Logs                                                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 14:23:01 Starting audio extraction...               │   │
│  │ 14:23:05 Audio extracted: 45.2s                     │   │
│  │ 14:23:10 Transcription complete: 156 words          │   │
│  │ 14:23:15 Translation complete                       │   │
│  │ 14:23:20 Synthesizing segment 1/10...               │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**模板**: `web/templates/pages/task.html`

```html
{{template "base" .}}
{{define "content"}}
<main class="container">
    <nav>
        <a href="/">← Back</a>
        <code>{{.Task.ID}}</code>
    </nav>

    <!-- Task Info -->
    <article>
        <header>
            <h1>{{.Task.Filename}}</h1>
            <span class="badge {{.Task.Status}}">{{.Task.Status}}</span>
        </header>

        <!-- Progress -->
        <div hx-get="/api/tasks/{{.Task.ID}}/progress"
             hx-trigger="every 1s"
             hx-swap="innerHTML">
            {{template "progress-bar" .Task}}
        </div>
    </article>

    <!-- Steps -->
    <article>
        <h2>Steps</h2>
        <div hx-get="/api/tasks/{{.Task.ID}}/steps"
             hx-trigger="every 2s"
             hx-swap="innerHTML">
            {{template "step-list" .Task}}
        </div>
    </article>

    <!-- Artifacts -->
    <article>
        <h2>Artifacts</h2>
        {{template "artifact-list" .Task}}
    </article>

    <!-- Logs -->
    <details>
        <summary>Logs</summary>
        <pre id="logs"
             hx-get="/api/tasks/{{.Task.ID}}/logs"
             hx-trigger="every 3s"
             hx-swap="beforeend"
             style="max-height: 300px; overflow-y: auto;">
            {{range .Task.Logs}}
            <div>{{.Timestamp.Format "15:04:01"}} {{.Message}}</div>
            {{end}}
        </pre>
    </details>
</main>

<!-- WebSocket for real-time updates -->
<script>
    const ws = new WebSocket(`ws://${location.host}/ws/tasks/{{.Task.ID}}`);
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        // Update progress, steps, logs in real-time
        htmx.trigger('#progress', 'update', data);
        htmx.trigger('#steps', 'update', data);
    };
</script>
{{end}}
```

### 3. 设置页 (`/settings`)

```
┌─────────────────────────────────────────────────────────────┐
│  ← Back                                  Settings          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  MiMo API                                                   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ API Key: ••••••••••••••••            [Change]       │   │
│  │ Status: ✅ Connected                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Default Settings                                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Source Language:     [Auto Detect          ▾]       │   │
│  │ Target Language:     [English              ▾]       │   │
│  │ Default Voice:       [Chloe                ▾]       │   │
│  │ Audio Format:        [WAV                  ▾]       │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  FFmpeg                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Path: /usr/bin/ffmpeg                                │   │
│  │ Version: 6.1.1 ✅                                    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  [Save Settings]                                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## API 设计

### REST API

| Method | Path | 描述 | 返回 |
|--------|------|------|------|
| `POST` | `/api/upload` | 上传视频 | `303 → /task/:id` |
| `GET` | `/api/tasks` | 任务列表 | HTML 片段 |
| `GET` | `/api/tasks/:id` | 任务详情 | HTML 页面 |
| `GET` | `/api/tasks/:id/status` | 任务状态 | JSON |
| `GET` | `/api/tasks/:id/progress` | 进度条 | HTML 片段 |
| `GET` | `/api/tasks/:id/steps` | 步骤列表 | HTML 片段 |
| `GET` | `/api/tasks/:id/logs` | 日志 | HTML 片段 |
| `GET` | `/api/tasks/:id/artifacts` | 产物列表 | HTML 片段 |
| `POST` | `/api/tasks/:id/cancel` | 取消任务 | `204` |
| `DELETE` | `/api/tasks/:id` | 删除任务 | `204` |
| `GET` | `/api/tasks/:id/download/:name` | 下载产物 | 文件流 |
| `GET` | `/api/settings` | 获取设置 | JSON |
| `PUT` | `/api/settings` | 更新设置 | JSON |
| `POST` | `/api/settings/test` | 测试连接 | JSON |

### WebSocket

| Path | 描述 |
|------|------|
| `GET /ws/tasks/:id` | 任务实时进度 |

**消息格式**：

```json
{
    "type": "progress",
    "task_id": "abc123",
    "step": "tts",
    "progress": 0.65,
    "message": "Synthesizing segment 3/10",
    "timestamp": "2024-01-15T14:23:20Z"
}
```

## HTMX 交互模式

### 1. 表单提交 + 重定向

```html
<form hx-post="/api/upload"
      hx-encoding="multipart/form-data"
      hx-redirect="true">
    <!-- 上传后自动跳转到任务页 -->
</form>
```

### 2. 轮询更新

```html
<div hx-get="/api/tasks/{{.ID}}/progress"
     hx-trigger="every 1s"
     hx-swap="innerHTML">
    <!-- 每秒更新进度 -->
</div>
```

### 3. 条件停止轮询

```html
<div hx-get="/api/tasks/{{.ID}}/progress"
     hx-trigger="every 1s until .completed"
     hx-swap="innerHTML">
    <!-- 完成后停止轮询 -->
</div>
```

### 4. 点击加载

```html
<button hx-get="/api/tasks/{{.ID}}/logs"
        hx-target="#log-container"
        hx-swap="beforeend">
    Load More Logs
</button>
```

### 5. 确认操作

```html
<button hx-delete="/api/tasks/{{.ID}}"
        hx-confirm="Are you sure you want to delete this task?"
        hx-target="closest tr"
        hx-swap="outerHTML">
    Delete
</button>
```

## 静态资源嵌入

使用 Go 1.16+ 的 `embed` 包：

```go
// cmd/server/main.go
package main

import (
    "embed"
    "html/template"
    "net/http"
)

//go:embed web/templates/*
var templateFS embed.FS

//go:embed web/static/*
var staticFS embed.FS

func main() {
    tmpl := template.Must(template.ParseFS(templateFS, "web/templates/**/*.html"))
    
    http.Handle("/static/", http.FileServer(http.FS(staticFS)))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl.ExecuteTemplate(w, "home.html", nil)
    })
    
    http.ListenAndServe(":8080", nil)
}
```

## 样式方案

使用 Pico CSS（无类 CSS 框架）：

```html
<!-- web/templates/layouts/base.html -->
<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Video Localization</title>
    <link rel="stylesheet" href="/static/css/pico.min.css">
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="/static/js/htmx.min.js"></script>
</head>
<body>
    <main class="container">
        {{template "content" .}}
    </main>
</body>
</html>
```

自定义样式 (`web/static/css/style.css`)：

```css
/* Drop zone */
.drop-zone {
    border: 2px dashed var(--muted-border-color);
    border-radius: var(--border-radius);
    padding: 2rem;
    text-align: center;
    transition: all 0.2s;
}

.drop-zone.drag-over {
    border-color: var(--primary);
    background: var(--primary-focus);
}

/* Progress bar */
.progress-bar {
    height: 8px;
    background: var(--muted-border-color);
    border-radius: 4px;
    overflow: hidden;
}

.progress-bar .fill {
    height: 100%;
    background: var(--primary);
    transition: width 0.3s ease;
}

/* Status badges */
.badge {
    display: inline-block;
    padding: 0.25em 0.5em;
    font-size: 0.875em;
    border-radius: var(--border-radius);
}

.badge.pending { background: var(--muted-color); color: white; }
.badge.running { background: var(--primary); color: white; }
.badge.completed { background: var(--green); color: white; }
.badge.failed { background: var(--red); color: white; }

/* Step indicators */
.step { display: flex; align-items: center; gap: 0.5rem; }
.step-icon { font-size: 1.2em; }
.step.completed .step-icon { color: var(--green); }
.step.running .step-icon { color: var(--primary); animation: spin 1s linear infinite; }
.step.pending .step-icon { color: var(--muted-color); }

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}
```

## 文件上传处理

```go
// internal/handler/upload.go
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
    // 限制上传大小
    r.Body = http.MaxBytesReader(w, r.Body, h.config.Server.UploadLimit)
    
    // 解析表单
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "File too large", http.StatusBadRequest)
        return
    }
    
    file, header, err := r.FormFile("video")
    if err != nil {
        http.Error(w, "Invalid file", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // 保存文件
    taskID := generateTaskID()
    uploadPath := filepath.Join(h.config.UploadDir, taskID, header.Filename)
    os.MkdirAll(filepath.Dir(uploadPath), 0755)
    
    dst, _ := os.Create(uploadPath)
    io.Copy(dst, file)
    
    // 创建任务
    task := &task.Task{
        ID:         taskID,
        Filename:   header.Filename,
        SourcePath: uploadPath,
        Status:     task.StatusPending,
        CreatedAt:  time.Now(),
    }
    h.taskStore.Create(task)
    
    // 启动异步处理
    go h.pipeline.StartAsync(r.Context(), task)
    
    // HTMX 重定向
    w.Header().Set("HX-Redirect", fmt.Sprintf("/task/%s", taskID))
    w.WriteHeader(http.StatusSeeOther)
}
```

## 进度更新

```go
// internal/handler/websocket.go
func (h *Handler) TaskProgressWS(w http.ResponseWriter, r *http.Request) {
    taskID := mux.Vars(r)["id"]
    
    // 升级为 WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // 订阅进度更新
    ch := h.progress.Subscribe(taskID)
    defer h.progress.Unsubscribe(taskID, ch)
    
    // 发送当前状态
    task, _ := h.taskStore.Get(taskID)
    conn.WriteJSON(map[string]interface{}{
        "type":     "status",
        "status":   task.Status,
        "progress": task.Progress,
        "step":     task.CurrentStep,
    })
    
    // 监听更新
    for {
        select {
        case update := <-ch:
            if err := conn.WriteJSON(update); err != nil {
                return
            }
        case <-r.Context().Done():
            return
        }
    }
}
```

## 安全考虑

1. **文件上传**
   - 限制文件大小（默认 500MB）
   - 验证文件类型
   - 隔离上传目录

2. **API 认证**（可选）
   - API Key 认证
   - Session 认证

3. **CORS**
   - 默认禁止跨域
   - 可配置允许的源

4. **输入验证**
   - 验证所有用户输入
   - 防止路径遍历

## 部署方式

### 开发模式

```bash
# 启动服务器（带热重载）
go run cmd/server/main.go --dev

# 访问
open http://localhost:8080
```

### 生产模式

```bash
# 编译
go build -o video-localization-mimo .

# 运行
./video-localization-mimo server --port 8080

# 或使用 systemd
sudo systemctl start video-localization-mimo
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./server", "server", "--port", "8080"]
```
