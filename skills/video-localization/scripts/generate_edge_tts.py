from __future__ import annotations

import argparse
import asyncio
import importlib
import json
import subprocess
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Synthesize a narration script to MP3/WAV using edge-tts."
    )
    parser.add_argument("--script", type=Path, required=True, help="Narration script")
    parser.add_argument("--mp3-out", type=Path, required=True, help="MP3 output")
    parser.add_argument("--wav-out", type=Path, required=True, help="WAV output")
    parser.add_argument(
        "--voice",
        required=True,
        help="edge-tts voice matching target language; run edge-tts --list-voices",
    )
    parser.add_argument("--rate", default="+0%", help="Speech rate, e.g. +8%%")
    return parser.parse_args()


async def synthesize(text: str, voice: str, rate: str, mp3_out: Path) -> None:
    edge_tts_module = importlib.import_module("edge_tts")
    communicate = getattr(edge_tts_module, "Communicate")
    communicator = communicate(text, voice, rate=rate)
    await communicator.save(str(mp3_out))


def convert_to_wav(mp3_out: Path, wav_out: Path) -> None:
    subprocess.run(["ffmpeg", "-y", "-i", str(mp3_out), str(wav_out)], check=True)


def main() -> None:
    args = parse_args()
    text = args.script.read_text(encoding="utf-8").strip()
    if not text:
        raise SystemExit("Script file is empty")

    args.mp3_out.parent.mkdir(parents=True, exist_ok=True)
    args.wav_out.parent.mkdir(parents=True, exist_ok=True)

    asyncio.run(synthesize(text, args.voice, args.rate, args.mp3_out))
    convert_to_wav(args.mp3_out, args.wav_out)
    print(
        json.dumps(
            {"mp3": str(args.mp3_out), "wav": str(args.wav_out), "voice": args.voice},
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
