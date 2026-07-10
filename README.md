# Skills

Personal agent skills that can be installed from this GitHub repository with
the `skills` CLI.

## Available skills

- `skills/clip` - FFmpeg-based travel vlog clipping, stitching, audio mixing,
  and export workflow.
- `skills/first-principles-design` - Concept-first software design workflow for
  converging expanding requirements into stable module boundaries and design
  documents.
- `skills/pdf2md` - PDF to Markdown conversion workflow using llm subagent for
  formatting.
- `skills/procurement-analysis` - Product/vendor comparison and procurement
  decision analysis workflow.
- `skills/team-debate` - Team-mode debate workflow for one-issue-at-a-time
  design decisions, principle admission, explicit verdicts, and debate-record
  synchronization.
- `skills/video-localization` - Single-video translation, dubbing, alignment,
  subtitles, and localized deliverables workflow.

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
