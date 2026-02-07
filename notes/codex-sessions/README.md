# Codex Session Handoffs

This directory is for Codex-to-Codex continuity between separate sessions.

## Why

- Preserve operational context across new Codex sessions.
- Avoid re-discovery of decisions, blockers, and next actions.
- Keep one canonical "start here" pointer for future sessions.

## Files

- `CURRENT.md`: single source of truth for what to read first.
- `INDEX.md`: append-only ledger of created session files.
- `NNN-YYYY-MM-DD-HHMM.md`: one session handoff file per Codex work session.

## Workflow

1. Start a session:
   - `scripts/codex-session start --title "..."`.
2. Work normally and update the active session file sections.
3. Close the session:
   - `scripts/codex-session close --summary "..." --next "..."`

## Conventions

- Keep `Next Session Must Start With` explicit and actionable.
- Record irreversible decisions under `Decisions`.
- Treat `CURRENT.md` as mandatory first-read for new sessions.
