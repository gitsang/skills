# Skills

Personal agent skills that can be installed from this GitHub repository with the `skills` CLI.

## Available skills

- `skills/clip` - FFmpeg-based travel vlog clipping, stitching, audio mixing, and export workflow.
- `skills/procurement-analysis` - Product/vendor comparison and procurement decision analysis workflow.
- `skills/video-localization` - Single-video translation, dubbing, alignment, subtitles, and localized deliverables workflow.

## Install from GitHub

Install a skill from this repository with:

```bash
npx skills add gitsang/skills --skill clip
```

```bash
npx skills add gitsang/skills --skill procurement-analysis
```

```bash
npx skills add gitsang/skills --skill video-localization
```

Install all skills from this repository with:

```bash
npx skills add gitsang/skills
```

Installed skills are copied into the agent skills directory, such as `.agents/skills/`.

## Notes

Use `npx skills add <owner>/<repo> --skill <name>` for agent skills.
