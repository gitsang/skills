---
name: video-localization
description: Use when localizing a single YouTube, web, or local video into another language with translated dubbing, sentence-level alignment, bilingual subtitles, final deliverables, or authorized reference-audio voice cloning.
---

# Video Localization

## Overview

Localize one video into another language with **reproducible, auditable, real deliverables**. The workflow favors verifiable artifacts over black-box automation or studio-grade promises.

Core principle: every step must leave evidence. Save commands, failures, fixes, transcripts, scripts, generated audio, subtitles, timing reports, and final outputs in a work directory next to the source video. Do not claim completion from prompt scaffolds, placeholder audio, total-duration-only alignment, or theoretical steps.

Supported modes:

- **Standard target-language dubbing**: synthesize narration with a TTS voice that explicitly supports the target language.
- **Reference-audio voice cloning**: use authorized reference audio. Default to a clean 3-15 second single-speaker clip extracted from the source video only when appropriate; switch to user-provided reference audio if supplied.

Out of scope: batch pipelines, automatic voiceprint purification, voice authentication, lip sync, rights bypassing, or guaranteeing 100% speaker replication.

## When to Use

Use this when:

- User provides a YouTube URL, web video, or local video and asks for translation, dubbing, subtitles, or localized video output.
- The task needs source-language transcription, target-language narration, sentence-level sync, bilingual subtitles, or title-named final files.
- User asks for voice cloning with a clear reference-audio source and acceptable usage rights.
- Manual script review is needed before TTS to avoid literal or awkward translation.

Do not use this when:

- Processing many videos as a batch system.
- The source, subtitles, translated script, or reference audio lack clear processing permission.
- User requests unaudited processing with no intermediate artifacts.
- The goal is lip-sync, identity verification, or impersonation without authorization.

## Language and Naming Conventions

Confirm the language direction before producing final artifacts:

| Placeholder | Example | Meaning |
|---|---|---|
| `{src_lang}` | `en`, `zh`, `es`, `fr` | Main source-video language |
| `{tgt_lang}` | `zh`, `en`, `de`, `ja` | Target narration/subtitle language |
| `{lang_pair}` | `en-zh`, `es-en` | Source-to-target direction |

Use ISO 639-1, BCP-47, or clear language names consistently. For `edge-tts`, use a concrete voice locale such as `zh-CN-XiaoxiaoNeural`, `en-US-AriaNeural`, `fr-FR-DeniseNeural`, or `ja-JP-NanamiNeural`; run `edge-tts --list-voices` to choose a voice matching `{tgt_lang}`.

## Work Directory Convention

Create an isolated work directory next to the source video. The directory name is the source filename stem:

```text
/path/to/Source-Video.mp4
/path/to/Source-Video/
/path/to/Source-Video/output/
```

Internal layout:

```text
/path/to/Source-Video/
  source/       # normalized inputs and reference audio/text
  artifacts/    # transcripts, scripts, TTS, subtitles, intermediate video
  notes/        # commands, issues, manual checks, constraints
  output/       # user-facing final files
  scripts/      # copied skill scripts
  report.md     # final summary
```

Keep shared `.venv/` and `models/` in the source video's parent directory, not inside the per-video work directory, so multiple localization tasks can reuse dependencies and model weights. Ensure that parent-level `.gitignore` ignores `.venv/` and `models/`.

Minimum artifacts:

| Artifact | Description |
|---|---|
| `source/video.mp4` | Normalized video input |
| `source/audio.mp3` | Extracted or user-provided audio |
| `artifacts/transcript.{src_lang}.txt` | Source-language transcript |
| `artifacts/transcript.{src_lang}.srt` | Timed source-language SRT |
| `artifacts/script.{tgt_lang}.txt` | Continuous target-language narration script |
| `artifacts/script.{tgt_lang}.segments.txt` | One target-language line per SRT block |
| `source/reference.wav` and `source/reference.txt` | Reference audio/text for voice cloning |
| `artifacts/narration.{tgt_lang}*.wav/mp3` | Real generated target-language audio |
| `artifacts/narration.{tgt_lang}.aligned.json` | Segment timing report |
| `artifacts/subtitles.{lang_pair}.ass` | Internal bilingual subtitle artifact |
| `output/{title}.{tgt_lang}.mp4` | Clean localized video |
| `output/{title}.{lang_pair}.ass` | External bilingual subtitle file |
| `output/{title}.{lang_pair}-bilingual.mp4` | Burned bilingual subtitle video |
| `notes/commands.md` | Actually executed commands |
| `notes/issues.md` | Blockers, fixes, fallbacks, constraints |
| `report.md` | User-facing summary of what actually ran |

