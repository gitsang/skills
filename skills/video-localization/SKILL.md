---
name: video-localization
description: Use when needing to localize a single video (YouTube or local file) into another language; generate dubbed audio, sentence-level alignment, bilingual subtitles, final video deliverables; or use external reference audio for voice cloning in the target language.
---

# Video Localization

## Overview

Localize a single video (YouTube or local file) from a source language into a target language, prioritizing **reproducibility, auditability, and real deliverables** over studio-grade quality or batch automation.

Core principle: Every step must produce verifiable artifacts. All commands, issues, fixes, and intermediate outputs are preserved in a work directory adjacent to the source video; final deliverables go into `output/` within that directory. Do not treat placeholder audio, total-duration alignment, or theoretical steps as completion.

Supports two dubbing modes:

- **Standard TTS dubbing**: Use a standard TTS voice in the target language.
- **Reference-audio-driven voice cloning**: By default, extract clear speech from the source video as reference; if the user provides reference audio, use it instead. The dubbed audio attempts to match the reference voice in the target language.

Clear boundaries: This skill does not handle automatic voiceprint purification from complex source videos, nor high-risk "impersonation" scenarios involving legal or authorization concerns.

## When to Use

Applicable when:

- User provides a YouTube link or local video and requests dubbing into a target language.
- User requests translated dubbing, sentence-level sync, bilingual subtitles, or title-named final videos.
- User values low cost, transparent process, and reproducible evidence over a professional studio product.
- User requests voice cloning, wanting to use source video speech as default reference, or provides separate reference audio.
- Manual review of the translated script is needed before TTS to avoid mechanical translation.

Not applicable when:

- Batch pipeline processing of many videos.
- Goal is automatic voiceprint purification, voice authentication, lip-sync, or 100% original voice replication.
- Source video, subtitles, or reference audio lack clear processing permissions.
- User requests an unauditable, black-box workflow without saving intermediate evidence.

## Work Directory and Deliverable Conventions

Each task must first create an independent work directory adjacent to the source video. The directory name uses the stem (filename without extension) of the source video:

```text
/path/to/Source-Video-Name.mp4  # Source video
/path/to/Source-Video-Name/     # Work directory
/path/to/Source-Video-Name/output/  # Final deliverables
```

Work directory internal structure:

```text
/path/to/Source-Video-Name/
  source/       # Original and standardized inputs
  artifacts/    # Transcription, script, audio, subtitle, video intermediates
  notes/        # Commands, issues, constraints, manual check records
  output/       # User-facing final deliverables
  scripts/      # Reusable scripts copied from the skill
  report.md     # Final report
```

### Virtual Environment and Model Reuse Convention

`.venv` (Python virtual environment) and `models/` (model weights directory) **should not be inside the work directory**, but in the parent directory of the source video (e.g., `yt-video/.venv` and `yt-video/models/`), so multiple video tasks can share them.

```text
/path/to/yt-video/
  .venv/            # Shared Python virtual environment (PyTorch, faster-whisper, qwen-tts, etc.)
  models/           # Shared model weights directory
    Qwen3-TTS-12Hz-0___6B-Base/   # ModelScope downloaded actual directory name
    Qwen3-TTS-12Hz-0.6B-Base -> Qwen3-TTS-12Hz-0___6B-Base  # symlink (compatible with original repo ID format)
  Source-Video-Name/   # Work directory (only source/, artifacts/, output/, etc.)
```

The work directory only keeps video-specific intermediate artifacts, without reinstalling environments or re-downloading models.

`yt-video/` contains a `.gitignore` that auto-ignores `.venv/` and `models/`.

Language code convention: Use ISO 639-1 two-letter codes (e.g., `en`, `zh`, `ja`, `ko`, `es`, `fr`, `de`).

Minimum recommended artifacts:

