#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_SCRIPT="$ROOT_DIR/scripts/codex-session"

PASS_COUNT=0
FAIL_COUNT=0
TMP_BASE="$(mktemp -d)"

cleanup() {
  rm -rf "$TMP_BASE"
}
trap cleanup EXIT

log() {
  printf '%s\n' "$*"
}

fail() {
  log "FAIL: $*"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

pass() {
  log "PASS: $*"
  PASS_COUNT=$((PASS_COUNT + 1))
}

assert_file_exists() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    return 1
  fi
}

assert_contains() {
  local file="$1"
  local expected="$2"
  grep -Fq -- "$expected" "$file"
}

assert_regex() {
  local file="$1"
  local pattern="$2"
  grep -Eq -- "$pattern" "$file"
}

setup_repo() {
  local repo
  repo="$(mktemp -d "$TMP_BASE/repo.XXXXXX")"
  mkdir -p "$repo/scripts"
  cp "$SOURCE_SCRIPT" "$repo/scripts/codex-session"
  chmod +x "$repo/scripts/codex-session"

  (
    cd "$repo"
    git init -q
    git config user.name "Test User"
    git config user.email "test@example.com"
    printf "seed\n" > .seed
    git add .seed
    git commit -qm "seed"
  )

  printf '%s\n' "$repo"
}

run_test() {
  local name="$1"
  shift

  if "$@"; then
    pass "$name"
  else
    fail "$name"
  fi
}

test_start_creates_handoff_files() {
  local repo
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session start --title "First Session" >/dev/null

    assert_file_exists "notes/codex-sessions/CURRENT.md"
    assert_file_exists "notes/codex-sessions/INDEX.md"
    assert_regex "notes/codex-sessions/CURRENT.md" "Active session file: \`notes/codex-sessions/001-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{4}\\.md\`"
  )
}

test_start_state_and_index() {
  local repo
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session start --title "First Session" >/dev/null

    assert_contains "notes/codex-sessions/CURRENT.md" "- State: \`open\`"
    assert_contains "notes/codex-sessions/CURRENT.md" "- Session title: \`First Session\`"
    assert_regex "notes/codex-sessions/CURRENT.md" "Active session file: \`notes/codex-sessions/001-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{4}\\.md\`"
    assert_contains "notes/codex-sessions/INDEX.md" "| \`001\` |"
  )
}

test_close_requires_summary() {
  local repo output
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session start --title "Needs summary" >/dev/null
    set +e
    output="$(./scripts/codex-session close 2>&1)"
    local status=$?
    set -e
    [[ $status -ne 0 ]]
    printf '%s\n' "$output" | grep -Fq -- "--summary is required for close"
  )
}

test_close_writes_closeout_and_sets_closed_state() {
  local repo active_rel
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session start --title "Close Session" >/dev/null
    active_rel="$(sed -n 's/^- Active session file: `\(.*\)`/\1/p' notes/codex-sessions/CURRENT.md | head -n 1)"
    [[ -n "$active_rel" ]]
    ./scripts/codex-session close --summary "Done here" --next "Start next task" --blockers "none" >/dev/null

    assert_contains "notes/codex-sessions/CURRENT.md" "- State: \`closed\`"
    assert_contains "notes/codex-sessions/CURRENT.md" "- Summary for next session: \`Done here\`"
    assert_contains "notes/codex-sessions/CURRENT.md" "- Start next task"
    assert_contains "$active_rel" "## Closeout -"
    assert_contains "$active_rel" "- Summary: Done here"
    assert_contains "$active_rel" "- Next step for future session: Start next task"
    assert_contains "$active_rel" "- Blockers: none"
  )
}

test_close_without_active_session_fails() {
  local repo output
  repo="$(setup_repo)"

  (
    cd "$repo"
    set +e
    output="$(./scripts/codex-session close --summary "no active" 2>&1)"
    local status=$?
    set -e
    [[ $status -ne 0 ]]
    printf '%s\n' "$output" | grep -Fq -- "No active session file"
  )
}

test_session_number_increments() {
  local repo file_count
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session start --title "One" >/dev/null
    ./scripts/codex-session close --summary "one done" --next "two" >/dev/null
    ./scripts/codex-session start --title "Two" >/dev/null

    file_count="$(find notes/codex-sessions -maxdepth 1 -type f -name '[0-9][0-9][0-9]-*.md' | wc -l | tr -d ' ')"
    [[ "$file_count" == "2" ]]
    assert_contains "notes/codex-sessions/INDEX.md" "| \`001\` |"
    assert_contains "notes/codex-sessions/INDEX.md" "| \`002\` |"
    assert_regex "notes/codex-sessions/CURRENT.md" "Active session file: \`notes/codex-sessions/002-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{4}\\.md\`"
  )
}

test_status_bootstraps_layout() {
  local repo
  repo="$(setup_repo)"

  (
    cd "$repo"
    ./scripts/codex-session status >/dev/null
    assert_file_exists "notes/codex-sessions/CURRENT.md"
    assert_file_exists "notes/codex-sessions/INDEX.md"
    assert_contains "notes/codex-sessions/CURRENT.md" "- State: \`idle\`"
  )
}

main() {
  if [[ ! -x "$SOURCE_SCRIPT" ]]; then
    log "Missing executable script: $SOURCE_SCRIPT"
    exit 1
  fi

  run_test "start creates handoff files" test_start_creates_handoff_files
  run_test "start writes open state and index row" test_start_state_and_index
  run_test "close requires summary" test_close_requires_summary
  run_test "close writes closeout and closed state" test_close_writes_closeout_and_sets_closed_state
  run_test "close without active session fails" test_close_without_active_session_fails
  run_test "session numbering increments" test_session_number_increments
  run_test "status bootstraps layout" test_status_bootstraps_layout

  log
  log "Passed: $PASS_COUNT"
  log "Failed: $FAIL_COUNT"

  if [[ "$FAIL_COUNT" -ne 0 ]]; then
    exit 1
  fi
}

main "$@"
