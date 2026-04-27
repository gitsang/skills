from __future__ import annotations

import argparse
import asyncio
import importlib
import json
import subprocess
import tempfile
import wave
from dataclasses import dataclass
from pathlib import Path
from typing import TYPE_CHECKING, Any, Protocol

if TYPE_CHECKING:
    import numpy as np  # type: ignore[import-untyped]


# numpy is imported lazily in main() to allow --help without numpy installed.
# TYPE_CHECKING guard lets type checkers resolve np.ndarray without a runtime import.


@dataclass
class SubtitleSegment:
    index: int
    start: float
    end: float
    source_text: str
    target_text: str


class QwenVoiceCloneModel(Protocol):
    def generate_voice_clone(
        self,
        text: str,
        language: str | None = None,
        ref_audio: str | None = None,
        ref_text: str | None = None,
        x_vector_only_mode: bool = False,
        non_streaming_mode: bool = False,
        **kwargs: object,
    ) -> tuple[list[np.ndarray], int]: ...


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build per-segment target-language dubbing aligned to source SRT."
    )
    parser.add_argument("--srt", type=Path, required=True, help="Source SRT path")
    parser.add_argument(
        "--tgt-segments",
        "--target-segments",
        "--translated-segments",
        dest="tgt_segments",
        type=Path,
        required=True,
        help="Per-segment target-language script path (one line per SRT block)",
    )
    parser.add_argument("--video", type=Path, required=True, help="Source video path")
    parser.add_argument(
        "--wav-out", type=Path, required=True, help="Aligned WAV output"
    )
    parser.add_argument(
        "--report-out", type=Path, required=True, help="Per-segment timing report JSON"
    )
    parser.add_argument(
        "--segment-dir", type=Path, required=True, help="Per-segment audio cache"
    )
    parser.add_argument(
        "--voice", help="edge-tts voice; required when --backend=edge-tts"
    )
    parser.add_argument("--rate", default="+0%", help="edge-tts speech rate")
    parser.add_argument(
        "--backend",
        choices=["edge-tts", "qwen3-tts"],
        default="edge-tts",
        help="TTS backend for per-segment synthesis",
    )
    parser.add_argument(
        "--language",
        "--target-lang",
        "--target-language",
        dest="language",
        default=None,
        help="Target language hint for Qwen3-TTS, e.g. Chinese, French, Japanese",
    )
    parser.add_argument(
        "--reference-audio",
        type=Path,
        help="Required when --backend=qwen3-tts: reference audio",
    )
    parser.add_argument(
        "--reference-text",
        type=Path,
        help="Required for qwen3-tts unless --x-vector-only is set",
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
        "--sample-rate", type=int, default=24000, help="Output WAV rate"
    )
    parser.add_argument(
        "--max-speedup",
        type=float,
        default=1.5,
        help="Maximum speed-up factor before trimming",
    )
    return parser.parse_args()


def parse_timestamp(value: str) -> float:
    hh, mm, rest = value.split(":")
    ss, ms = rest.split(",")
    return int(hh) * 3600 + int(mm) * 60 + int(ss) + int(ms) / 1000


def parse_srt(path: Path) -> list[tuple[int, float, float, str]]:
    blocks = path.read_text(encoding="utf-8").strip().split("\n\n")
    parsed: list[tuple[int, float, float, str]] = []
    for block in blocks:
        lines = [line.strip() for line in block.splitlines() if line.strip()]
        if len(lines) < 3:
            continue
        index = int(lines[0])
        start_raw, end_raw = lines[1].split(" --> ")
        text = " ".join(lines[2:])
        parsed.append(
            (index, parse_timestamp(start_raw), parse_timestamp(end_raw), text)
        )
    return parsed


def read_segments(srt_path: Path, tgt_segments_path: Path) -> list[SubtitleSegment]:
    source_segments = parse_srt(srt_path)
    target_lines = [
        line.rstrip("\n")
        for line in tgt_segments_path.read_text(encoding="utf-8").splitlines()
    ]
    if len(target_lines) != len(source_segments):
        raise SystemExit(
            f"Segment count mismatch: {len(source_segments)} srt blocks vs {len(target_lines)} target lines"
        )

    return [
        SubtitleSegment(index, start, end, source_text, target_text)
        for (index, start, end, source_text), target_text in zip(
            source_segments, target_lines, strict=True
        )
    ]


def ffprobe_duration(path: Path) -> float:
    result = subprocess.run(
        [
            "ffprobe",
            "-v",
            "error",
            "-show_entries",
            "format=duration",
            "-of",
            "default=noprint_wrappers=1:nokey=1",
            str(path),
        ],
        check=True,
        capture_output=True,
        text=True,
    )
    return float(result.stdout.strip())


