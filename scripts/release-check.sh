#!/usr/bin/env bash
# Local release pre-flight — the same gates the Release workflow's
# pre-release-checks job enforces, runnable before spending a dispatch.
# Invoked by `make release`; also useful standalone.
#
# Usage: scripts/release-check.sh [any args ignored]
# The same gates run for real releases and dry-run rehearsals alike.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
FAIL=0

pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; FAIL=1; }
warn() { echo "  ⚠️  $1"; }

BRANCH=$(git branch --show-current)
echo "Release pre-flight on branch '${BRANCH}' @ $(git rev-parse --short HEAD)"

# 1. Clean tree — the release commits version bumps and must start from a
#    known state.
if [ -n "$(git status --porcelain)" ]; then
  fail "working tree is dirty — commit or stash first"
  git status --short | head -5 | sed 's/^/       /'
else
  pass "working tree clean"
fi

# 2. On main — the release workflow dispatches from the current branch and
#    pushes version bumps back to it.
if [ "$BRANCH" = "main" ]; then
  pass "on main"
else
  warn "not on main — releases are normally cut from main"
fi

# 3. Local tree in sync with origin, so CI validates what you release.
git fetch --quiet origin main 2>/dev/null || warn "could not fetch origin (offline?)"
if git rev-parse --verify origin/main >/dev/null 2>&1; then
  LOCAL=$(git rev-parse HEAD)
  REMOTE=$(git rev-parse origin/main)
  if [ "$LOCAL" = "$REMOTE" ]; then
    pass "HEAD == origin/main"
  else
    fail "HEAD ($LOCAL) != origin/main ($REMOTE) — push or pull first"
  fi
fi

# 4. Version files consistent — same invariant CI's version-drift job checks.
#    Anchored on the NEWEST of VERSION and the latest release tag: in a
#    release window the tree is legitimately one step ahead of the tag.
git fetch --quiet --tags origin 2>/dev/null || warn "could not fetch tags (offline?)"
LATEST_TAG_VERSION=$(git tag -l 'v*' | sort -V | tail -n1 | sed 's/^v//')
LATEST_TAG_VERSION=${LATEST_TAG_VERSION:-$(cat VERSION)}
TARGET=$(printf '%s\n%s\n' "$(cat VERSION)" "$LATEST_TAG_VERSION" | sort -V | tail -n1)
if bash scripts/set-version.sh "$TARGET" --check >/dev/null 2>&1; then
  pass "version files consistent with v${TARGET}"
else
  fail "version-file drift detected (CI version-drift would fail):"
  bash scripts/set-version.sh "$TARGET" --check || true
fi

# 5. CI green on this exact commit.
if command -v gh > /dev/null 2>&1; then
  LATEST_RUN=$(gh run list --workflow=ci.yml --branch=main --limit=1 \
    --json conclusion,status,headSha --jq '.[0]' 2>/dev/null || true)
  if [ -n "$LATEST_RUN" ] && [ "$LATEST_RUN" != "null" ]; then
    RUN_SHA=$(echo "$LATEST_RUN" | jq -r '.headSha')
    RUN_STATUS=$(echo "$LATEST_RUN" | jq -r '.status')
    RUN_CONCLUSION=$(echo "$LATEST_RUN" | jq -r '.conclusion')
    HEAD_SHA=$(git rev-parse HEAD)
    if [ "$RUN_STATUS" != "completed" ]; then
      fail "CI is still running — wait for it to finish"
    elif [ "$RUN_CONCLUSION" != "success" ]; then
      fail "latest CI run failed (${RUN_CONCLUSION})"
    elif [ "$RUN_SHA" != "$HEAD_SHA" ]; then
      warn "latest CI run is for a different commit ($(echo "$RUN_SHA" | cut -c1-8))"
    else
      pass "CI green on HEAD"
    fi
  else
    warn "could not query CI runs (gh not authenticated?)"
  fi

  # 6. No open release-blocker issues — the workflow refuses to release
  #    while any exist, so find out now instead of after the dispatch.
  BLOCKERS=$(gh issue list --label release-blocker --state open \
    --json number,title --jq 'length' 2>/dev/null || echo "?")
  if [ "$BLOCKERS" = "0" ]; then
    pass "no open release-blocker issues"
  elif [ "$BLOCKERS" = "?" ]; then
    warn "could not query release-blocker issues"
  else
    fail "$BLOCKERS open release-blocker issue(s):"
    gh issue list --label release-blocker --state open --json number,title \
      --jq '.[] | "       #\(.number): \(.title)"' || true
  fi
else
  warn "gh CLI not found — skipping CI/blocker checks"
fi

echo
if [ "$FAIL" != "0" ]; then
  echo "Pre-flight FAILED — fix the above before releasing."
  exit 1
fi
echo "Pre-flight passed — clear to dispatch the release."
