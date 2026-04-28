# Skills

Personal agent skills that can be installed from this GitHub repository with the `skills` CLI.

## Available skills

- `skills/clip` - FFmpeg-based travel vlog clipping, stitching, audio mixing, and export workflow.
- `skills/procurement-analysis` - Product/vendor comparison and procurement decision analysis workflow.
- `skills/video-localization` - Single-video translation, dubbing, alignment, subtitles, and localized deliverables workflow.

## Install from GitHub

Install a skill from this repository with:

```bash
npx skills add https://github.com/gitsang/skills/tree/main/skills/clip
```

```bash
npx skills add https://github.com/gitsang/skills/tree/main/skills/procurement-analysis
```

```bash
npx skills add https://github.com/gitsang/skills/tree/main/skills/video-localization
```

Installed skills are copied into the agent skills directory, such as `.agents/skills/`.

## Notes

Use `npx skills add <github-url>` for agent skills. The `npx skill` command is CodeBuddy-specific.