| Artifact | Description |
|---|---|
| `source/video.mp4` | Standardized video input |
| `source/audio.mp3` | Audio extracted from video or user-provided |
| `artifacts/transcript.{src}.txt` | Source language transcription text |
| `artifacts/transcript.{src}.srt` | Timed source language subtitles |
| `artifacts/script.{tgt}.txt` | Continuous target-language narration script |
| `artifacts/script.{tgt}.segments.txt` | Per-segment target-language script for sentence-level sync |
| `source/reference.*` | Reference audio and text for voice cloning; default from source video clip, or user-provided |
| `artifacts/narration*.wav/mp3` | Real generated target-language TTS audio |
| `artifacts/final*.mp4` | Candidate dubbed video or internal synthesis output |
| `artifacts/subtitles*.ass` | Internal subtitle output |
| `output/{title}.mp4` | Clean dubbed video for user delivery |
| `output/{title}.ass` | External subtitle file for user delivery |
| `output/{title}-bilingual.mp4` | Bilingual subtitle burned video for user delivery |
| `notes/commands.md` | Actually executed commands |
| `notes/issues.md` | Blockers, failures, fixes, and constraints |
| `report.md` | Final user-facing summary |

## Recommended Workflow

### 1. Acquire and Verify Source Material

- If a YouTube link, try `yt-dlp` first, but do not misjudge `429`, bot verification, or login as ordinary parameter issues.
- When page review, login state, network requests, or CAPTCHA are needed, follow repository browser rules and prefer `chrome-devtools`.
- If the browser is also blocked, stop trying to bypass and ask the user to provide a local source file.
- Record the source acquisition method, failure info, and final input source used.

### 2. Standardize Media Input

- Use `ffprobe` to save input metadata, confirm duration, video stream, audio stream.
- Convert any user filename into stable paths within the work directory, e.g., `source/video.mp4` and `source/audio.mp3`.
- Subsequent scripts only reference standardized paths, avoiding spaces, non-ASCII, or special characters causing script failures.

### 3. Source Language Transcription

- Prefer `faster-whisper` for plain text and SRT output.
- After transcription, sample-check quality, especially proper nouns, numbers, and sentence breaks.
- If model download fails before starting, check proxy and cache paths first, then consider alternative tools.

### 4. Generate Target-Language Script

- The translated script should be natural, concise, and suitable for narration; preserve original meaning over literal word-by-word translation.
- Continuous narration: prepare `script.{tgt}.txt`.
- Sentence-level sync: prepare `script.{tgt}.segments.txt`, one line per SRT block.
- This is the manual review checkpoint: confirm the file contains the final translated script, not a prompt scaffold.

### 5. Prepare Reference Audio (Voice Cloning Mode Only)

By default, extract a clear single-speaker segment from the source video as `source/reference.wav`, and write the corresponding source text to `source/reference.txt`. If the user provides reference audio, use the user's file instead, and try to get the corresponding reference text from the user.

Reference audio should ideally satisfy:

- Single speaker.
- Low background noise.
- 3-15 seconds of clear speech.
- No music, multi-speaker overlap, or strong reverb.
- Provide corresponding reference text to improve cloning backend stability; when extracting from source video, use the transcribed text for the corresponding time range.

When extracting a reference segment from the source video, must record: time range, reference text source, audio quality constraints, and that this is not an external reference sample. When user provides reference audio, must record user file path, playability, duration, speaker purity, and reference text source.

Source video reference segment example:

```bash
ffmpeg -y -ss 00:00:02.640 -to 00:00:15.000 \
  -i source/video.mp4 \
  -vn -ac 1 -ar 24000 -c:a pcm_s16le \
  source/reference.wav

printf '%s\n' 'Exact spoken text for this reference clip.' > source/reference.txt
```

### 6. Generate Real Target-Language TTS

- Standard dubbing: prefer lightweight, reproducible, no-API-Key TTS in the target language, e.g., `edge-tts`.
- Voice cloning: use a zero-shot voice cloning backend, e.g., `Qwen3-TTS`.
- Silence files, blank files, or placeholder files do not count as completion.
- After generation, verify with audio tools: non-silent, reasonable duration, sample format usable by `ffmpeg`.

### 7. Align to Original Timeline

- If sentence-level sync matters, do not synthesize a full narration and then stretch to total duration.
- Use original SRT `start` times as anchors, placing each target-language TTS segment at its corresponding subtitle block start time.
- When target text is too long, first apply limited speedup, then truncate if necessary, and record the timing report.
- "Video total duration unchanged" does not mean "audio-video sync succeeded." Sync means sentences return to their original time anchors.

### 8. Generate Subtitle Artifacts