## Recommended Workflow

### 1. Acquire and verify source media

- For YouTube or web links, try `yt-dlp` first, but do not treat `429`, bot verification, login checks, or CAPTCHA as ordinary parameter errors.
- If browser state, login, network requests, or CAPTCHA need inspection, use the repository's browser tooling rules.
- If browser access is also blocked, stop bypass attempts and ask the user for a local media or subtitle file.
- Record acquisition method, failures, final source path, and source permissions.

### 2. Normalize media inputs

- Use `ffprobe` to save metadata and confirm duration, video stream, and audio stream.
- Copy or transcode to stable paths: `source/video.mp4` and `source/audio.mp3`.
- Quote paths with spaces. Prefer canonical internal paths for scripts.

### 3. Transcribe source audio

- Prefer `faster-whisper` for text and SRT.
- Specify `--language`/`--source-lang` when the source language is known; otherwise allow auto-detection.
- Sample-check transcript quality: names, numbers, technical terms, sentence breaks, and SRT block count.
- If model download fails, check proxy/cache first; do not fabricate subtitles.

### 4. Produce target-language script

- Write natural, concise, spoken target-language narration; preserve meaning over literal translation.
- For continuous narration, create `artifacts/script.{tgt_lang}.txt`.
- For sentence-level sync, create `artifacts/script.{tgt_lang}.segments.txt` with exactly one line per SRT block.
- Treat `scaffold_rewrite_prompt.py` output as a manual prompt scaffold, not the final script. Overwrite it with reviewed final target-language text before TTS.

### 5. Prepare reference audio for voice cloning

Default source-extracted reference:

```bash
ffmpeg -y -ss 00:00:02.640 -to 00:00:15.000 \
  -i source/video.mp4 \
  -vn -ac 1 -ar 24000 -c:a pcm_s16le \
  source/reference.wav

printf '%s\n' 'Exact spoken text for this reference clip.' > source/reference.txt
```

Reference audio should be single-speaker, low-noise, 3-15 seconds, no music/crosstalk/reverb, and paired with exact reference text unless using an explicit x-vector-only mode. Record time range or file path, source, authorization, quality constraints, and text source.

### 6. Generate real target-language TTS

- Standard mode: use `edge-tts` or another backend that supports `{tgt_lang}`.
- Voice cloning mode: use Qwen3-TTS or another authorized zero-shot backend.
- Silent, blank, corrupt, or placeholder files are not completion.
- Verify generated audio is playable, non-silent, reasonable duration, and usable by `ffmpeg`.

### 7. Align to source timeline

- For sentence-level sync, do not stretch one full narration to the total video duration.
- Use each SRT block start time as an anchor and place generated target-language segments on the original timeline.
- If a target segment is too long, first apply bounded speed-up, then trim only if necessary. Save and report speed-up/truncation data.

### 8. Generate subtitles

- Produce standalone subtitles before deciding whether to burn them into video.
- Bilingual subtitles must reuse source SRT timing and pair each source block with its matching target-language segment.
- Distinguish external subtitles, hard subtitles, and soft subtitles in the report.
- Name subtitle files with `{lang_pair}` (e.g. `subtitles.en-zh.ass`) so multiple language pairs can coexist without filename collision.

### 9. Compose final video

- Replace or overlay the source audio track with the dubbed audio.
- Use `ffprobe` to confirm final MP4 contains both video and audio streams.
- Export final named files into `output/`; do not leave only internal `artifacts/final*.mp4` files.

### 10. Package and report

- Copy the clean video, subtitle file, and burned subtitle version to `output/` using user-requested names.
- Use `{lang_pair}` in subtitle and bilingual video names; use `{tgt_lang}` in the clean dubbed video name.
- `report.md` must describe only commands and artifacts that actually ran in the current environment.
- Include limitations: transcription quality, TTS backend, voice-clone reference quality, speed-up/truncation, network/model constraints.

