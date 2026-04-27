from __future__ import annotations

import argparse
import importlib
import json
from pathlib import Path


def format_ts(seconds: float) -> str:
    milliseconds = int(round(seconds * 1000))
    hours, rem = divmod(milliseconds, 3_600_000)
    minutes, rem = divmod(rem, 60_000)
    secs, millis = divmod(rem, 1000)
    return f"{hours:02d}:{minutes:02d}:{secs:02d},{millis:03d}"


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Transcribe audio to plain text and SRT using faster-whisper."
    )
    parser.add_argument("--audio", type=Path, required=True, help="Input audio file")
    parser.add_argument("--txt-out", type=Path, required=True, help="Plain text output")
    parser.add_argument("--srt-out", type=Path, required=True, help="SRT output")
    parser.add_argument(
        "--model",
        default="small",
        help="Whisper model name or local path (tiny/base/small/medium/large)",
    )
    parser.add_argument(
        "--language",
        "--source-lang",
        dest="language",
        default=None,
        help="Source language code, e.g. en, zh, fr, ja; omit for auto-detection",
    )
    parser.add_argument("--device", default="cpu", help="Inference device: cpu or cuda")
    parser.add_argument(
        "--compute-type", default="int8", help="faster-whisper compute_type"
    )
    parser.add_argument("--beam-size", type=int, default=5, help="Beam size")
    parser.add_argument(
        "--disable-vad-filter",
        action="store_true",
        help="Disable VAD filter, which is enabled by default",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    args.txt_out.parent.mkdir(parents=True, exist_ok=True)
    args.srt_out.parent.mkdir(parents=True, exist_ok=True)

    faster_whisper_module = importlib.import_module("faster_whisper")
    whisper_model = getattr(faster_whisper_module, "WhisperModel")
    model = whisper_model(
        args.model, device=args.device, compute_type=args.compute_type
    )

    transcribe_kwargs: dict[str, object] = {
        "beam_size": args.beam_size,
        "vad_filter": not args.disable_vad_filter,
    }
    if args.language is not None:
        transcribe_kwargs["language"] = args.language

    segments, info = model.transcribe(str(args.audio), **transcribe_kwargs)
    segment_list = list(segments)

    args.txt_out.write_text(
        "\n".join(
            segment.text.strip() for segment in segment_list if segment.text.strip()
        )
        + "\n",
        encoding="utf-8",
    )

    with args.srt_out.open("w", encoding="utf-8") as handle:
        output_index = 1
        for segment in segment_list:
            text = segment.text.strip()
            if not text:
                continue
            handle.write(f"{output_index}\n")
            handle.write(f"{format_ts(segment.start)} --> {format_ts(segment.end)}\n")
            handle.write(f"{text}\n\n")
            output_index += 1

    print(
        json.dumps(
            {
                "language": info.language,
                "duration": info.duration,
                "segments": len(segment_list),
                "txt_out": str(args.txt_out),
                "srt_out": str(args.srt_out),
            },
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