- When user requests subtitles, prefer standalone subtitle files before deciding whether to burn.
- Bilingual subtitles should reuse source SRT timing and merge per-segment target script by block.
- Distinguish external subtitles, hard subtitles, and soft subtitles; do not pass one off as another.

### 9. Compose Final Video

- Replace or overlay the source video audio track with the dubbed audio track.
- Use `ffprobe` to confirm the final MP4 contains both video and audio streams.
- Keep clean and subtitle version generation commands for future re-runs.

### 10. Delivery Packaging and Report

- Do not stop at internal filenames like `final-voiceover-aligned.mp4`.
- Final deliverables must be copied or exported to `output/`, not left only in `artifacts/`.
- If user requests title-based naming, export the requested filenames in `output/`.
- Report should only describe what actually ran in the current environment, not imagined ideal paths.

Common delivery set:

```text
output/{title}.mp4              # Clean dubbed video
output/{title}.ass              # External subtitle file
output/{title}-bilingual.mp4    # Burned bilingual subtitle video
```

## Built-in Scripts

Always copy skill scripts to the work directory's `scripts/` before running, to avoid polluting the skill directory and to enable complete archival.

| Script | Purpose |
|---|---|
| `scripts/transcribe_with_faster_whisper.py` | Transcribe audio to source-language text and SRT |
| `scripts/scaffold_rewrite_prompt.py` | Generate a reviewable prompt scaffold for target-language translation |
| `scripts/generate_edge_tts.py` | Generate full target-language script as MP3/WAV via edge-tts |
| `scripts/generate_qwen3_voice_clone.py` | Generate full target-language voice-cloned MP3/WAV using reference audio |
| `scripts/build_aligned_dub.py` | Per-segment target-language TTS aligned to original subtitle start times |
| `scripts/build_bilingual_ass.py` | Generate bilingual ASS subtitles using source timeline |

Common command templates (replace `{src}` and `{tgt}` with language codes):

```bash
python scripts/transcribe_with_faster_whisper.py \
  --audio source/audio.mp3 \
  --txt-out artifacts/transcript.{src}.txt \
  --srt-out artifacts/transcript.{src}.srt

python scripts/scaffold_rewrite_prompt.py \
  --transcript artifacts/transcript.{src}.txt \
  --output artifacts/script.{tgt}.txt \
  --source-lang {src} \
  --target-lang {tgt} \
  --mode continuous

python scripts/generate_edge_tts.py \
  --script artifacts/script.{tgt}.txt \
  --mp3-out artifacts/narration.{tgt}.mp3 \
  --wav-out artifacts/narration.{tgt}.wav \
  --voice <edge-tts-voice-id>

python scripts/generate_qwen3_voice_clone.py \
  --script artifacts/script.{tgt}.txt \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --mp3-out artifacts/narration.{tgt}.clone.mp3 \
  --wav-out artifacts/narration.{tgt}.clone.wav

python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.{src}.srt \
  --translated-segments artifacts/script.{tgt}.segments.txt \
  --video source/video.mp4 \
  --backend edge-tts \
  --voice <edge-tts-voice-id> \
  --wav-out artifacts/narration.{tgt}.aligned.wav \
  --report-out artifacts/narration.{tgt}.aligned.json \
  --segment-dir artifacts/aligned-segments

python scripts/build_bilingual_ass.py \
  --srt artifacts/transcript.{src}.srt \
  --translated-segments artifacts/script.{tgt}.segments.txt \
  --source-lang {src} \
  --target-lang {tgt} \
  --ass-out artifacts/subtitles.{tgt}-{src}.ass
```

Voice cloning per-segment alignment with `build_aligned_dub.py` requires reference audio. `--model-id` points to the shared model directory:

```bash
python scripts/build_aligned_dub.py \
  --srt artifacts/transcript.{src}.srt \
  --translated-segments artifacts/script.{tgt}.segments.txt \
  --video source/video.mp4 \
  --backend qwen3-tts \
  --reference-audio source/reference.wav \
  --reference-text source/reference.txt \
  --model-id ./models/Qwen3-TTS-12Hz-0.6B-Base \
  --segment-dir artifacts/aligned-segments-clone \
  --wav-out artifacts/narration.{tgt}.clone.aligned.wav \
  --report-out artifacts/narration.{tgt}.clone.aligned.json
```

