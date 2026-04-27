---
name: video-localization
description: Use when needing to localize a video (YouTube or local) into any target language; generate dubbed audio, sentence-level alignment, bilingual subtitles, final video deliverables; or use external reference audio for voice cloning in the target language.
---

# Video Localization

## Overview

Convert a single YouTube or local video into a dubbed version in any target language, prioritizing **reproducible, auditable, real deliverables** over studio quality or batch automation.

Core principle: every step produces inspectable artifacts. All commands, issues, fixes, and intermediates go into a working directory next to the source video; final deliverables go into `output/`. Do not treat placeholder audio, total-duration alignment, or theoretical steps as completion.

Two dubbing modes:

- **Standard TTS dubbing**: synthesize using a standard TTS voice for the target language.
- **Reference-audio voice cloning**: use an external reference clip to clone the speaker's voice; defaults to extracting a clean speech segment from the source video, or uses a user-supplied reference if provided.

Out of scope: automatic voice fingerprint purification, legal/rights issues of speaker impersonation.

## When to Use

Apply when:

- User provides a YouTube URL or local video and requests a dubbed version in another language.
- User asks for dubbed audio, sentence-level sync, bilingual subtitles, or title-named final video.
- User values low cost, transparent process, and reproducible evidence over studio output.
- User requests voice cloning with either source-extracted or user-provided reference audio.
- Manual review of the translated script is required before TTS synthesis.

Do NOT apply when:

- Batch pipeline processing many videos.
- Goal is automated voice fingerprinting, liveness detection, lip sync, or 100% speaker replication.
- Source video, subtitles, or reference audio lack clear usage rights.
- User requires a black-box, non-auditable process.

## Working Directory Convention

Create an isolated working directory next to the source video. Use the source file's stem as the directory name:

```text
/path/to/Source-Video.mp4          # source video
/path/to/Source-Video/             # working directory
/path/to/Source-Video/output/      # final deliverables
```

Internal layout:

```text
/path/to/Source-Video/
  source/       # normalized inputs
  artifacts/    # transcripts, scripts, audio, subtitles, intermediate video
  notes/        # commands, issues, manual inspection logs
  output/       # user-facing final files
  scripts/      # copies of skill scripts
  report.md     # final summary
```

### Virtualenv and Model Reuse

`.venv` and `models/` live **in the parent directory** (e.g. `video-loc/.venv`, `video-loc/models/`) so multiple video tasks share them.

```text
/path/to/video-loc/
  .venv/           # shared Python env (PyTorch, faster-whisper, TTS libs, etc.)
  models/          # shared model weights
  Source-Video/    # working directory (source/, artifacts/, output/ only)
```

`video-loc/.gitignore` must ignore `.venv/` and `models/`.

### Minimum Required Artifacts

| Artifact | Description |
|---|---|
| `source/video.mp4` | normalized source video |
| `source/audio.mp3` | extracted or user-supplied audio |
| `artifacts/transcript.{src_lang}.txt` | source-language transcription |
| `artifacts/transcript.{src_lang}.srt` | timestamped source-language subtitles |
| `artifacts/script.{tgt_lang}.txt` | continuous target-language narration script |
| `artifacts/script.{tgt_lang}.segments.txt` | per-segment target-language script (sentence-level sync) |
| `source/reference.*` | reference audio + text for voice cloning mode |
| `artifacts/narration*.wav/mp3` | real synthesized TTS audio |
| `artifacts/final*.mp4` | candidate dubbed video or intermediate composite |
| `artifacts/subtitles*.ass` | internal subtitle artifacts |
| `output/{title}.mp4` | clean dubbed video for delivery |
| `output/{title}.ass` | external subtitle file |
| `output/{title}-bilingual.mp4` | bilingual subtitle burn-in video |
| `notes/commands.md` | actual commands executed |
| `notes/issues.md` | blockers, failures, fixes, limitations |
| `report.md` | user-facing final summary |

## Recommended Workflow

### 1. Acquire and Verify Source Media

- For YouTube URLs, try `yt-dlp` first; do not treat `429`, bot verification, or CAPTCHA as ordinary parameter errors.
- When page state, login, network requests, or CAPTCHA need inspection, use `chrome-devtools` per repository browser rules.
- If browser access is also blocked, stop workarounds and ask the user to provide a local file.
- Record acquisition method, failures, and final input source.

### 2. Normalize Media Inputs

