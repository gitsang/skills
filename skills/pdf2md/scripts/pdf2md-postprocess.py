#!/usr/bin/env python3
"""After LLM translation: insert images at proportional positions using page-line-count ratio.

Usage:
  python pdf2md-postprocess.py <translated.md> <meta.json> <images_dir>
"""

import json, os, re, sys
from pathlib import Path


def load_markdown_elements(path):
    with open(path) as f:
        content = f.read()
    elements = []
    current = []
    for line in content.split("\n"):
        stripped = line.strip()
        if line.startswith("#") or stripped == "":
            if current:
                elements.append("\n".join(current))
                current = []
            if line.startswith("#"):
                elements.append(line)
        else:
            current.append(line)
    if current:
        elements.append("\n".join(current))
    return elements, content


def main():
    if len(sys.argv) < 4:
        print("Usage: pdf2md-postprocess.py <translated.md> <meta.json> <images_dir>")
        sys.exit(1)

    md_path, meta_path, img_dir = sys.argv[1:4]
    with open(meta_path) as f:
        meta = json.load(f)

    elements, _ = load_markdown_elements(md_path)
    total_lines = meta["total_lines"]
    total_elements = len([e for e in elements if not e.startswith("![")

    images = meta["images"]
    page_line_counts = meta["page_line_counts"]
    cumulative_per_page = [0]
    for c in page_line_counts:
        cumulative_per_page.append(cumulative_per_page[-1] + c)

    image_positions = []
    for img in images:
        pg = img["page"]
        fname = img["file"]
        page_idx = pg - meta["pages"][0]
        if 0 <= page_idx < len(cumulative_per_page) - 1:
            pos = cumulative_per_page[page_idx + 1] - page_line_counts[page_idx] // 2
        else:
            pos = cumulative_per_page[-1] // 2
        image_positions.append((pos, pg, fname))

    image_positions.sort(key=lambda x: -x[0])
    for pos, pg, fname in image_positions:
        pct = pos / total_lines if total_lines else 0
        idx = min(int(pct * len(elements)), len(elements) - 1)
        ref = f"![p{pg}](images/{fname})"
        elements.insert(idx, ref)

    result = "\n\n".join(e for e in elements).strip()
    result = re.sub(r"\n{4,}", "\n\n\n", result)

    output = md_path.replace(".md", "_with_images.md")
    with open(output, "w") as f:
        f.write(result + "\n")

    img_count = len(re.findall(r"!\[p\d+\]\(images/", result))
    print(f"Output: {output}")
    print(f"Images inserted: {img_count}")


if __name__ == "__main__":
    main()
