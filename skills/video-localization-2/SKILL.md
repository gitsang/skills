---
name: video-localization
description: Use when localizing a single YouTube, web, or local video into another language with translated voiceover, dubbing, subtitles, sentence-level alignment, final video deliverables, or reference-audio voice cloning.
---

# Video Localization

## 概述

将单个 YouTube、网页或本地视频本地化到另一种语言时，优先追求**可复现、可审计、真实交付**，而不是追求棚录级质量或黑盒自动化。

核心原则：先明确源语言 `{source_lang}` 与目标语言 `{target_lang}`，每一步都产出可检查文件，所有命令、问题、修复方式和中间产物沉淀到原始视频同级的工作目录中；最终交付物单独放到 `output/`。不要把中间产物、占位音频、总时长对齐或理论步骤当作完成。

支持两种配音模式：

- **标准目标语言旁白**：使用适合 `{target_lang}` 的 TTS 音色。
- **参考音频驱动的目标语言音色克隆**：默认从源视频中截取清晰人声作为参考；如果用户提供参考音频，则使用用户文件，并尽量要求对应参考文本。

明确边界：本 skill 不负责自动从复杂原视频中提纯声纹，不处理法律或授权风险高的“冒充原说话人”场景，也不承诺口型同步或原声百分百复刻。

## 什么时候用

适用于：

- 用户给出 YouTube 链接、网页视频或本地视频，要求翻译成本地化版本。
- 用户要求任意语言方向的视频翻译、配音、旁白、字幕或双语字幕，例如英文→中文、中文→英文、西班牙语→英文、法语→德语。
- 用户要求句子级同步、按原时间轴对齐、按标题命名的最终视频交付物。
- 用户要求音色克隆，希望默认使用源视频人声作为参考，或提供了单独参考音频。
- 需要在 TTS 前人工审阅目标语言稿，避免机械直译。

不适用于：

- 批量流水线处理大量视频。
- 目标是自动声纹提纯、声纹鉴权、口型同步或原声百分百复刻。
- 源视频、字幕、翻译稿或参考音频没有明确处理权限。
- 用户要求不可审计、不能保存中间证据的黑盒流程。

## 语言与命名约定

用 BCP-47 或简短语言码统一命名，贯穿转写、翻译、字幕和音频文件：

| 占位符 | 示例 | 说明 |
| --- | --- | --- |
| `{source_lang}` | `en`, `zh`, `es`, `fr` | 源视频主要语言 |
| `{target_lang}` | `zh`, `en`, `de`, `ja` | 目标旁白/字幕语言 |
| `{lang_pair}` | `en-zh`, `es-en` | 源语言到目标语言方向 |

最低建议产物：

| 产物 | 说明 |
| --- | --- |
| `source/video.mp4` | 标准化后的视频输入 |
| `source/audio.mp3` | 从视频抽取或用户提供的音频 |
| `artifacts/transcript.{source_lang}.txt` | 源语言转写文本 |
| `artifacts/transcript.{source_lang}.srt` | 带时间轴的源语言字幕 |
| `artifacts/script.{target_lang}.txt` | 连续目标语言口播稿 |
| `artifacts/script.{target_lang}.segments.txt` | 句子级同步时逐段目标语言稿，每行对应一个 SRT block |
| `source/reference.*` | 音色克隆模式下的参考音频和参考文本 |
| `artifacts/narration.{target_lang}*.wav/mp3` | 真实生成的目标语言 TTS 音频 |
| `artifacts/final*.mp4` | 候选本地化视频或内部合成产物 |
| `artifacts/subtitles*.ass/srt/vtt` | 内部字幕产物 |
| `output/{title}.{target_lang}.mp4` | 面向用户交付的干净目标语言旁白视频 |
| `output/{title}.{lang_pair}.ass` | 面向用户交付的外挂双语字幕文件 |
| `output/{title}.{lang_pair}-bilingual.mp4` | 面向用户交付的双语字幕烧录版视频 |
| `notes/commands.md` | 实际执行过的命令 |
| `notes/issues.md` | 阻塞、失败、修复方式和限制 |
| `report.md` | 面向用户的最终总结 |

## 工作目录与交付约定

每个任务必须先在原始视频同级创建独立工作目录。目录名使用原始视频文件名去掉扩展名后的 stem：

```text
/path/to/Source-Video-Name.mp4
/path/to/Source-Video-Name/
/path/to/Source-Video-Name/output/
```

工作目录内部结构：