- Run `ffprobe` to save input metadata: duration, video stream, audio stream.
- Canonicalize to stable paths: `source/video.mp4`, `source/audio.mp3`.
- All scripts reference canonical paths; avoid spaces, special characters, non-ASCII in paths.

### 3. Transcribe Source Audio

- Preferred tool: `faster-whisper` — outputs plain text and SRT.
- Specify `--language` if source language is known; otherwise let Whisper auto-detect.
- Sample-check quality after transcription: proper nouns, numbers, sentence breaks.

### 4. Translate to Target Language Script

- Translation should be natural, concise, and suitable for voiceover in the target language; preserve original meaning over literal word-for-word translation.
- Continuous dubbing: produce `script.{tgt_lang}.txt`.
- Sentence-level sync: produce `script.{tgt_lang}.segments.txt` — one line per SRT block, same count.
- **Manual review gate**: confirm the file contains the final translated script, not a prompt scaffold.

### 5. Prepare Reference Audio (voice cloning mode only)

Default: extract a 3-15 s clean single-speaker segment from the source video as `source/reference.wav`; write the corresponding source transcript text to `source/reference.txt`.
If the user supplies a reference audio file, use that instead and request its corresponding transcript text.

Reference audio quality targets:
- Single speaker, low background noise, 3-15 s of clear speech.
- No music, crosstalk, or strong reverb.
- Matching reference transcript increases clone stability.

Record: time range (or file path), reference text source, audio quality notes, and whether this is source-extracted or user-provided.

```bash
ffmpeg -y -ss 00:00:02.640 -to 00:00:15.000 \
  -i source/video.mp4 \
  -vn -ac 1 -ar 24000 -c:a pcm_s16le \
  source/reference.wav

printf '%s\n' 'Exact spoken text for this reference clip.' > source/reference.txt
```

### 6. Synthesize Target-Language TTS

- Standard dubbing: use a lightweight, reproducible TTS that supports the target language — e.g. `edge-tts` (100+ language locales via `--voice` flag).
- Voice cloning: use a zero-shot voice cloning backend — e.g. `Qwen3-TTS`.
- Silent files, blank files, or placeholder files are **not** completion.
- After generation: verify audio is non-silent, duration is reasonable, format is usable by `ffmpeg`.

**Choosing the right `edge-tts` voice**: run `edge-tts --list-voices` and filter by target language code (e.g. `fr-FR`, `de-DE`, `ja-JP`, `es-MX`).

### 7. Align to Source Timeline

- For sentence-level sync, do **not** stretch a single continuous narration to match total video duration.
- Use each SRT block's `start` timestamp as an anchor; place each TTS segment at the corresponding subtitle start.
- If the target-language segment is too long: first apply limited speed-up, then trim only if necessary. Record the timing report.
- "Total video duration unchanged" does not equal "audio is sentence-synchronized."

### 8. Generate Subtitle Artifacts

- Prefer exporting a standalone subtitle file before deciding whether to burn in.
- Bilingual subtitles: reuse source-language SRT timestamps and map per-segment translated lines into matching blocks.
- Distinguish external subtitles, hard-coded burn-in, and soft subtitles — do not substitute one for another.

### 9. Compose Final Video

- Replace or overlay source audio track with the dubbed audio.
- Verify with `ffprobe` that the final MP4 contains both video and audio streams.
- Retain clean version and subtitle version commands for re-runs.

### 10. Package and Report

- Do not stop at internal artifact names like `final-dubbed-aligned.mp4`.
- Copy or export all final files to `output/` with user-requested names.
- Report only what actually ran successfully in the current environment.

Typical delivery set:

```text
output/{title}.mp4              # clean dubbed video
output/{title}.ass              # external subtitle file
output/{title}-bilingual.mp4    # bilingual subtitle burn-in video
```

## Bundled Scripts

Copy skill scripts to the working directory's `scripts/` before running.

| Script | Purpose |
|---|---|
| `scripts/transcribe_with_faster_whisper.py` | Transcribe audio to text and SRT |
| `scripts/scaffold_rewrite_prompt.py` | Generate a human-review prompt scaffold for translation |
| `scripts/generate_edge_tts.py` | Synthesize continuous script to MP3/WAV via edge-tts |
| `scripts/generate_qwen3_voice_clone.py` | Synthesize with reference-audio voice cloning via Qwen3-TTS |
| `scripts/build_aligned_dub.py` | Per-segment synthesis + alignment to source SRT timeline |
| `scripts/build_bilingual_ass.py` | Generate bilingual ASS subtitle from source SRT + translated segments |

Common command patterns:

