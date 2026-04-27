from __future__ import annotations

import argparse
from pathlib import Path


CONTINUOUS_PROMPT_TEMPLATE = (
    "Rewrite the following {src_lang} video transcript into a natural voiceover script in {tgt_lang}. "
    "Requirements:\n"
    "1. Preserve original meaning\n"
    "2. Use natural spoken {tgt_lang} — avoid literal translations\n"
    "3. Keep sentences short and suitable for dubbing\n"
    "4. Avoid overly formal or written register\n"
    "5. Output only the {tgt_lang} script, no explanations\n"
)

SEGMENT_PROMPT_TEMPLATE = (
    "Rewrite the following {src_lang} subtitles line-by-line into a {tgt_lang} dubbing script. "
    "Requirements:\n"
    "1. One subtitle block per output line\n"
    "2. Preserve original meaning, avoid word-for-word translation\n"
    "3. Prefer concise, natural phrasing suitable for dubbing\n"
    "4. Do not output numbering, explanations, or blank lines\n"
    "5. The number of output lines MUST equal the number of input subtitle blocks\n"
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate a human-review translation prompt scaffold."
    )
    parser.add_argument(
        "--transcript", type=Path, required=True, help="Source transcript input path"
    )
    parser.add_argument(
        "--output", type=Path, required=True, help="Scaffold output path"
    )
    parser.add_argument(
        "--source-lang",
        default="en",
        help="Source language code (e.g. en, zh, fr, ja)",
    )
    parser.add_argument(
        "--target-lang",
        default="zh",
        help="Target language code (e.g. zh, en, de, fr, ja, ko, es)",
    )
    parser.add_argument(
        "--mode",
        choices=["continuous", "segments"],
        default="continuous",
        help="continuous=full narration script; segments=per-segment dubbing script",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    transcript = args.transcript.read_text(encoding="utf-8").strip()

    if args.mode == "continuous":
        prompt = CONTINUOUS_PROMPT_TEMPLATE.format(
            src_lang=args.source_lang, tgt_lang=args.target_lang
        )
    else:
        prompt = SEGMENT_PROMPT_TEMPLATE.format(
            src_lang=args.source_lang, tgt_lang=args.target_lang
        )

    output = (
        "# Manual step required\n\n"
        f"Submit the prompt below to your LLM of choice, then overwrite this file "
        f"with the final {args.target_lang} script.\n\n"
        f"{prompt}\n\n=== TRANSCRIPT START ===\n{transcript}\n=== TRANSCRIPT END ===\n"
    )
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(output, encoding="utf-8")
    print(args.output)


if __name__ == "__main__":
    main()