```text
/path/to/Source-Video-Name/
  source/       # 原始和标准化输入
  artifacts/    # 转写、翻译稿、音频、字幕、视频等中间产物
  notes/        # 命令、问题、限制、人工检查记录
  output/       # 面向用户交付的最终文件
  scripts/      # 从 skill 复制过来的可复用脚本
  report.md     # 最终报告
```

`.venv`（Python 虚拟环境）和 `models/`（大模型权重目录）**不应放在工作目录内部**，而应放在原始视频所在父级目录中，以便多个视频任务共享复用。工作目录内部只保留该视频相关产物，不重复安装环境或下载模型。

## 推荐工作流

### 1. 获取并验证源素材

- 如果是 YouTube 或网页链接，先尝试合适下载工具（如 `yt-dlp`），但不要把 `429`、bot verification、登录验证或 CAPTCHA 误判为普通参数问题。
- 需要复核网页、登录态、网络请求或 CAPTCHA 时，按仓库浏览器规则优先使用 `chrome-devtools`。
- 如果浏览器也被拦截，停止绕路，改让用户提供本地源文件或字幕文件。
- 保存源获取方式、失败信息和最终采用的输入来源。

### 2. 标准化媒体输入

- 用 `ffprobe` 保存输入元信息，确认时长、视频流、音频流。
- 在工作目录内把任意用户文件名转换成稳定路径，例如 `source/video.mp4` 与 `source/audio.mp3`。
- 后续脚本只引用标准化路径，避免空格、中文、特殊字符导致脚本失败。

### 3. 转写源语言

- 优先使用适合 `{source_lang}` 的 ASR 工具输出纯文本和 SRT；多语言场景可用 `faster-whisper`。
- 转写完成后抽样检查质量，特别是专有名词、数字、断句和语言识别是否正确。
- 如果已有可靠字幕，可把字幕作为源转写输入，但必须记录来源和质量限制。

### 4. 生成目标语言口播稿

- 目标语言稿要自然、简洁、适合口播；保留原意优先于逐词直译。
- 连续旁白场景：准备 `script.{target_lang}.txt`。
- 句子级同步场景：准备 `script.{target_lang}.segments.txt`，每一行对应一个 SRT block。
- 这是人工审阅关口：确认文件里是最终目标语言稿，而不是 prompt scaffold。
- 不同目标语言长度差异很大；宁可适度改写，也不要机械直译导致无法对齐。

### 5. 准备参考音频（仅音色克隆模式）

默认从源视频中截取一段清晰单人讲话作为 `source/reference.wav`，并把该片段对应原文写入 `source/reference.txt`。如果用户提供参考音频，则改用用户文件，并尽量要求用户同时提供对应参考文本。

参考音频尽量满足：单人说话、背景噪声低、3-15 秒清晰人声、不混入音乐/多人对话/强混响、参考文本精确匹配。必须记录参考音频来源、截取时间范围或文件路径、文本来源和质量限制。

```bash
ffmpeg -y -ss 00:00:02.640 -to 00:00:15.000 \
  -i source/video.mp4 \
  -vn -ac 1 -ar 24000 -c:a pcm_s16le \
  source/reference.wav

printf '%s\n' 'Exact spoken text for this reference clip.' > source/reference.txt
```

### 6. 生成真实目标语言 TTS

- 选择支持 `{target_lang}` 的 TTS 后端；中文可优先考虑 `edge-tts` 普通话音色，音色克隆可使用支持 zero-shot voice cloning 的后端，例如 `Qwen3-TTS`。
- 其他目标语言应选择对应语言质量可靠的 TTS，不要沿用中文专用音色。
- 沉默文件、空白文件、占位文件不算完成。
- 生成后必须试听或用音频工具确认非静音、时长合理、采样格式可被 `ffmpeg` 使用。

### 7. 按原始时间轴对齐

- 如果用户在意句子级同步，不要先合成整段旁白再拉伸到总时长。
- 以原始 SRT 的 `start` 时间作为锚点，把每段目标语言 TTS 放回对应字幕块开始时间。
- 目标语言过长时，先有限度改写或加速，再在必要时裁切，并记录 timing report。
- “视频总时长没变”不等于“音画同步成功”。同步要看句子是否回到原始时间锚点。

### 8. 生成字幕产物

- 用户要求字幕时，优先输出独立字幕文件，再决定是否烧录。
- 双语字幕应复用源语言 SRT 时间轴，并把逐段目标语言稿按 block 合并进去。
- 区分外挂字幕、硬字幕、软字幕，不要用一个产物冒充另一个产物。

