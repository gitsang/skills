---
name: pdf2md
description: Use when converting PDF files to Markdown format, extracting text and images from PDFs while preserving reading order and chapter structure. Triggers on: “pdf 转 markdown”, “pdf to md”, “extract pdf to markdown”, “pdf 提取”, “pdf 转换”, or any task involving converting a PDF document into clean, readable Markdown. 只转换，不翻译。
---

# PDF 转 Markdown（pdf2md）

从 PDF 中提取文字和图片，按阅读顺序输出为 Markdown，保留章节结构和页面标记。

核心原则：**还原 PDF 的阅读顺序，不打乱、不翻译、不丢弃。**

## 工作流（两阶段）

```
阶段 1: pdf2md 提取  →  _source.md (保留所有原始标记，只读)
阶段 2: LLM 格式化   →  ChX.md (去除 page 标记，IMG 转 md 语法)
阶段 3: 合并         →  合并文档 (调整图片路径为 images/)
```

### ⚠️ 铁律（长期教训总结）

| 序号 | 规则 | 原因 |
|------|------|------|
| 1 | **永远不要修改 `_source.md` 文件** | 它们是唯一的原始参照，弄丢了只能重新提取 PDF |
| 2 | **subagent prompt 必须用绝对路径** | `Output: Ch4.md` 会导致 agent 写错位置，用 `Output: /full/path/Ch4.md` |
| 3 | **操作前验证 `_source.md` 有标记** | 确认 `grep '<!-- page ' *_source.md` 行数正确后再格式化 |
| 4 | **sed/批量操作严格排除 `_source.md`** | 用 `for f in Ch[0-9]*.md; do [[ $f == *_source.md ]] && continue; ...` 不可靠时改用显式列表 |
| 5 | **先确认 PDF 准确页数再定章节范围** | 不同版本的同一 PDF 页数可能不同（303 vs 305），用 `pdfplumber` 先 `len(pdf.pages)` |

### 1. 确认章节范围

两种方式：

- **手动指定**：`--chapters "Ch1:1-10,Ch2:11-30"` —— 格式为 `章节名:起始页-结束页`
- **自动检测**：`--toc` —— 从目录页自动解析章节

如果不指定，脚本会自动尝试从目录页检测。

### 2. 运行提取

```bash
# 手动指定章节
python scripts/pdf2md.py input.pdf --chapters "Ch1:1-10,Ch2:11-20" -o output_dir/

# 自动从目录检测
python scripts/pdf2md.py input.pdf --toc -o output_dir/
```

### 3. 检查输出

最终输出目录结构：

```
output_dir/
├── metadata.json                            # 源文件、总页数、章节信息
├── images/                                  # 提取的图片（按章节重命名）
│   ├── Ch1_p5_Im0.png
│   └── Ch2_p15_Im0.jpg
└── chapters/
    ├── Ch1_source.md                        # ⚠️ 只读！pdf2md 原始提取
    ├── Ch1_meta.json
    ├── Ch1.md                               # LLM 格式化后（无 page 标记，IMG→![]()）
    ├── Ch2_source.md                        # ⚠️ 只读！
    ├── Ch2_meta.json
    └── Ch2.md                               # LLM 格式化后
```

**两套文件职责：**

| 文件 | 角色 | 页标记 | 图标记 | 修改 |
|------|------|--------|--------|------|
| `ChX_source.md` | 原始参照 | `<!-- page N -->` | `<!-- IMG src=X -->` | **永远不改** |
| `ChX.md` | 可读成品 | 无 | `![](../images/X)` | 可编辑 |

### 4. 格式化（LLM 处理 `_source.md` → `ChX.md`）

提取后 `_source.md` 是原始文本（pdfplumber 提取的粗糙输出），需要用 subagent 格式化为干净的 Markdown。

**标记转换规则（核心区别）：**

| 标记 | `_source.md`（保留） | `ChX.md`（格式化后） |
|------|---------------------|----------------------|
| `<!-- page N -->` | ✅ 保留 | ❌ 移除 |
| `<!-- IMG src=xxx -->` | ✅ 保留 | `![](../images/xxx)` |

**subagent prompt 模板：**

