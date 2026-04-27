from __future__ import annotations

import argparse
from pathlib import Path


def build_prompt(source_lang: str, target_lang: str, mode: str) -> str:
    """Build a language-agnostic translation prompt."""
    if mode == "continuous":
        return (
            f"Translate the following video subtitles from {source_lang} to {target_lang} "
            f"for a narration-style voiceover. Requirements:\n"
            f"1. Preserve the original meaning\n"
            f"2. Use natural, conversational {target_lang}\n"
            f"3. Keep sentences short and suitable for voice dubbing\n"
            f"4. Avoid overly formal or written style\n"
            f"5. Output only the {target_lang} translation, no explanations"
        )
    else:
        return (
            f"Translate each subtitle line below from {source_lang} to {target_lang} for voice dubbing. Requirements:\n"
            f"1. Each subtitle block corresponds to exactly one line of {target_lang}\n"
            f"2. Preserve meaning, avoid literal word-by-word translation\n"
            f"3. Prioritize concise, natural, dubbing-friendly output\n"
            f"4. Do not output numbers, explanations, or blank lines\n"
            f"5. The final {target_lang} line count must match the source subtitle block count"
        )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate a reviewable prompt scaffold for target-language translation."
    )
    parser.add_argument(
        "--transcript", type=Path, required=True, help="Source-language text input path"
    )
    parser.add_argument(
        "--output", type=Path, required=True, help="Scaffold output path"
    )
    parser.add_argument(
        "--source-lang", default="en", help="Source language (e.g., en, ja, ko, es)"
    )
    parser.add_argument(
        "--target-lang",
        default="zh",
        help="Target language (e.g., zh, en, ja, ko, es, fr, de)",
    )
    parser.add_argument(
        "--mode",
        choices=["continuous", "segments"],
        default="continuous",
        help="continuous=full script; segments=per-segment translation",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    transcript = args.transcript.read_text(encoding="utf-8").strip()
    prompt = build_prompt(args.source_lang, args.target_lang, args.mode)

    output = (
        "# Manual step required\n\n"
        f"Please submit the prompt below to your chosen LLM, then overwrite this file "
        f"with the final {args.target_lang} translation.\n\n"
        f"{prompt}\n\n"
        f"=== TRANSCRIPT START ===\n{transcript}\n=== TRANSCRIPT END ===\n"
    )
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(output, encoding="utf-8")
    print(args.output)


if __name__ == "__main__":
    main()