### 9. 合成最终视频

- 用目标语言旁白音轨替换或覆盖源视频音轨。
- 用 `ffprobe` 确认最终 MP4 同时包含视频流和音频流。
- 保留 clean 版和字幕版的生成命令，确保后续可重跑。

### 10. 交付包装与报告

- 不要停在 `final-voiceover-aligned.mp4` 这类内部文件名。
- 最终交付物必须复制或导出到 `output/`，不要只留在 `artifacts/`。
- 如果用户要求按标题命名，在 `output/` 中导出他们要求的最终文件名。
- 报告只写当前环境真实跑通的做法，不写想象中的理想路径。

常见交付集合：

```text
output/{title}.{target_lang}.mp4
output/{title}.{lang_pair}.ass
output/{title}.{lang_pair}-bilingual.mp4
```

## 内置脚本

优先把 skill 自带脚本复制到工作目录的 `scripts/` 下再运行，避免污染 skill 目录，也方便报告完整归档。

脚本接口使用通用语言命名：源语言输入使用 `--srt` / `--transcript`，目标语言逐段稿使用 `--target-segments`，目标语言字号使用 `--target-size`。不要再使用旧的语言专用参数名。

| 脚本 | 用途 |
| --- | --- |
| `scripts/transcribe_with_faster_whisper.py` | 把音频转成源语言文本和 SRT |
| `scripts/scaffold_rewrite_prompt.py` | 为目标语言改写生成可人工审阅的 prompt scaffold |
| `scripts/generate_edge_tts.py` | 把整段目标语言稿生成 MP3/WAV；必须显式指定适合目标语言的 `--voice` |
| `scripts/generate_qwen3_voice_clone.py` | 用参考音频和目标语言稿生成音色克隆 MP3/WAV；需确认后端支持目标语言 |
| `scripts/build_aligned_dub.py` | 逐段生成目标语言配音并按源字幕开始时间拼到整条时间线上 |
| `scripts/build_bilingual_ass.py` | 按源语言时间轴生成源/目标双语 ASS 字幕 |

中文目标语言示例：

```bash
python scripts/transcribe_with_faster_whisper.py \
  --audio source/audio.mp3 \
  --txt-out artifacts/transcript.en.txt \
  --srt-out artifacts/transcript.en.srt

python scripts/scaffold_rewrite_prompt.py \
  --transcript artifacts/transcript.en.txt \
  --output artifacts/script.zh.txt \
  --source-language 英文 \
  --target-language 中文 \
  --mode continuous

python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.en.srt \
  --target-segments artifacts/script.zh.segments.txt \
  --video source/video.mp4 \
  --backend edge-tts \
  --voice zh-CN-XiaoxiaoNeural \
  --wav-out artifacts/narration.zh.aligned.wav \
  --report-out artifacts/narration.zh.aligned.json \
  --segment-dir artifacts/aligned-segments

python scripts/build_bilingual_ass.py \
  --srt artifacts/transcript.en.srt \
  --target-segments artifacts/script.zh.segments.txt \
  --ass-out artifacts/subtitles.en-zh.ass
```

## 最小工具链

- `ffmpeg` 和 `ffprobe`：抽取音频、检查媒体流、合成最终视频。
- Python 虚拟环境：放在视频父级目录中共享使用，包含 ASR、TTS、字幕和音频处理依赖。
- ASR：`faster-whisper` 或其他支持 `{source_lang}` 的转写工具。
- TTS：选择支持 `{target_lang}` 的后端；中文普通旁白可用 `edge-tts`，音色克隆可评估 `qwen-tts` 等后端。
- `numpy`、`soundfile` 等音频处理依赖：逐段对齐和 WAV 拼接需要。

## 人工审阅关口

在这些节点必须停下来检查真实产物：

- 源获取失败后，确认是否是网络、登录态、CAPTCHA 或会话级阻塞。
- 转写完成后，抽样检查源语言稿和 SRT block 数。
- 翻译/改写完成后，确认是最终目标语言稿，不是提示词草稿。
- 做逐段同步时，确认逐段目标语言稿行数与 SRT block 数一致。
- 做音色克隆时，确认参考音频存在、可播放、单人清晰、授权清楚。
- TTS 完成后，确认输出是真实语音，不是静音或占位。
- 合成完成后，用 `ffprobe` 确认视频流和音频流都存在。
- 交付前，确认最终命名文件真实存在于 `output/` 并符合用户要求。

