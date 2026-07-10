---
name: pdf2md
description: Use when converting PDF files to Markdown format, extracting text and images from PDFs while preserving reading order and chapter structure. Triggers on: “pdf 转 markdown”, “pdf to md”, “extract pdf to markdown”, “pdf 提取”, “pdf 转换”, or any task involving converting a PDF document into clean, readable Markdown. 只转换，不翻译。
---

# PDF → Markdown 转换管线

将 PDF 文档拆解为结构化 Markdown：提取文字和图片，按章分文件，保留阅读顺序和页面参照。

处理管线分为四个阶段，每个阶段有明确的输入、输出和验证方式：

```
PDF 文件
  │
  ├─[步骤 1] 确认章节范围
  ├─[步骤 2] pdf2md 提取 → _source.md + images
  ├─[步骤 3] Subagent 格式化 → ChX.md
  └─[步骤 4] 合并文档 → merged.md
```

## 双文件模式

整个管线的基础是两个文件各司其职：

| 文件 | 来源 | 作用 | 内容特征 |
|------|------|------|----------|
| `ChX_source.md` | pdf2md 脚本自动生成 | 不可更改的原始参照 | 保留 `<!-- page N -->` 页标记和 `<!-- IMG src=... -->` 图标记 |
| `ChX.md` | Subagent 从 `_source.md` 格式化 | 可读的成品 Markdown | 页标记已移除，图标记转为 `![](../images/xxx)` |

**关键约定**：所有修改只发生在 `ChX.md` 上。`_source.md` 文件从生成那一刻起就不再变更——它是后续翻译、重新格式化的唯一可信来源。一旦 `_source.md` 损坏，只能重新从 PDF 提取。

## 步骤 1：确认 PDF 和章节范围

拿到 PDF 后，先确认两个信息：

### 确认总页数

```bash
python3 -c "import pdfplumber; pdf=pdfplumber.open('input.pdf'); print(len(pdf.pages))"
```

同一份文档的不同版本页数可能不同（例如 303 页 vs 305 页），直接用旧数据会导致提取失败。

### 确认章节页码范围

两种方式获取章节的起止页码：

- **手动浏览**：打印前几页和后几页的文本，对照目录找到每章起始页
  ```bash
  python3 -c "
  import pdfplumber
  with pdfplumber.open('input.pdf') as pdf:
      for i in [0,1,2,3,4,5, len(pdf.pages)-4, len(pdf.pages)-3, len(pdf.pages)-2, len(pdf.pages)-1]:
          t = pdf.pages[i].extract_text() or ''
          print(f'=== Page {i+1} ===')
          print(t[:200])
  "
  ```

- **自动目录检测**：使用 `--toc` 参数让脚本从目录页自动解析，适合目录格式标准的 PDF

汇总为 `--chapters` 参数：`"Ch1:起始页-结束页,Ch2:起始页-结束页,..."`。

## 步骤 2：提取 → _source.md + images

```bash
python scripts/pdf2md.py input.pdf \
  --chapters "Ch1:9-13,Ch2:14-36,Ch3:37-128,Ch4:129-155,Ch5:156-217,Ch6:218-239,Ch7:240-280,Ch8:281-301,Ch9:302-302,Ch10:303-303" \
  -o output_dir/
```

产出目录结构：

```
output_dir/
├── metadata.json
├── images/
│   ├── Ch1_p9_img2.jpg        # 命名: Ch{章节}_p{页码}_img{序号}
│   └── Ch3_p48_img31.png
└── chapters/
    ├── Ch1_source.md           # 原始提取（只读）
    ├── Ch1_meta.json           # 行数、图片位置元数据
    └── ...
```

**提取后立即验证**：确认每个 `_source.md` 的页标记和图标记数量合理：

```bash
for f in chapters/Ch*_source.md; do
  echo "$(basename $f): $(grep -c '<!-- page ' $f) pages, $(grep -c '<\!-- IMG' $f) images"
done
```

页数应等于章节页码范围差值，图数应 > 0（如果章节有配图）。如果全是 0，说明提取失败或页码范围有误，返回步骤 1 修正。

## 步骤 3：格式化 → ChX.md

`_source.md` 是 pdfplumber 的粗糙输出——换行断裂、表格散乱、夹杂页眉页脚。用 subagent 将每章格式化为干净的可读 Markdown。

### 分派策略

**小章**（< 1500 行）：每个源文件分派一个 subagent，输出 `ChX.md`。