**Tesla P4 inference performance**: Tesla P4 (compute capability 6.1) does not support flash-attn, each TTS segment takes ~8-10 seconds. 187 segments (~8 min video) totals ~25-30 minutes. Scripts support segment caching for resume after interruption.

## Minimal Toolchain

- `ffmpeg` and `ffprobe`: Extract audio, inspect media streams, compose final video.
- Python virtual environment: Shared in parent directory (e.g., `yt-video/.venv`), containing PyTorch (P4-compatible), faster-whisper, qwen-tts, edge-tts.
- `faster-whisper`: Source language transcription and SRT generation.
- `edge-tts`: Standard target-language dubbing (multi-language support).
- `qwen-tts`: External reference audio-driven voice cloning.
- `modelscope`: Download Qwen model weights from ModelScope (recommended).
- `numpy`, `soundfile`, etc.: Per-segment alignment and WAV concatenation.

## Manual Review Checkpoints

Must stop and verify real artifacts at these nodes:

- After source acquisition failure, confirm whether it's network, login, CAPTCHA, or session-level block.
- After transcription, sample-check source text and SRT block count.
- After translation, confirm it's the final translated script, not a prompt draft.
- For per-segment sync, confirm translated segment line count matches SRT block count.
- For voice cloning, confirm reference audio exists, is playable, single-speaker, clear.
- After TTS, confirm output is real speech, not silence or placeholder.
- After composition, use `ffprobe` to confirm video and audio streams both exist.
- Before delivery, confirm final named files truly exist in `output/` and match user requirements.

## Troubleshooting

### python index

- index: https://pypi.tuna.tsinghua.edu.cn/simple
- extra-index-url: https://mirrors.nju.edu.cn/pytorch/whl/cu121

### `yt-dlp` encounters `429`, bot verification, or CAPTCHA

**Symptoms**: Downloading YouTube source video or subtitles produces `429`, bot verification, CAPTCHA, login check, or long timeout.

**How to resolve**: Do not treat these as ordinary parameter errors and retry endlessly. First review page state via browser; if browser is also blocked by CAPTCHA or login, stop bypassing and ask the user to provide local media or subtitle files, and write failure info to `notes/issues.md`.

### Hugging Face, YouTube, ytscribe, PyPI, etc. network timeout

**Symptoms**: `faster-whisper`, `yt-dlp`, `webfetch`, `pip`, `uv`, `curl` external service access times out, TLS EOF, connection reset, causing transcription, subtitle, or dependency download failures.

**How to resolve**: First verify proxy is actually effective:

```bash
curl -I https://huggingface.co
curl -I https://www.youtube.com
curl -I https://pypi.org/simple/
```

Record failed domains, errors, and proxy variables. Do not misjudge network issues as script issues. After proxy recovers, re-run original commands; if only a specific domain is unreachable, prefer mirror sources or local cache over changing the entire workflow.

### Python/httpx proxy format error, e.g., `Invalid port: ':1]'`

**Symptoms**: CLI `curl` can access the network, but Python packages (e.g., `httpx`, `huggingface_hub`) report proxy URL parse errors, commonly `Invalid port: ':1]'`.

**How to resolve**: Check proxy environment variables:

```bash
env | grep -i proxy
```

Common cause: IPv6 syntax in `NO_PROXY=localhost,127.0.0.1,[::1]` is parsed incorrectly by some libraries. Temporarily remove when running critical commands:

```bash
unset NO_PROXY && python your_script.py
```

### faster-whisper model download failure or no local cache

**Symptoms**: Source transcription fails, unable to download Whisper/faster-whisper model, or no local model snapshot.

**How to resolve**: First check proxy and cache directories: `~/.cache/huggingface`, `~/.cache/ctranslate2`, `~/.cache/whisper`. After network recovers, re-run transcription command. If still failing, ask user to provide source SRT/transcription for this video, or switch to an already-installed, offline-capable ASR tool; do not hand-write fake SRT.

### Hugging Face large model weights cannot be downloaded

**Symptoms**: Downloading Qwen3-TTS or other large model weights from HuggingFace produces SSL EOF, connection reset, timeout, or redirect to `cas-bridge.xethub.hf.co` followed by SSL EOF.

