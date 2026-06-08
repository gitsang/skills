# Video Localization MiMo

基于小米 MiMo API 的视频本地化工具，使用 Go 语言实现，支持 CLI 和 Web 两种使用方式。

## 概述

将视频从一种语言本地化为另一种语言，包括：
- 语音识别（ASR）- 使用 MiMo-V2.5-ASR
- 翻译润色 - 使用 MiMo LLM
- 语音合成（TTS）- 使用 MiMo-V2.5-TTS 系列
- 语音克隆 - 使用 MiMo-V2.5-TTS-VoiceClone
- 字幕生成 - 本地处理
- 视频合成 - ffmpeg

## 两种使用方式

### CLI 模式
```bash
video-localization-mimo transcribe --audio source.mp3 --lang zh
video-localization-mimo localize --video source.mp4 --src zh --tgt en
```

### Web 模式
```bash
video-localization-mimo server --port 8080
# 浏览器访问 http://localhost:8080
```

## 相比原 Python 方案的优势

| 方面 | Python 方案 | Go + MiMo API |
|------|-------------|---------------|
| 依赖 | Python + PyTorch + CUDA + 模型文件 | Go 二进制 + ffmpeg |
| 安装 | virtualenv + pip + 模型下载（数 GB） | 单二进制 + API Key |
| GPU | 本地需要 | 不需要（云端推理） |
| 维护 | 模型版本、依赖冲突 | API 版本稳定 |
| 分发 | 复杂 | 单文件 |
| UI | 无 | Web UI（可选） |

## 快速开始

```bash
# 安装
go install github.com/gitsang/skills/video-localization-mimo@latest

# 配置 API Key
export MIMO_API_KEY="your-api-key"

# CLI 使用
video-localization-mimo transcribe --audio source.mp3 --lang zh
video-localization-mimo translate --input transcript.zh.txt --src zh --tgt en
video-localization-mimo tts --script script.en.txt --voice Chloe --output narration.mp3
video-localization-mimo compose --video source.mp4 --audio narration.mp3 --output final.mp4

# Web 使用
video-localization-mimo server --port 8080
```

## 文档

- [架构设计](architecture.md)
- [API 规范](api-spec.md)
- [Web 应用设计](web-design.md)
- [实现计划](implementation.md)

## 系统要求

- Go 1.21+
- ffmpeg（用于音视频处理）
- 网络连接（调用 MiMo API）
- MiMo API Key（从 [platform.xiaomimimo.com](https://platform.xiaomimimo.com) 获取）

## 许可证

MIT
