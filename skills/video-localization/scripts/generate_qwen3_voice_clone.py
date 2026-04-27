from __future__ import annotations

import argparse
import importlib
import json
import subprocess
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Synthesize a script using Qwen3-TTS reference-audio voice cloning."
    )
    parser.add_argument(
        "--script", type=Path, required=True, help="Target-language script"
    )
    parser.add_argument(
        "--reference-audio", type=Path, required=True, help="Reference audio"
    )
    parser.add_argument(
        "--reference-text",
        type=Path,
        help="Reference-audio transcript; required unless --x-vector-only is set",
    )
    parser.add_argument("--mp3-out", type=Path, required=True, help="MP3 output")
    parser.add_argument("--wav-out", type=Path, required=True, help="WAV output")
    parser.add_argument(
        "--language",
        "--target-lang",
        "--target-language",
        dest="language",
        default=None,
        help="Target language hint for Qwen3-TTS, e.g. Chinese, French, Japanese",
    )
    parser.add_argument(
        "--x-vector-only",
        action="store_true",
        help="Use Qwen3-TTS x-vector-only mode without reference-text prompt",
    )
    parser.add_argument(
        "--model-id",
        default="Qwen/Qwen3-TTS-12Hz-0.6B-Base",
        help="Qwen3-TTS model ID or local path",
    )
    parser.add_argument(
        "--sample-rate", type=int, default=24000, help="MP3 sample rate"
    )
    return parser.parse_args()


def synthesize(
    script: Path,
    reference_audio: Path,
    reference_text: Path | None,
    wav_out: Path,
    model_id: str,
    language: str | None,
    x_vector_only: bool,
) -> int:
    text = script.read_text(encoding="utf-8").strip()
    if not text:
        raise SystemExit("Script file is empty")

    ref_text = None
    if reference_text is not None:
        ref_text = reference_text.read_text(encoding="utf-8").strip()
    if ref_text is None and not x_vector_only:
        raise SystemExit("--reference-text is required unless --x-vector-only is set")

    torch = importlib.import_module("torch")
    soundfile = importlib.import_module("soundfile")
    qwen_tts_module = importlib.import_module("qwen_tts")
    qwen_model = getattr(qwen_tts_module, "Qwen3TTSModel")

    device_map = "cuda:0" if torch.cuda.is_available() else "cpu"
    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    model = qwen_model.from_pretrained(model_id, device_map=device_map, dtype=dtype)

    kwargs: dict[str, Any] = {
        "text": text,
        "ref_audio": str(reference_audio),
        "ref_text": ref_text,
        "x_vector_only_mode": x_vector_only,
        "non_streaming_mode": True,
    }
    if language is not None:
        kwargs["language"] = language

    wavs, sample_rate = model.generate_voice_clone(**kwargs)
    soundfile.write(str(wav_out), wavs[0], sample_rate)
    return int(sample_rate)


def convert_to_mp3(wav_out: Path, mp3_out: Path, sample_rate: int) -> None:
    subprocess.run(
        ["ffmpeg", "-y", "-i", str(wav_out), "-ar", str(sample_rate), str(mp3_out)],
        check=True,
    )


def main() -> None:
    args = parse_args()
    args.mp3_out.parent.mkdir(parents=True, exist_ok=True)
    args.wav_out.parent.mkdir(parents=True, exist_ok=True)

    sample_rate = synthesize(
        args.script,
        args.reference_audio,
        args.reference_text,
        args.wav_out,
        args.model_id,
        args.language,
        args.x_vector_only,
    )
    convert_to_mp3(args.wav_out, args.mp3_out, sample_rate or args.sample_rate)
    print(
        json.dumps(
            {
                "reference_audio": str(args.reference_audio),
                "reference_text": str(args.reference_text)
                if args.reference_text
                else None,
                "language": args.language,
                "x_vector_only": args.x_vector_only,
                "mp3": str(args.mp3_out),
                "wav": str(args.wav_out),
            },
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