## Bundled Scripts

Copy scripts into the work directory's `scripts/` folder before running so the task archive is self-contained.

| Script | Purpose |
|---|---|
| `scripts/transcribe_with_faster_whisper.py` | Transcribe audio to text and SRT |
| `scripts/scaffold_rewrite_prompt.py` | Generate a human-review translation prompt scaffold |
| `scripts/generate_edge_tts.py` | Synthesize a continuous script to MP3/WAV via edge-tts |
| `scripts/generate_qwen3_voice_clone.py` | Synthesize with reference-audio voice cloning via Qwen3-TTS |
| `scripts/build_aligned_dub.py` | Per-segment synthesis aligned to source SRT start times |
| `scripts/build_bilingual_ass.py` | Generate bilingual ASS subtitles from source SRT and target segments |

Common commands:

```bash
python scripts/transcribe_with_faster_whisper.py \
  --audio source/audio.mp3 \
  --txt-out artifacts/transcript.{src_lang}.txt \
  --srt-out artifacts/transcript.{src_lang}.srt \
  --language {src_lang}

python scripts/scaffold_rewrite_prompt.py \
  --transcript artifacts/transcript.{src_lang}.txt \
  --output artifacts/script.{tgt_lang}.txt \
  --source-lang <SourceLanguageName> \
  --target-lang <TargetLanguageName> \
  --mode continuous

python scripts/generate_edge_tts.py \
  --script artifacts/script.{tgt_lang}.txt \
  --voice <voice-id-for-{tgt_lang}> \
  --mp3-out artifacts/narration.{tgt_lang}.mp3 \
  --wav-out artifacts/narration.{tgt_lang}.wav

python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.{src_lang}.srt \
  --tgt-segments artifacts/script.{tgt_lang}.segments.txt \
  --video source/video.mp4 \
  --backend edge-tts \
  --voice <voice-id-for-{tgt_lang}> \
  --wav-out artifacts/narration.{tgt_lang}.aligned.wav \
  --report-out artifacts/narration.{tgt_lang}.aligned.json \
  --segment-dir artifacts/aligned-segments

python scripts/build_bilingual_ass.py \
  --srt artifacts/transcript.{src_lang}.srt \
  --tgt-segments artifacts/script.{tgt_lang}.segments.txt \
  --lang-pair {src_lang}-{tgt_lang} \
  --ass-out artifacts/subtitles.{src_lang}-{tgt_lang}.ass
```

Voice-cloned aligned dubbing:

```bash
python scripts/generate_qwen3_voice_clone.py \
  --script artifacts/script.{tgt_lang}.txt \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --language <TargetLanguageName> \
  --mp3-out artifacts/narration.{tgt_lang}.clone.mp3 \
  --wav-out artifacts/narration.{tgt_lang}.clone.wav

python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.{src_lang}.srt \
  --tgt-segments artifacts/script.{tgt_lang}.segments.txt \
  --video source/video.mp4 \
  --backend qwen3-tts \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --language <TargetLanguageName> \
  --model-id ../models/Qwen3-TTS-12Hz-0.6B-Base \
  --segment-dir artifacts/aligned-segments-clone \
  --wav-out artifacts/narration.{tgt_lang}.clone.aligned.wav \
  --report-out artifacts/narration.{tgt_lang}.clone.aligned.json
```

Scripts accept compatibility aliases: `--translated-segments`, `--target-segments`, and `--tgt-segments`; `--language`, `--target-lang`, and `--target-language`; `--source-lang` and `--source-language`.

> **Note**: `--source-lang` / `--target-lang` / `--language` for `scaffold_rewrite_prompt.py` and `generate_qwen3_voice_clone.py` accept language **names** (e.g. `English`, `Chinese`, `French`). `--language` for `transcribe_with_faster_whisper.py` and `build_aligned_dub.py` accepts ISO 639-1 **codes** (e.g. `en`, `zh`, `fr`).

## Minimal Toolchain

