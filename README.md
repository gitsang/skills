# Skills

Personal agent skills.

## Available skills

- `skills/team-debate` - Team-mode debate workflow for one-issue-at-a-time
  design decisions, principle admission, explicit verdicts, and debate-record
  synchronization.
- `skills/pdf2md` - PDF to Markdown conversion workflow using llm subagent for
  formatting.
- `skills/first-principles-design` - Concept-first software design workflow for
  converging expanding requirements into stable module boundaries and design
  documents.
- `skills/procurement-analysis` - Product/vendor comparison and procurement
  decision analysis workflow.

### SKILL in development

- `skills/travel-planner` - flyai based travel planner and plan-site builder.
- `skills/clip` - FFmpeg-based travel vlog clipping, stitching, audio mixing,
  and export workflow.
- `skills/video-localization` - Single-video translation, dubbing, alignment,
  subtitles, and localized deliverables workflow.
- `skills/video-localization-mimo` - Single-video translation, dubbing,
  alignment, subtitles, and localized deliverables workflow using mimo.

## Install from GitHub

Install a skill from this repository with:

```bash
npx skills add gitsang/skills --skill {SKILL_NAME}
```

Install all skills from this repository with:

```bash
npx skills add gitsang/skills
```

Installed skills are copied into the agent skills directory, such as `.agents/skills/`.