**How to resolve**: Prefer ModelScope for downloading model snapshots locally:

```bash
uv pip install modelscope -i https://pypi.tuna.tsinghua.edu.cn/simple

python3 -c "
from modelscope import snapshot_download
model_dir = snapshot_download('Qwen/Qwen3-TTS-12Hz-0.6B-Base', cache_dir='.')
"
```

Place model directory in the parent directory of the video (e.g., `yt-video/models/`) for reuse.

**ModelScope download path naming note**: ModelScope replaces `.` in filenames with `___`. E.g., `Qwen3-TTS-12Hz-0.6B-Base` downloads as `Qwen3-TTS-12Hz-0___6B-Base`. Create a symlink for original repo ID compatibility:

```bash
cd yt-video/models/
ln -s Qwen3-TTS-12Hz-0___6B-Base Qwen3-TTS-12Hz-0.6B-Base
```

Then point `--model-id` to the local directory (e.g., `./models/Qwen3-TTS-12Hz-0.6B-Base`).

Report must record actual model source and local path.

### `uv python install 3.12` GitHub release download failure

**Symptoms**: `uv python install 3.12` downloading `python-build-standalone` encounters GitHub release TLS EOF, connection reset, or download timeout.

**How to resolve**: Use uv's `--mirror`, and first test with `curl -I -L` that the mirror can return Python tarballs:

```bash
uv python install 3.12 \
  --mirror 'https://gh.llkk.cc/https://github.com/astral-sh/python-build-standalone/releases/download'
```

After successful install, use an isolated virtual environment for TTS/GPU dependencies, do not pollute the main environment.

### Current Python version too new, old GPU-compatible PyTorch has no suitable wheel

**Symptoms**: System only has Python 3.13/3.14, but needs to install older CUDA PyTorch; pip cannot find a matching wheel, or resolves to an incompatible version.

**How to resolve**: Create a Python 3.12/3.11 environment, e.g., `.venv` (place at `yt-video/` level). Do not forcibly downgrade Python or mix multiple torch installs in the existing environment. Pin Python first, then install target torch, then add Qwen dependencies.

### Tesla P4 reports `no kernel image is available for execution on the device`

**Symptoms**: Qwen model starts loading on GPU then fails, PyTorch reports `no kernel image is available for execution on the device`; logs show Tesla P4 compute capability `6.1`, but current torch only supports `sm_75+`.

**How to resolve**: Root cause is PyTorch wheel does not support P4's `sm_61`. Switch to an older CUDA PyTorch that supports this architecture, e.g., Python 3.12 + `torch==2.4.1+cu121`. After install, verify with a minimal CUDA tensor script:

```python
import torch
print(torch.__version__, torch.version.cuda)
print(torch.cuda.get_device_name(0), torch.cuda.get_device_capability(0))
x = torch.ones(4, device='cuda')
print((x * 2).sum().item())
```

Only continue Qwen inference after this verification passes.

### Installing `qwen-tts` auto-upgrades to incompatible PyTorch

**Symptoms**: P4-compatible torch was installed, but installing `qwen-tts` causes pip resolver to upgrade torch, re-introducing GPU architecture incompatibility.

**How to resolve**: First install and verify target PyTorch, then install `qwen-tts` with `--no-deps`, followed by adding dependencies one by one based on import errors. This prevents pip from re-resolving and overwriting torch. After adding deps, re-verify `torch.__version__` and CUDA tensor.

### `transformers`, `huggingface-hub`, `accelerate` version conflicts

**Symptoms**: Importing Qwen or Transformers produces version constraint errors, e.g., `transformers 4.57.x` requires `huggingface-hub < 1.0`, or `qwen-tts` requires a specific `accelerate` version.

**How to resolve**: Pin versions based on actual errors, do not blindly upgrade all packages. E.g., pin `huggingface-hub` to `<1.0`, pin `accelerate` to the version required by `qwen-tts`. After each adjustment, run import checks and confirm torch was not replaced.

### System lacks `sox` command, Qwen import warns `SoX could not be found`

**Symptoms**: Qwen or audio library import prints `SoX could not be found`, but the script may continue running.

**How to resolve**: First determine if this is just a warning. If generation can continue, record the constraint; if it actually blocks, install system SoX, or confirm whether the script only needs the Python `sox` package. Do not refactor the entire TTS flow because of one warning.