async def synthesize_edge_segment(
    text: str, output_path: Path, voice: str, rate: str
) -> None:
    edge_tts_module = importlib.import_module("edge_tts")
    communicate = getattr(edge_tts_module, "Communicate")
    communicator = communicate(text, voice, rate=rate)
    await communicator.save(str(output_path))


def synthesize_qwen_segment(
    text: str,
    wav_output_path: Path,
    reference_audio: Path,
    reference_text: str | None,
    language: str | None,
    x_vector_only: bool,
    model: QwenVoiceCloneModel,
) -> None:
    soundfile = importlib.import_module("soundfile")
    kwargs: dict[str, Any] = dict(
        text=text,
        ref_audio=str(reference_audio),
        ref_text=reference_text,
        x_vector_only_mode=x_vector_only,
        non_streaming_mode=True,
    )
    if language is not None:
        kwargs["language"] = language
    wavs, sample_rate = model.generate_voice_clone(**kwargs)
    soundfile.write(str(wav_output_path), wavs[0], sample_rate)


def synthesize_with_selected_backend(
    *,
    backend: str,
    text: str,
    output_path: Path,
    voice: str | None,
    rate: str,
    language: str | None,
    reference_audio: Path | None,
    reference_text: str | None,
    x_vector_only: bool,
    qwen_model: QwenVoiceCloneModel | None,
) -> None:
    if backend == "edge-tts":
        if voice is None:
            raise SystemExit("--voice is required when --backend=edge-tts")
        asyncio.run(synthesize_edge_segment(text, output_path, voice, rate))
        return

    if reference_audio is None or qwen_model is None:
        raise SystemExit("--reference-audio is required when --backend=qwen3-tts")
    if reference_text is None and not x_vector_only:
        raise SystemExit("--reference-text is required unless --x-vector-only is set")

    synthesize_qwen_segment(
        text,
        output_path,
        reference_audio,
        reference_text,
        language,
        x_vector_only,
        qwen_model,
    )


def load_qwen_model(model_id: str) -> QwenVoiceCloneModel:
    torch = importlib.import_module("torch")
    qwen_tts_module = importlib.import_module("qwen_tts")
    qwen_model_cls = getattr(qwen_tts_module, "Qwen3TTSModel")
    device_map = "cuda:0" if torch.cuda.is_available() else "cpu"
    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    return qwen_model_cls.from_pretrained(model_id, device_map=device_map, dtype=dtype)


def mp3_to_wav(src: Path, dst: Path, sample_rate: int) -> None:
    subprocess.run(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(src),
            "-ac",
            "1",
            "-ar",
            str(sample_rate),
            "-c:a",
            "pcm_s16le",
            str(dst),
        ],
        check=True,
        capture_output=True,
        text=True,
    )


def read_wav_as_float(path: Path, sample_rate: int) -> np.ndarray:
    np = importlib.import_module("numpy")
    with wave.open(str(path), "rb") as wav_file:
        channels = wav_file.getnchannels()
        sample_width = wav_file.getsampwidth()
        frame_rate = wav_file.getframerate()
        frames = wav_file.readframes(wav_file.getnframes())

    if channels != 1 or sample_width != 2 or frame_rate != sample_rate:
        raise SystemExit(f"Unexpected WAV format for {path}")

    return np.frombuffer(frames, dtype=np.int16).astype(np.float32) / 32768.0


def write_wav_from_float(path: Path, samples: np.ndarray, sample_rate: int) -> None:
    np = importlib.import_module("numpy")
    clipped = np.clip(samples, -1.0, 1.0)
    pcm = (clipped * 32767.0).astype(np.int16)
    with wave.open(str(path), "wb") as wav_file:
        wav_file.setnchannels(1)
        wav_file.setsampwidth(2)
        wav_file.setframerate(sample_rate)
        wav_file.writeframes(pcm.tobytes())


def build_atempo_filter(speedup: float) -> str:
    factors: list[float] = []
    remaining = speedup
    while remaining > 2.0:
        factors.append(2.0)
        remaining /= 2.0
    while remaining < 0.5:
        factors.append(0.5)
        remaining /= 0.5
    factors.append(remaining)
    return ",".join(f"atempo={factor:.6f}" for factor in factors)


def speed_up_wav(src: Path, dst: Path, speedup: float, sample_rate: int) -> None:
    subprocess.run(
        [
            "ffmpeg",
            "-y",
            "-i",
            str(src),
            "-filter:a",
            build_atempo_filter(speedup),
            "-ac",
            "1",
            "-ar",
            str(sample_rate),
            "-c:a",
            "pcm_s16le",
            str(dst),
        ],
        check=True,
        capture_output=True,
        text=True,
    )