```bash
python scripts/transcribe_with_faster_whisper.py \
  --audio source/audio.mp3 \
  --txt-out artifacts/transcript.en.txt \
  --srt-out artifacts/transcript.en.srt
  # Add --language en to force; omit for auto-detect

python scripts/scaffold_rewrite_prompt.py \
  --transcript artifacts/transcript.en.txt \
  --output artifacts/script.fr.txt \
  --target-lang French \
  --mode continuous

python scripts/generate_edge_tts.py \
  --script artifacts/script.fr.txt \
  --voice fr-FR-DeniseNeural \
  --mp3-out artifacts/narration.fr.mp3 \
  --wav-out artifacts/narration.fr.wav

python scripts/generate_qwen3_voice_clone.py \
  --script artifacts/script.fr.txt \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --mp3-out artifacts/narration.fr.clone.mp3 \
  --wav-out artifacts/narration.fr.clone.wav

python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.en.srt \
  --tgt-segments artifacts/script.fr.segments.txt \
  --video source/video.mp4 \
  --backend edge-tts \
  --voice fr-FR-DeniseNeural \
  --wav-out artifacts/narration.fr.aligned.wav \
  --report-out artifacts/narration.fr.aligned.json \
  --segment-dir artifacts/aligned-segments

python scripts/build_bilingual_ass.py \
  --srt artifacts/transcript.en.srt \
  --tgt-segments artifacts/script.fr.segments.txt \
  --ass-out artifacts/subtitles.fr-en.ass
```

Voice cloning aligned dubbing:

```bash
python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.en.srt \
  --tgt-segments artifacts/script.fr.segments.txt \
  --video source/video.mp4 \
  --backend qwen3-tts \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --model-id ./models/Qwen3-TTS-12Hz-0.6B-Base \
  --segment-dir artifacts/aligned-segments-clone \
  --wav-out artifacts/narration.fr.clone.aligned.wav \
  --report-out artifacts/narration.fr.clone.aligned.json
```

**Tesla P4 inference note**: Tesla P4 (compute capability 6.1) does not support flash-attn; each TTS segment takes ~8-10 s. Scripts support per-segment caching and resume after interruption.

## Minimum Toolchain

- `ffmpeg` / `ffprobe`: audio extraction, stream verification, final video composition.
- Python virtualenv (in parent dir): PyTorch (P4-compatible build), faster-whisper, edge-tts, qwen-tts, numpy, soundfile.
- `faster-whisper`: transcription and SRT generation.
- `edge-tts`: standard TTS dubbing for 100+ language locales.
- `qwen-tts`: zero-shot voice cloning from reference audio.
- `modelscope`: download Qwen model weights (recommended over Hugging Face).

## Manual Review Gates

Stop and inspect real artifacts at these checkpoints:

- After source acquisition failure: confirm root cause (network, login, CAPTCHA, session block).
- After transcription: sample-check transcript text and SRT block count.
- After translation: confirm file contains final translated script, not a prompt scaffold.
- Before segment-level sync: confirm translated segment line count equals SRT block count.
- Before voice cloning: confirm reference audio exists, is playable, single-speaker, clean.
- After TTS synthesis: confirm output is real speech, not silence or a placeholder.
- After composition: `ffprobe` confirms video + audio streams present.
- Before delivery: confirm final named files actually exist in `output/` matching user requirements.

## Troubleshooting

### Python package index

- index: https://pypi.tuna.tsinghua.edu.cn/simple
- extra-index-url: https://mirrors.nju.edu.cn/pytorch/whl/cu121

### `yt-dlp` returns 429, bot verification, or CAPTCHA

Do not retry as ordinary parameter errors. Check page state in browser; if browser is also blocked, stop and ask user for local media. Record failures in `notes/issues.md`.

### Network timeouts (Hugging Face, PyPI, YouTube, etc.)

Verify proxy is active:

```bash
curl -I https://huggingface.co
curl -I https://pypi.org/simple/
```

Record failing domains and proxy variables. Do not misattribute network issues as script bugs.

### Python proxy parse error

Often caused by `NO_PROXY=localhost,127.0.0.1,[::1]` IPv6 notation that some libraries misparse. Temporarily unset `NO_PROXY` when running affected commands.

### faster-whisper model download failure

Check proxy and cache dirs (`~/.cache/huggingface`, `~/.cache/ctranslate2`). Retry after network recovery. If still failing, ask user to provide an SRT or transcript; do not fabricate SRT content.

### Hugging Face large model download failure

Use ModelScope instead:

