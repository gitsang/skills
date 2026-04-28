# Skills

Personal agent skills that can be installed from this GitHub repository with the `skill` CLI.

## Available skills

- `skills/clip` - FFmpeg-based travel vlog clipping, stitching, audio mixing, and export workflow.
- `skills/procurement-analysis` - Product/vendor comparison and procurement decision analysis workflow.
- `skills/video-localization` - Single-video translation, dubbing, alignment, subtitles, and localized deliverables workflow.

## Install from GitHub

The current `npx skill` CLI accepts one package specifier and uses `SKILL_BASE_URL` to override the default skill repository.

Install a skill from this repository with:

```bash
SKILL_BASE_URL='https://github.com/gitsang/skills/tree/main' npx skill skills/clip
```

```bash
SKILL_BASE_URL='https://github.com/gitsang/skills/tree/main' npx skill skills/procurement-analysis
```

```bash
SKILL_BASE_URL='https://github.com/gitsang/skills/tree/main' npx skill skills/video-localization
```

Installed skills are copied into the current project's `.codebuddy/skills/` directory.

## Notes

`npx skill add <github-url>` is not supported by the current CLI. Use `SKILL_BASE_URL` plus the package path instead.