```
Read /absolute/path/ChX_source.md, format it, WRITE to /absolute/path/ChX.md.

Rules:
1. REMOVE all <!-- page N --> lines entirely
2. CONVERT <!-- IMG src=FILENAME --> to ![](../images/FILENAME)
3. Join broken mid-sentence lines into paragraphs
4. Fix tables to | col | col | markdown format
5. Add heading hierarchy: ## X. Title → ### X.Y → #### X.Y.Z
6. Remove per-page header/footer boilerplate
7. English only, NO TRANSLATION
8. No commentary or notes
```

**大章节拆分**：>1500 行的 `_source.md` 分成多个 subagent 并行处理，最后用 `cat part1.md part2.md > ChX.md` 合并。

**格式化完成后验证**：
```bash
# _source.md 标记必须完好
grep -c '<!-- page ' chapters/Ch*_source.md
# Ch*.md 无残留标记
grep -rl '<!-- page ' chapters/Ch*.md    # 应该无输出
grep -rl '<\!-- IMG' chapters/Ch*.md     # 应该无输出
```

### 5. 合并文档

将所有 `ChX.md` 合并为一个完整文档，加上目录：

```bash
{
  echo "# Document Title"
  echo "## Table of Contents"
  for f in chapters/Ch*.md; do
    title=$(head -3 "$f" | grep '^## ' | head -1 | sed 's/^## //')
    printf -- "- [%s: %s](#%s)\n" "$(basename $f .md)" "$title" "$(basename $f .md)"
  done
  for f in chapters/Ch*.md; do cat "$f" && echo; done
} > merged.md

# 修正图片路径：../images/ → images/（合并文档在根目录）
sed -i 's|](/\.\./images/|](/images/|g' merged.md
```

### 6. 后处理（图片插入，可选）

如果还需要精确按比例插入图片（如翻译后行数变化），用 postprocess：

```bash
python scripts/pdf2md-postprocess.py <处理后的.md> <meta.json> <images目录>
```

原理：根据 `meta.json` 中的每页行数和图片位置，按比例计算图片在目标文档中的插入点。

## 快速参考

| 操作 | 命令 |
|------|------|
| 基本提取 | `python scripts/pdf2md.py input.pdf -o out/` |
| 指定章节 | `python scripts/pdf2md.py input.pdf --chapters "Ch1:1-10" -o out/` |
| 自动目录检测 | `python scripts/pdf2md.py input.pdf --toc -o out/` |
| 图片后处理 | `python scripts/pdf2md-postprocess.py translated.md meta.json images/` |

## 依赖

```bash
pip install pdfplumber
```

`mutool` 用于图片提取（可选，提取失败时图片位会保留但缺文件）：

```bash
# Ubuntu/Debian
sudo apt install mupdf-tools
# macOS
brew install mupdf-tools
```

## 常见问题

| 问题 | 原因 | 处理 |
|------|------|------|
| `_source.md` 标记被清空 | sed/批量操作误伤了源文件 | **永久教训**：永远不批量操作 `_source.md`，用显式列表排除 |
| subagent 输出文件缺失 | prompt 用了相对路径 `Output: Ch4.md` | **总是用绝对路径** |
| PDF 提取失败 | 章节页码范围超出 PDF 实际页数 | 先用 `len(pdf.pages)` 确认总页数再定范围 |
| 章节检测不准 | 目录格式不标准 | 改用 `--chapters` 手动指定页码 |
| 图片提取失败 | 缺少 `mutool` | 安装 mupdf-tools 或接受缺图 |
| 正文顺序错乱 | 多栏排版 | pdfplumber 按 top→x0 排序，多栏可能需要调整 |
| 标题未识别 | 字体无 Bold 属性 | 检查字体嵌入，或后期手动标注 |
| 后处理图片位置偏移 | LLM 处理改变了行数分布 | 按比例插值是近似值，手工微调 |

## 使用边界

- **只负责格式转换，不翻译内容。** 如果需要翻译，那是另一个工作流。
- 适合**文字为主的 PDF**（论文、报告、书籍）。扫描件需要先 OCR。
- 多栏排版阅读顺序**可能不完美**，取决于 pdfplumber 的字符排序。
- 复杂表格不还原为 Markdown 表格，保持原始文字排列。