## 故障排查

### 下载源视频遇到 `429`、bot verification 或 CAPTCHA

不要把这类问题当作普通参数错误反复重试。先用浏览器复核页面状态；若浏览器也被 CAPTCHA 或登录拦住，停止绕路，改让用户提供本地媒体或字幕文件，并把失败信息写入 `notes/issues.md`。

### Hugging Face、YouTube、PyPI 等网络请求超时

先验证代理是否真的生效，记录失败域名、错误和代理变量。代理恢复后重跑原命令；如果只有某个域名不可达，优先换镜像源或本地缓存，而不是更换整个工作流。

### faster-whisper 模型下载失败或无本地缓存

先检查代理和缓存目录：`~/.cache/huggingface`、`~/.cache/ctranslate2`、`~/.cache/whisper`。若仍失败，可让用户提供这个视频对应的 SRT/转写，或改用已经安装且可离线运行的 ASR 工具；不要手写伪造 SRT。

### TTS 后端不支持目标语言

表现为发音错误、直接读拼音/罗马音、输出乱码或静音。更换支持 `{target_lang}` 的 TTS 后端，不要把中文专用音色用于非中文目标语言，也不要把失败音频当作完成。

### 音色克隆不像参考音色

优先检查参考音频质量：是否单人、是否 3-15 秒、是否有音乐/混响/多人重叠、参考文本是否精确匹配。源视频默认参考段不理想时，重新选择更干净的人声片段；用户参考音频不理想时，请用户提供更干净样本。

### 目标语言旁白比视频短很多

如果只在意总时长，可以补静音；如果在意句子同步，必须逐段对齐，不能只把整段音频拉伸到视频长度。

### 视频长度对了但句子越说越漂

连续旁白不能保证句子同步。改用原始 SRT 的 start time 作为锚点逐段构建配音，并保存 timing report；目标语言过长时记录改写、加速或裁切情况。

### 用户要求按标题命名但只有内部 artifact

增加显式 packaging/export 步骤，把最终文件复制或导出到 `output/`，使用用户要求的标题命名，并验证文件真实存在。

### 用户要求字幕但只有视频文件

补出 `.ass`、`.srt`、`.vtt`、硬字幕视频或用户指定格式。报告中要区分外挂字幕、硬字幕和软字幕，不要用一个产物冒充另一个。

### 文件名含空格导致 ffmpeg 失败

始终对含空格的路径加引号。标准化路径时不要去除空格（那是文件名的一部分），只需在命令行中正确引用。

### 逐段稿行数与 SRT block 数不一致

逐段目标语言稿的每一行对应 SRT 中的一个字幕块。生成目标语言稿时，确保行数与 SRT block 数严格一致。如果翻译中某段被合并或拆分，需要调整逐段稿使其一一对应。

## 红旗

任一条成立，就不要宣称完成：

- 没有明确 `{source_lang}`、`{target_lang}` 和交付语言方向。
- 只有 prompt scaffold，没有最终目标语言稿。
- TTS 后端不支持目标语言，或输出是沉默文件、空白文件、占位文件。
- 宣称做了音色克隆，但没有合法参考音频来源或质量记录。
- 只做了总时长对齐，却称为句子级同步成功。
- 最终视频从未用 `ffprobe` 复核。
- 报告声称下载成功，但实际没有跑通。
- 用户要求最终命名交付，但 `output/` 里没有最终文件。
- 用户要求字幕交付，但只交了视频。
- 工作目录的 `notes/` 没有记录命令、失败和限制。

## 完成前检查清单

- [ ] 已确认源语言、目标语言、字幕语言组合和用户要求的交付格式。
- [ ] 工作目录位于原始视频同级，且目录名等于原始视频 stem。
- [ ] 工作目录内部结构完整，并包含 `output/`。
- [ ] 源视频、音频、转写、目标语言稿、TTS、字幕和最终视频都有明确路径。
- [ ] 关键命令写入 `notes/commands.md`。
- [ ] 阻塞点、降级方案、质量限制写入 `notes/issues.md` 或 `report.md`。
- [ ] 如果逐段同步，已保存 timing report 并说明改写、加速或裁切情况。
- [ ] 如果音色克隆，已记录参考音频来源、质量、授权和限制。
- [ ] 最终 MP4 经过 `ffprobe` 验证含视频流和音频流。
- [ ] 用户要求的最终文件名、字幕形式和包装形式全部存在于 `output/`。