- `ffmpeg` / `ffprobe`: media extraction, conversion, stream verification, final composition.
- Python virtual environment in the parent directory.
- `faster-whisper`: transcription and SRT generation.
- `edge-tts`: standard TTS for supported locales.
- `qwen-tts`, PyTorch, `soundfile`: reference-audio voice cloning.
- `numpy`: segment mixing and WAV timeline construction.
- `modelscope`: recommended fallback for Qwen model weights when Hugging Face downloads fail.

Tesla P4 note: P4 compute capability is 6.1 and may require Python 3.12 plus `torch==2.4.1+cu121`. Verify CUDA with a minimal tensor script before Qwen inference. P4 does not support flash-attn; expect slow per-segment synthesis. Segment caching allows resume after interruption.

## Troubleshooting

| Symptom | Response |
|---|---|
| `yt-dlp` returns `429`, bot verification, login, CAPTCHA | Inspect page state; if blocked, ask for local media/subtitles and record failure. |
| Hugging Face, YouTube, PyPI, or model downloads time out | Verify proxy and failing domains; use mirrors/cache where appropriate; do not rewrite workflow prematurely. |
| Python proxy parse error like `Invalid port: ':1]'` | Check proxy env vars; some libraries misparse `NO_PROXY=[::1]`; temporarily unset only for affected commands. |
| faster-whisper model unavailable | Check `~/.cache/huggingface`, `~/.cache/ctranslate2`, proxy, and local cache; request SRT/transcript if unavailable. |
| Qwen model download fails on Hugging Face | Prefer ModelScope and create a symlink if `.` becomes `___` in the local directory name. |
| Current Python too new for GPU PyTorch | Create a Python 3.12/3.11 virtualenv; do not pollute system Python. |
| `qwen-tts` upgrades PyTorch incompatibly | Install verified PyTorch first, then install `qwen-tts --no-deps` and add dependencies based on import errors. |
| `SoX could not be found` warning | Determine whether it blocks execution; if synthesis continues, log the limitation and proceed. |
| TTS backend does not support target language | Choose a backend/voice that supports `{tgt_lang}`; failed pronunciation or silence is not completion. |
| Voice clone does not match reference | Re-check single speaker, 3-15 s duration, no music/reverb/crosstalk, exact reference text. |
| Final duration correct but sentences drift | Rebuild per-segment audio anchored to SRT starts; total-duration alignment is insufficient. |
| Segment count mismatch | Ensure target segment lines exactly equal SRT block count; no merged, split, blank, or numbered lines. |
| User requested subtitles but only video exists | Produce requested external, hard, or soft subtitle deliverable and label it correctly. |
| User requested title-named output but only artifacts exist | Add packaging/export step into `output/` and verify final filenames exist. |

## Red Flags

Do not claim completion if any are true:

- Source/target language direction is unclear.
- Only a prompt scaffold exists; final target-language script has not been reviewed.
- TTS output is silent, blank, corrupt, or placeholder.
- Voice cloning is claimed without authorized reference audio and quality notes.
- Only total-duration alignment was done but sentence-level sync is claimed.
- Final MP4 was not verified with `ffprobe` for video and audio streams.
- `notes/commands.md` and `notes/issues.md` do not record real commands, failures, fixes, and constraints.
- Final user-requested files are missing from `output/`.

## Pre-Delivery Checklist

- [ ] `{src_lang}`, `{tgt_lang}`, `{lang_pair}`, output filenames, and subtitle requirements are clear.
- [ ] Work directory is next to the source video and contains `source/`, `artifacts/`, `notes/`, `output/`, `scripts/`, and `report.md`.
- [ ] Source video/audio, transcript, reviewed target script, TTS audio, subtitles, timing report, and final video have documented paths.
- [ ] Key commands are written to `notes/commands.md`.
- [ ] Blockers, fallbacks, quality limits, and authorization notes are written to `notes/issues.md` or `report.md`.
- [ ] If sentence-level sync was requested, timing report records speed-up/truncation.
- [ ] If voice cloning was used, reference audio source, quality, reference text, and **authorization** are recorded in `notes/issues.md` or `report.md`.
- [ ] Final MP4 passed `ffprobe` video/audio stream verification.
- [ ] All requested final filenames and subtitle formats exist in `output/`.