**大章**（≥ 1500 行）：拆成 2-3 份，每份约 700-1000 行，并行处理。例如 2800 行的 Ch3：
```
Ch3_source.md (2800 行)
  → Ch3_part1.md (L1-933)    ─┐
  → Ch3_part2.md (L934-1866)  ├─ 并行 subagent
  → Ch3_part3.md (L1867-2800) ─┘
      ↓
    cat Ch3_part{1,2,3}.md > Ch3.md
```

### Subagent 提示词

给每个 subagent 的指令包含：要读的源文件、要写的目标文件（**完整绝对路径**）、格式化规则。

```
Read /absolute/path/ChX_source.md.
Write formatted output to /absolute/path/ChX.md.

Format the raw PDF extraction into clean markdown by doing these:
- Remove every <!-- page N --> comment line from the text
- Replace each <!-- IMG src=FILENAME --> with an image link: ![](../images/FILENAME)
- Join lines that were split mid-sentence by pdfplumber into continuous paragraphs
- Convert broken table text into proper markdown tables (| col | col | format)
- Apply heading hierarchy: ## X. Chapter Title → ### X.Y Section → #### X.Y.Z Subsection
- Strip repeated page headers and footers from the body text
- Keep all original text in English — no translation, no summarization, no commentary
```

**要点说明**：

- 页标记 `<!-- page N -->` 仅在 `_source.md` 中有意义——它表示原始 PDF 的页边界。格式化后的文档不需要这些标记，直接删除整行即可。
- 图标记 `<!-- IMG src=xxx -->` 转为标准 Markdown 图片语法。因为 `ChX.md` 位于 `chapters/` 子目录，图片路径用 `../images/xxx`。
- Subagent 只负责格式整理，不翻译、不增删实质性内容。

### 格式化后验证

所有 subagent 完成后，批量检查结果：

```bash
# ChX.md 中不应残留任何 <!-- page 或 <!-- IMG 标记
grep -rl '<!-- page ' chapters/Ch*.md | grep -v _source    # 应无输出
grep -rl '<\!-- IMG' chapters/Ch*.md | grep -v _source     # 应无输出

# _source.md 的标记必须保持完好（确认没被误改）
grep -c '<!-- page ' chapters/Ch*_source.md
```

## 步骤 4：合并文档

所有 `ChX.md` 就绪后，合并为单一 Markdown 文件，附带自动生成的目录：

```bash
output_file="merged.md"

{
  echo "# Document Title"
  echo ""
  echo "## Table of Contents"
  for f in chapters/Ch*.md; do
    # 从文件中提取章节标题
    title=$(head -5 "$f" | grep '^## ' | head -1 | sed 's/^## //')
    ch="$(basename "$f" .md)"
    printf -- "- [%s: %s](#%s)\n" "$ch" "$title" "$ch"
  done
  echo ""
  for f in chapters/Ch*.md; do
    cat "$f"
    echo ""
  done
} > "$output_file"

# 图片路径从 ../images/ 修正为 images/（合并文档放在根目录）
sed -i 's|(../images/|(images/|g' "$output_file"
```

## 补充：图片按比例后处理

如果文档经过翻译或大幅结构调整，行数比例发生了变化，可以用后处理脚本根据原始 `_meta.json` 中的位置信息按比例重新插入图片：

```bash
python scripts/pdf2md-postprocess.py <处理后的.md> <ChX_meta.json> <images目录>
```

原理：元数据记录了每张图在原始 PDF 中的页码和在该页的相对位置。后处理脚本按当前文档与该页总行数的比例，计算图片的新插入位置。

---

## 快速参考

| 操作 | 命令 |
|------|------|
| 查看 PDF 总页数 | `python3 -c "import pdfplumber; print(len(pdfplumber.open('f.pdf').pages))"` |
| 提取（指定章节） | `python scripts/pdf2md.py input.pdf --chapters "Ch1:1-10" -o out/` |
| 提取（自动目录） | `python scripts/pdf2md.py input.pdf --toc -o out/` |
| 图片后处理 | `python scripts/pdf2md-postprocess.py doc.md meta.json images/` |

## 依赖安装

```bash
pip install pdfplumber
```

`mutool` 用于图片提取（可选，缺少时图片位保留占位但无文件）：

```bash
sudo apt install mupdf-tools    # Debian/Ubuntu
brew install mupdf-tools        # macOS
```

## 已知限制

- 多栏排版的阅读顺序取决于 pdfplumber 的字符排序策略，可能不完全准确。
- 复杂表格不还原为 Markdown 表格结构，保持 pdfplumber 输出的原始文字排列。
- 图片提取依赖 `mutool`，未安装时图片不会提取但位置标记保留。