```bash
uv pip install modelscope -i https://pypi.tuna.tsinghua.edu.cn/simple

python3 -c "
from modelscope import snapshot_download
model_dir = snapshot_download('Qwen/Qwen3-TTS-12Hz-0.6B-Base', cache_dir='.')
"
```

ModelScope replaces `.` with `___` in directory names. Create a symlink for compatibility:

```bash
cd video-loc/models/
ln -s Qwen3-TTS-12Hz-0___6B-Base Qwen3-TTS-12Hz-0.6B-Base
```

### `uv python install` fails (GitHub release download)

Use a mirror:

```bash
uv python install 3.12 \
  --mirror 'https://gh.llkk.cc/https://github.com/astral-sh/python-build-standalone/releases/download'
```

### Python version too new for GPU-compatible PyTorch

Create a Python 3.12 virtualenv in the parent directory. Do not downgrade the system Python or mix torch versions.

### Tesla P4: `no kernel image is available for execution on the device`

P4 is compute capability 6.1 (sm_61). Use Python 3.12 + `torch==2.4.1+cu121`. Validate before running Qwen inference:

```python
import torch
print(torch.__version__, torch.version.cuda)
print(torch.cuda.get_device_name(0), torch.cuda.get_device_capability(0))
x = torch.ones(4, device='cuda')
print((x * 2).sum().item())
```

### `qwen-tts` install upgrades PyTorch to incompatible version

Install and validate target PyTorch first, then install `qwen-tts` with `--no-deps`. Add dependencies one by one. Re-validate `torch.__version__` and CUDA tensor after each change.

### `transformers` / `huggingface-hub` / `accelerate` version conflicts

Pin versions to what the error requires. Do not blindly upgrade all packages. Run import checks after each pin change.

### `SoX could not be found` warning

Determine whether it blocks execution. If synthesis continues, log the limitation. Do not restructure the entire TTS pipeline over a warning.

### No user-supplied reference audio for voice cloning

Extract a 3-15 s clean single-speaker segment from the source video. Record: time range, reference text source, quality limits. If user later provides reference audio, switch to that file.

### Voice clone does not match reference

Check reference audio quality: single speaker, 3-15 s, no music or crosstalk, reference transcript is accurate. Re-extract a cleaner segment or ask user for a better sample.

### Target-language audio much shorter than video

For total-duration alignment: pad with silence. For sentence-level sync: use per-segment alignment with SRT anchors.

### Sentence drift despite correct total duration

Continuous narration cannot guarantee sentence sync. Rebuild using source SRT `start` timestamps as anchors. Save timing report; document speed-up or trim applied.

### User requests title-named output but only internal artifacts exist

Add an explicit packaging step: copy final files to `output/` with user-specified title names. Verify the files exist before reporting completion.

### User requests subtitles but only a video was delivered

Produce `.ass`, burn-in video, or whatever subtitle format was requested. Distinguish external subtitles, hard-coded, and soft.

### Path with spaces causes ffmpeg failure

Always quote paths. Normalize to canonical paths in `source/`; do not strip spaces — just quote them in every command.

### Translated segment count does not match SRT block count

Each translated segment line must correspond exactly to one SRT block. Adjust translation so line count equals SRT block count. Use `wc -l` and SRT block index for verification.

## Red Flags

Do not claim completion if any of these are true:

- Only a prompt scaffold exists; no final translated script.
- Voice cloning claimed but no reference audio or authorization.
- TTS output is a silent, blank, or placeholder file.
- Only total-duration alignment done but sentence-level sync claimed.
- Final video never verified with `ffprobe`.
- Report claims YouTube download succeeded but it was never actually completed.
- User requested title-named delivery but `output/` is empty or wrong.
- User requested subtitles but only a video was delivered.
- `notes/` has no record of commands, failures, or limitations.

## Pre-Delivery Checklist

- [ ] Working directory is next to source video; directory name equals source video stem.
- [ ] Internal structure is complete including `output/`.
- [ ] Source video, audio, transcript, translated script, TTS audio, subtitles, and final video all have documented paths.
- [ ] Key commands written to `notes/commands.md`.
- [ ] Blockers, fallbacks, quality limitations written to `notes/issues.md` or `report.md`.
- [ ] If sentence-level sync: timing report saved; speed-up and trim documented.
- [ ] If voice cloning: reference audio source, quality, and limitations documented.
- [ ] Final MP4 verified with `ffprobe` for video + audio streams.
- [ ] All user-requested final filenames, subtitle formats, and packaging exist in `output/`.
