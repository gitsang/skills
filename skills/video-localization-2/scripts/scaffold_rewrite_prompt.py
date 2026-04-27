from __future__ import annotations

import argparse
from pathlib import Path


CONTINUOUS_PROMPT = """把下面{source_language}视频字幕改写成适合{target_language}视频解说的口播稿。要求：
1. 保留原意
2. 使用自然、地道的{target_language}口语
3. 句子简短，适合配音
4. 不要书面腔
5. 只输出最终{target_language}稿，不要解释
"""

SEGMENT_PROMPT = """把下面{source_language}字幕逐条改写成{target_language}配音稿。要求：
1. 每个字幕块对应一行{target_language}
2. 保留原意，避免逐词硬译
3. 优先简洁、自然、适合配音
4. 不要输出编号、解释或空行
5. 最终输出的{target_language}行数必须和{source_language}字幕块数量一致
"""


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="生成目标语言改写用的 prompt scaffold。"
    )
    parser.add_argument(
        "--transcript", type=Path, required=True, help="源语言文本输入路径"
    )
    parser.add_argument("--output", type=Path, required=True, help="scaffold 输出路径")
    parser.add_argument("--source-language", default="英文", help="源语言名称")
    parser.add_argument("--target-language", default="中文", help="目标语言名称")
    parser.add_argument(
        "--mode",
        choices=["continuous", "segments"],
        default="continuous",
        help="continuous=整段目标语言稿；segments=逐段目标语言稿",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    transcript = args.transcript.read_text(encoding="utf-8").strip()
    prompt_template = CONTINUOUS_PROMPT if args.mode == "continuous" else SEGMENT_PROMPT
    prompt = prompt_template.format(
        source_language=args.source_language,
        target_language=args.target_language,
    )

    output = (
        "# Manual step required\n\n"
        "请将下面 prompt 交给你选定的 LLM，然后用最终目标语言稿覆盖本文件。\n\n"
        f"{prompt}\n\n=== TRANSCRIPT START ===\n{transcript}\n=== TRANSCRIPT END ===\n"
    )
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(output, encoding="utf-8")
    print(args.output)


if __name__ == "__main__":
    main()
