from __future__ import annotations

import argparse
import json
from pathlib import Path


ASS_HEADER_TEMPLATE = """[Script Info]
ScriptType: v4.00+
PlayResX: {play_res_x}
PlayResY: {play_res_y}
WrapStyle: 2
ScaledBorderAndShadow: yes

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Bilingual,{font_name},{font_size},&H00FFFFFF,&H000000FF,&H00101010,&H64000000,0,0,0,0,100,100,0,0,1,2,0,2,{margin_l},{margin_r},{margin_v},1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
"""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate bilingual ASS subtitles from source SRT and target segments."
    )
    parser.add_argument("--srt", type=Path, required=True, help="Source SRT path")
    parser.add_argument(
        "--tgt-segments",
        "--target-segments",
        "--translated-segments",
        dest="tgt_segments",
        type=Path,
        required=True,
        help="Per-segment target-language script path",
    )
    parser.add_argument("--ass-out", type=Path, required=True, help="ASS output path")
    parser.add_argument(
        "--lang-pair",
        default=None,
        help=(
            "Language pair code used in the output filename (e.g. en-zh, fr-de). "
            "Informational only; does not affect subtitle content."
        ),
    )
    parser.add_argument(
        "--font-name",
        default="Sans",
        help=(
            "Font name. Use a CJK-capable font (e.g. 'Noto Sans CJK SC') "
            "when target or source language is Chinese, Japanese, or Korean."
        ),
    )
    parser.add_argument("--font-size", type=int, default=36, help="Base font size")
    parser.add_argument(
        "--target-size",
        "--tgt-size",
        dest="target_size",
        type=int,
        default=40,
        help="Target-language subtitle font size",
    )
    parser.add_argument(
        "--source-size",
        "--src-size",
        dest="source_size",
        type=int,
        default=28,
        help="Source-language subtitle font size",
    )
    parser.add_argument("--play-res-x", type=int, default=1920, help="ASS canvas width")
    parser.add_argument(
        "--play-res-y", type=int, default=1080, help="ASS canvas height"
    )
    parser.add_argument("--margin-l", type=int, default=80, help="Left margin")
    parser.add_argument("--margin-r", type=int, default=80, help="Right margin")
    parser.add_argument("--margin-v", type=int, default=48, help="Bottom margin")
    return parser.parse_args()


def parse_timestamp(value: str) -> tuple[int, int, int, int]:
    hh, mm, rest = value.split(":")
    ss, ms = rest.split(",")
    return int(hh), int(mm), int(ss), int(ms)


def to_ass_timestamp(value: str) -> str:
    hh, mm, ss, ms = parse_timestamp(value)
    centiseconds = round(ms / 10)
    if centiseconds == 100:
        centiseconds = 0
        ss += 1
        if ss == 60:
            ss = 0
            mm += 1
            if mm == 60:
                mm = 0
                hh += 1
    return f"{hh}:{mm:02d}:{ss:02d}.{centiseconds:02d}"


def escape_ass_text(value: str) -> str:
    return value.replace("\\", r"\\").replace("{", r"\{").replace("}", r"\}")


def parse_srt_blocks(path: Path) -> list[tuple[str, str, str]]:
    blocks = path.read_text(encoding="utf-8").strip().split("\n\n")
    items: list[tuple[str, str, str]] = []
    for block in blocks:
        lines = [line.strip() for line in block.splitlines() if line.strip()]
        if len(lines) < 3:
            continue
        start_raw, end_raw = lines[1].split(" --> ")
        text = " ".join(lines[2:])
        items.append((start_raw, end_raw, text))
    return items


def main() -> None:
    args = parse_args()
    srt_entries = parse_srt_blocks(args.srt)
    target_lines = [
        line.rstrip("\n")
        for line in args.tgt_segments.read_text(encoding="utf-8").splitlines()
    ]
    if len(srt_entries) != len(target_lines):
        raise SystemExit(
            f"Segment count mismatch: {len(srt_entries)} subtitle blocks vs {len(target_lines)} target lines"
        )

    header = ASS_HEADER_TEMPLATE.format(
        play_res_x=args.play_res_x,
        play_res_y=args.play_res_y,
        font_name=args.font_name,
        font_size=args.font_size,
        margin_l=args.margin_l,
        margin_r=args.margin_r,
        margin_v=args.margin_v,
    )
    lines = [header]
    for (start_raw, end_raw, source_text), target_text in zip(
        srt_entries, target_lines, strict=True
    ):
        start_ass = to_ass_timestamp(start_raw)
        end_ass = to_ass_timestamp(end_raw)
        target_escaped = escape_ass_text(target_text)
        source_escaped = escape_ass_text(source_text)
        text = (
            f"{{\\fs{args.target_size}}}{target_escaped}"
            f"\\N{{\\fs{args.source_size}}}{source_escaped}"
        )
        lines.append(f"Dialogue: 0,{start_ass},{end_ass},Bilingual,,0,0,0,,{text}")

    args.ass_out.parent.mkdir(parents=True, exist_ok=True)
    args.ass_out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(
        json.dumps(
            {
                "ass_out": str(args.ass_out),
                "lang_pair": args.lang_pair,
                "blocks": len(srt_entries),
            },
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
