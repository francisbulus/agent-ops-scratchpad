## Codex Session Handoff Rules

- Before starting substantial work, read `notes/codex-sessions/CURRENT.md`.
- Then read the file listed under `Active session file`.
- Continue from `Next Session Must Start With` unless the user asks to override.
- Before ending a meaningful work session, run:
  - `scripts/codex-session close --summary "..." --next "..."`
- When beginning a new session handoff file, run:
  - `scripts/codex-session start --title "..."`