### Voice cloning without user-provided reference audio

**Symptoms**: User requests voice cloning but did not upload independent reference audio or reference text.

**How to resolve**: By default, select 3-15 seconds of clear single-speaker speech from the source video as `source/reference.wav`, and write corresponding transcribed text to `source/reference.txt`. If the user later provides reference audio, switch to the user's file. Regardless of source, record reference audio source, time range or file path, text source, and quality constraints.

### Voice cloning does not sound like reference

**Symptoms**: Generated voice is noticeably different from source video or user reference audio, or stability is poor.

**How to resolve**: First check reference audio quality: single speaker? 3-15 seconds? Any music/reverb/multi-speaker overlap? Does reference text precisely match? If the default source video reference segment is not ideal, re-select a cleaner speech segment; if user reference audio is not ideal, ask for a cleaner sample.

### Dubbed audio is much shorter than video

**Symptoms**: Full target-language audio is significantly shorter than the video, with lots of blank space at the end.

**How to resolve**: If only total duration matters, pad with silence; if sentence sync matters, per-segment alignment is required, not just stretching the full audio to video length.

### Video length is correct but sentences drift over time

**Symptoms**: Final video total duration is correct, but target-language sentences increasingly drift from the visuals or subtitles.

**How to resolve**: Continuous narration cannot guarantee sentence sync. Switch to per-segment dubbing anchored on original SRT start times, save the timing report; record speedup or truncation when target text is too long.

### User requests title-based naming but only internal artifacts exist

**Symptoms**: Work directory only has `artifacts/final-*.mp4`, no user-ready final named files.

**How to resolve**: Add an explicit packaging/export step, copy or export final files to `output/` using user-requested title names, and verify files truly exist.

### User requests subtitles but only video files exist

**Symptoms**: Deliverables only include MP4, no external subtitles, hard subtitles, or user-requested subtitle format.

**How to resolve**: Produce `.ass`, hard-subtitle video, or user-specified subtitle format. Report must distinguish external, hard, and soft subtitles; do not pass one off as another.

### Filename spaces cause ffmpeg failure

**Symptoms**: ffmpeg or Python scripts fail processing video filenames with spaces, e.g., `Invalid argument` or file not found.

**How to resolve**: Always quote paths with spaces. In ffmpeg commands, use `"output/filename with spaces.mp4"`. When standardizing paths, do not remove spaces (they are part of the filename), just quote correctly in command lines.

### Translated segment line count does not match SRT block count

**Symptoms**: `build_aligned_dub.py` reports `Segment count mismatch` error.

**How to resolve**: Each line of the translated segment file corresponds to one subtitle block in the SRT. When generating the translation, ensure the translated line count strictly matches the SRT block count. If source segments were merged or split during translation, adjust the translated file to maintain a one-to-one correspondence. Use `wc -l` and SRT block numbers for final verification.

## Red Flags

If any of these are true, do not claim completion:

- Only prompt scaffold exists, no final translated script.
- Claims voice cloning was done, but no external reference audio or authorized source exists.
- TTS output is silence, blank, or placeholder files.
- Only did total-duration alignment but calls it sentence-level sync success.
- Final video was never verified with `ffprobe`.
- Report claims YouTube download succeeded but it actually did not run through.
- User requested final named delivery, but `output/` has no final files.
- User requested subtitle delivery, but only video was delivered.
- Work directory `notes/` does not record commands, failures, and constraints.

## Pre-completion Checklist

- [ ] Work directory is adjacent to source video, directory name equals source video stem.
- [ ] Work directory internal structure is complete, including `output/`.
- [ ] Source video, audio, transcription, translated script, TTS, subtitles, and final video all have clear paths.
- [ ] Key commands written to `notes/commands.md`.
- [ ] Blockers, fallbacks, quality constraints written to `notes/issues.md` or `report.md`.
- [ ] If per-segment sync, timing report saved with speedup/truncation details.
- [ ] If voice cloning, reference audio source, quality, and constraints recorded.
- [ ] Final MP4 verified via `ffprobe` for video and audio streams.
- [ ] User-requested final filenames, subtitle format, and packaging all exist in `output/`.
