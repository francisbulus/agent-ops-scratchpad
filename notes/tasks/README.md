# MVP Tickets Tracker

This folder tracks implementation work as testable tickets.

## Structure

- `BOARD.md`: canonical status board.
- `TEMPLATE.md`: ticket template.
- `tickets/`: one markdown file per ticket.

## Status Model

- `backlog`: defined, not ready.
- `ready`: clear and unblocked.
- `in_progress`: active work.
- `blocked`: waiting on dependency/decision.
- `done`: merged with acceptance criteria met.

## Rules

- Keep each ticket small enough to complete and verify in 1 PR.
- Every ticket must include a test plan and acceptance criteria.
- Update both the ticket `Status` and `BOARD.md` when status changes.
- Link PR/commit in the ticket before moving to `done`.
