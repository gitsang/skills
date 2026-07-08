#!/usr/bin/env python3
"""pdf2md: PDF text + image extraction with reading-order positioning for LLM translation.

Usage:
  python pdf2md.py <input.pdf> --chapters "Ch1:1-10,Ch2:11-30" [--out-dir ./output]
  python pdf2md.py <input.pdf> --toc                          # auto chapters from TOC

Output per chapter in out_dir/:
  chapters/ChX_source.md   ← text with <!-- page N -->, <!-- IMG src=... --> markers
  chapters/ChX_meta.json   ← line counts, image list
  images/                  ← embedded images renamed to ChX_pN_name.ext
"""

import pdfplumber, os, sys, json, re, subprocess, shutil, argparse
from pathlib import Path


def extract_toc(pdf):
    toc_pages = []
    for i, page in enumerate(pdf.pages[:15]):
        t = page.extract_text() or ""
        if re.search(r"Table\s+of\s+Contents|Contents", t, re.I):
            toc_pages.append(i)
    return toc_pages


def parse_chapters_from_toc(pdf, toc_page_indices):
    chapters = {}
    toc_text = ""
    for idx in toc_page_indices:
        toc_text += pdf.pages[idx].extract_text() or ""
    pattern = re.compile(r"^(\d+(?:\.\d+)?)\s+(.+?)\.{3,}\s*(\d+)\s*$", re.MULTILINE)
    for m in pattern.finditer(toc_text):
        num, name, page = m.group(1), m.group(2).strip(), int(m.group(3))
        parts = num.split(".")
        if len(parts) == 1:
            ch_key = f"Ch{num}"
            if ch_key not in chapters:
                chapters[ch_key] = {"name": name, "start": page, "subs": []}
    return chapters


def extract_page_content(page, page_num):
    chars = sorted(page.chars, key=lambda x: (x["top"], x["x0"]))
    lines = []
    current_line = []
    current_top = None
    for c in chars:
        top = round(c["top"], 1)
        if current_top is None or abs(top - current_top) <= 3:
            current_line.append(c)
            current_top = top
        else:
            lines.append(current_line)
            current_line = [c]
            current_top = top
    if current_line:
        lines.append(current_line)
    body_lines = []
    for l in lines:
        t = "".join(c["text"] for c in l).strip()
        top = l[0]["top"]
        size = l[0].get("size", 11)
        font = l[0].get("fontname", "")
        if 70 < top < 750 and t:
            is_heading = "Bold" in font and size > 11
            body_lines.append((top, t, is_heading, size))
    images = []
    for img in page.images:
        w, h = img["width"], img["height"]
        if h <= 3 or (w < 50 and h < 50) or (h <= 35 and w > 500):
            continue
        images.append(
            {
                "name": img.get("name", "?"),
                "x0": img["x0"],
                "top": img["top"],
                "x1": img["x1"],
                "bottom": img["bottom"],
                "width": w,
                "height": h,
            }
        )
    return body_lines, images


def extract_images_with_mutool(pdf_path, out_dir):
    img_dir = os.path.join(out_dir, "images")
    os.makedirs(img_dir, exist_ok=True)
    try:
        subprocess.run(
            ["mutool", "extract", "-r", pdf_path],
            cwd=img_dir,
            capture_output=True,
            timeout=60,
        )
        for f in os.listdir(img_dir):
            if f.startswith("font-"):
                os.remove(os.path.join(img_dir, f))
    except:
        print("Warning: mutool extraction failed")
    return img_dir


def build_reading_order(body_lines, images):
    elements = []
    for top, text, is_h, size in body_lines:
        elements.append(
            {
                "type": "heading" if is_h else "text",
                "y": top,
                "text": text,
                "size": size,
                "img_name": None,
            }
        )
    for img in images:
        elements.append(
            {
                "type": "image",
                "y": img["top"],
                "text": None,
                "img_name": img["name"],
                "img_info": img,
            }
        )
    elements.sort(key=lambda x: (x["y"], x["type"] != "heading"))
    return elements


