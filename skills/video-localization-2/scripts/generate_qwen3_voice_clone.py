from __future__ import annotations

import argparse
import importlib
import subprocess
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="使用 Qwen3-TTS 与参考音频生成目标语言音色克隆语音。"
    )
    parser.add_argument("--script", type=Path, required=True, help="目标语言稿路径")
    parser.add_argument(
        "--reference-audio", type=Path, required=True, help="参考音色音频路径"
    )
    parser.add_argument(
        "--reference-text", type=Path, required=True, help="参考音频对应文本路径"
    )
    parser.add_argument("--mp3-out", type=Path, required=True, help="MP3 输出路径")
    parser.add_argument("--wav-out", type=Path, required=True, help="WAV 输出路径")
    parser.add_argument(
        "--model-id",
        default="Qwen/Qwen3-TTS-12Hz-0.6B-Base",
        help="Qwen3-TTS Base 模型 ID",
    )
    parser.add_argument(
        "--sample-rate", type=int, default=24000, help="输出 MP3 时使用的采样率"
    )
    parser.add_argument(
        "--target-language", required=True, help="传给 Qwen3-TTS 的目标语言名称"
    )
    return parser.parse_args()


def synthesize(
    script: Path,
    reference_audio: Path,
    reference_text: Path,
    wav_out: Path,
    model_id: str,
    target_language: str,
) -> int:
    text = script.read_text(encoding="utf-8").strip()
    if not text:
        raise SystemExit("Target-language script is empty")

    ref_text = reference_text.read_text(encoding="utf-8").strip()
    if not ref_text:
        raise SystemExit("Reference text is empty")

    torch = importlib.import_module("torch")
    soundfile = importlib.import_module("soundfile")
    qwen_tts_module = importlib.import_module("qwen_tts")
    qwen_model = getattr(qwen_tts_module, "Qwen3TTSModel")

    device_map = "cuda:0" if torch.cuda.is_available() else "cpu"
    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    model = qwen_model.from_pretrained(model_id, device_map=device_map, dtype=dtype)
    wavs, sample_rate = model.generate_voice_clone(
        text=text,
        language=target_language,
        ref_audio=str(reference_audio),
        ref_text=ref_text,
        non_streaming_mode=True,
    )
    soundfile.write(str(wav_out), wavs[0], sample_rate)
    return sample_rate


def convert_to_mp3(wav_out: Path, mp3_out: Path, sample_rate: int) -> None:
    subprocess.run(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(wav_out),
            "-ar",
            str(sample_rate),
            str(mp3_out),
        ],
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
        args.target_language,
    )
    convert_to_mp3(args.wav_out, args.mp3_out, sample_rate or args.sample_rate)
    print(
        {
            "reference_audio": str(args.reference_audio),
            "reference_text": str(args.reference_text),
            "mp3": str(args.mp3_out),
            "wav": str(args.wav_out),
        }
    )


if __name__ == "__main__":
    main()