def main() -> None:
    args = parse_args()
    np = importlib.import_module("numpy")

    if args.backend == "edge-tts" and args.voice is None:
        raise SystemExit("--voice is required when --backend=edge-tts")
    if args.backend == "qwen3-tts" and args.reference_audio is None:
        raise SystemExit("--reference-audio is required when --backend=qwen3-tts")
    if (
        args.backend == "qwen3-tts"
        and args.reference_text is None
        and not args.x_vector_only
    ):
        raise SystemExit("--reference-text is required unless --x-vector-only is set")

    segments = read_segments(args.srt, args.tgt_segments)
    args.segment_dir.mkdir(parents=True, exist_ok=True)
    args.wav_out.parent.mkdir(parents=True, exist_ok=True)
    args.report_out.parent.mkdir(parents=True, exist_ok=True)

    video_duration = ffprobe_duration(args.video)
    total_samples = int(round(video_duration * args.sample_rate))
    final_audio = np.zeros(total_samples, dtype=np.float32)
    report: list[dict[str, object]] = []
    qwen_reference_text = (
        args.reference_text.read_text(encoding="utf-8").strip()
        if args.reference_text is not None
        else None
    )
    qwen_model: QwenVoiceCloneModel | None = None

    with tempfile.TemporaryDirectory(prefix="aligned-dub-") as temp_dir_str:
        temp_dir = Path(temp_dir_str)
        for idx, segment in enumerate(segments):
            start_sample = int(round(segment.start * args.sample_rate))
            next_start = (
                segments[idx + 1].start if idx + 1 < len(segments) else video_duration
            )
            slot_duration = max(0.1, next_start - segment.start)
            slot_samples = int(round(slot_duration * args.sample_rate))

            mp3_path = args.segment_dir / f"segment-{segment.index:03d}.mp3"
            wav_path = args.segment_dir / f"segment-{segment.index:03d}.wav"

            if args.backend == "qwen3-tts":
                if not wav_path.exists():
                    if qwen_model is None:
                        qwen_model = load_qwen_model(args.model_id)
                    synthesize_with_selected_backend(
                        backend=args.backend,
                        text=segment.target_text,
                        output_path=wav_path,
                        voice=args.voice,
                        rate=args.rate,
                        language=args.language,
                        reference_audio=args.reference_audio,
                        reference_text=qwen_reference_text,
                        x_vector_only=args.x_vector_only,
                        qwen_model=qwen_model,
                    )
            else:
                if not mp3_path.exists():
                    synthesize_with_selected_backend(
                        backend=args.backend,
                        text=segment.target_text,
                        output_path=mp3_path,
                        voice=args.voice,
                        rate=args.rate,
                        language=args.language,
                        reference_audio=args.reference_audio,
                        reference_text=qwen_reference_text,
                        x_vector_only=args.x_vector_only,
                        qwen_model=qwen_model,
                    )
                mp3_to_wav(mp3_path, wav_path, args.sample_rate)

            samples = read_wav_as_float(wav_path, args.sample_rate)
            original_samples = len(samples)
            speedup_applied = 1.0

            if len(samples) > slot_samples:
                requested_speedup = len(samples) / slot_samples
                speedup_applied = min(requested_speedup, args.max_speedup)
                sped_wav = temp_dir / f"segment-{segment.index:03d}-sped.wav"
                speed_up_wav(wav_path, sped_wav, speedup_applied, args.sample_rate)
                samples = read_wav_as_float(sped_wav, args.sample_rate)

            if len(samples) > slot_samples:
                samples = samples[:slot_samples]

            end_sample = min(total_samples, start_sample + len(samples))
            valid_samples = end_sample - start_sample
            if valid_samples > 0:
                final_audio[start_sample:end_sample] += samples[:valid_samples]

            report.append(
                {
                    "index": segment.index,
                    "start": segment.start,
                    "end": segment.end,
                    "slot_duration": slot_duration,
                    "source_text": segment.source_text,
                    "target_text": segment.target_text,
                    "backend": args.backend,
                    "language": args.language,
                    "x_vector_only": args.x_vector_only,
                    "generated_duration": round(original_samples / args.sample_rate, 3),
                    "placed_duration": round(valid_samples / args.sample_rate, 3),
                    "speedup_applied": round(speedup_applied, 3),
                    "truncated": original_samples > valid_samples,
                }
            )

    write_wav_from_float(args.wav_out, final_audio, args.sample_rate)
    args.report_out.write_text(
        json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(
        json.dumps(
            {"segments": len(report), "output": str(args.wav_out)}, ensure_ascii=False
        )
    )


if __name__ == "__main__":
    main()