def map_images_to_pages(pdf, image_dir, ch_start, ch_end):
    image_map = {}
    for f in os.listdir(image_dir):
        m = re.match(r"image-(\d+)\.(png|jpg)", f)
        if m:
            objid = int(m.group(1))
            for i in range(ch_start, ch_end):
                for img in pdf.pages[i].images:
                    stream = img.get("stream")
                    if not stream:
                        continue
                    oid = getattr(stream, "objid", None)
                    if oid == objid:
                        name = img.get("name", "?")
                        image_map[(i + 1, name)] = f
    return image_map


def main():
    parser = argparse.ArgumentParser(
        description="PDF to Markdown extraction for LLM translation"
    )
    parser.add_argument("input", help="Input PDF file")
    parser.add_argument("--out-dir", "-o", default="./pdf2md_output")
    parser.add_argument(
        "--chapters", "-c", help="Chapter defs, e.g. 'Ch1:1-10,Ch2:11-30'"
    )
    parser.add_argument("--toc", action="store_true", help="Auto chapters from TOC")
    args = parser.parse_args()

    out_dir = Path(args.out_dir)
    ch_dir = out_dir / "chapters"
    img_dir = out_dir / "images"
    ch_dir.mkdir(parents=True, exist_ok=True)
    print(f"Processing: {args.input}")

    with pdfplumber.open(args.input) as pdf:
        total_pages = len(pdf.pages)
        extract_images_with_mutool(args.input, str(out_dir))

        chapters = {}
        if args.chapters:
            for part in args.chapters.split(","):
                m = re.match(r"(\w+):(\d+)-(\d+)", part.strip())
                if m:
                    chapters[m.group(1)] = (int(m.group(2)) - 1, int(m.group(3)))
        if args.toc or not chapters:
            toc_indices = extract_toc(pdf)
            if toc_indices:
                parsed = parse_chapters_from_toc(pdf, toc_indices)
                for ch, info in parsed.items():
                    chapters[ch] = (info["start"] - 1, info["start"] + 9)
                print(f"Auto-detected {len(chapters)} chapters")

        for ch_name, (ch_start, ch_end) in sorted(chapters.items()):
            img_map = map_images_to_pages(pdf, str(img_dir), ch_start, ch_end)
            for (pg, name), fname in list(img_map.items()):
                src = img_dir / fname
                if src.exists():
                    ext = fname.rsplit(".", 1)[1]
                    dst = img_dir / f"{ch_name}_p{pg}_{name}.{ext}"
                    if not dst.exists():
                        shutil.move(str(src), str(dst))
                    img_map[(pg, name)] = dst.name

            all_text = []
            page_line_counts = []
            for i in range(ch_start, min(ch_end, total_pages)):
                page = pdf.pages[i]
                body_lines, images = extract_page_content(page, i + 1)
                elements = build_reading_order(body_lines, images)
                page_line_counts.append(
                    len([e for e in elements if e["type"] in ("text", "heading")])
                )
                all_text.append(f"<!-- page {i + 1} -->")
                for elem in elements:
                    if elem["type"] in ("text", "heading"):
                        all_text.append(elem["text"])
                    elif elem["type"] == "image":
                        key = (i + 1, elem["img_name"])
                        actual_file = img_map.get(key, "")
                        if actual_file:
                            all_text.append(f"<!-- IMG src={actual_file} -->")
                all_text.append("")

            (ch_dir / f"{ch_name}_source.md").write_text("\n".join(all_text))
            meta = {
                "pages": list(range(ch_start + 1, min(ch_end, total_pages) + 1)),
                "page_line_counts": page_line_counts,
                "images": [
                    {"page": pg, "file": fname} for (pg, _), fname in img_map.items()
                ],
                "total_lines": sum(page_line_counts),
            }
            (ch_dir / f"{ch_name}_meta.json").write_text(json.dumps(meta, indent=2))
            print(
                f"  {ch_name}: {len(page_line_counts)}p, {sum(page_line_counts)} lines, {len(img_map)} images"
            )

        metadata = {
            "source": args.input,
            "total_pages": total_pages,
            "chapters": {
                k: {"start": v[0] + 1, "end": v[1]} for k, v in chapters.items()
            },
        }
        (out_dir / "metadata.json").write_text(json.dumps(metadata, indent=2))
        print(f"Done -> {out_dir}/")


if __name__ == "__main__":
    main()
