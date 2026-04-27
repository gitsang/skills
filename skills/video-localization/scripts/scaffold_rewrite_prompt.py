from __future__ import annotations

import argparse
from pathlib import Path


def build_prompt(source_lang: str, target_lang: str, mode: str) -> str:
    if mode == "continuous":
        return (
            f"Rewrite the following {source_lang} video transcript into a natural "
            f"voiceover script in {target_lang}. Requirements:\n"
            "1. Preserve the original meaning\n"
            f"2. Use natural spoken {target_lang}; avoid literal translation\n"
            "3. Keep sentences short and suitable for dubbing\n"
            "4. Avoid overly formal or written register\n"
            f"5. Output only the final {target_lang} script, no explanations"
        )

    return (
        f"Rewrite each {source_lang} subtitle block into one {target_lang} dubbing "
        "line. Requirements:\n"
        f"1. Each input subtitle block corresponds to exactly one {target_lang} line\n"
        "2. Preserve meaning and avoid word-for-word translation\n"
        "3. Prefer concise, natural phrasing suitable for dubbing\n"
        "4. Do not output numbering, explanations, or blank lines\n"
        "5. The output line count MUST match the input subtitle block count"
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate a human-review translation prompt scaffold."
    )
    parser.add_argument(
        "--transcript", type=Path, required=True, help="Input transcript"
    )
    parser.add_argument("--output", type=Path, required=True, help="Scaffold output")
    parser.add_argument(
        "--source-lang",
        "--source-language",
        dest="source_lang",
        default="source language",
        help="Source language name, e.g. English, Chinese, Spanish",
    )
    parser.add_argument(
        "--target-lang",
        "--target-language",
        dest="target_lang",
        default="target language",
        help="Target language name, e.g. Chinese, English, French",
    )
    parser.add_argument(
        "--mode",
        choices=["continuous", "segments"],
        default="continuous",
        help="continuous=full narration script; segments=one target line per SRT block",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    transcript = args.transcript.read_text(encoding="utf-8").strip()
    prompt = build_prompt(args.source_lang, args.target_lang, args.mode)

    output = (
        "# Manual step required\n\n"
        "Submit the prompt below to your chosen LLM, review the result, then overwrite "
        f"this file with the final {args.target_lang} script. This scaffold is not "
        "a deliverable.\n\n"
        f"{prompt}\n\n=== TRANSCRIPT START ===\n{transcript}\n=== TRANSCRIPT END ===\n"
    )
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(output, encoding="utf-8")
    print(args.output)


if __name__ == "__main__":
    main()
