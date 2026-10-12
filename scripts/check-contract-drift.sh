#!/usr/bin/env bash
# =============================================================================
# Contract-drift check: Go OpenAPI spec <-> committed TypeScript types
#
# The TypeScript SDK types in sdk/src/types-generated.ts must always match the
# OpenAPI spec that the Go server actually serves. Go structs and TS types are
# edited in the same PR but nothing previously verified them together — this
# gap produced four runtime bugs (issue #382). This script fetches the spec
# from a RUNNING server, regenerates the TS types, and diffs them against the
# committed file (issues #408 / #389).
#
# Usage:
#   scripts/check-contract-drift.sh [BASE_URL]              # check (CI mode)
#   scripts/check-contract-drift.sh --update [BASE_URL]     # regenerate commit file
#
# Environment:
#   CONTRACT_DRIFT_TOKEN    Admin JWT. The server serves the FULL spec
#                           (including database-table paths) only to admin
#                           roles; without a token the minimal unauthenticated
#                           spec is checked. The token must come from the same
#                           server instance as BASE_URL.
#   CONTRACT_DRIFT_BASE_URL Fallback for BASE_URL when $1 is unset
#                           (default: http://localhost:8080).
#
# The script starts nothing itself: point it at an already-running Fluxbase
# server. The `contract-drift` CI job boots one (postgres + freshly built
# binary, matching the test-e2e-ui job's setup).
# =============================================================================
set -euo pipefail

UPDATE=0
if [ "${1:-}" = "--update" ]; then
  UPDATE=1
  shift
fi
BASE_URL="${1:-${CONTRACT_DRIFT_BASE_URL:-http://localhost:8080}}"
COMMITTED="sdk/src/types-generated.ts"
PLACEHOLDER_MARKER="CONTRACT-DRIFT-PLACEHOLDER"

cd "$(dirname "$0")/.."

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
SPEC="$WORK/openapi.json"
GENERATED="$WORK/types-generated.ts"
COMMITTED_NORM="$WORK/types-committed.ts"

# Keep the inputs/outputs of the check in a stable location (CI uploads this
# on failure) so a drift is reviewable without re-running the job.
DRIFT_ARTIFACTS="${TMPDIR:-/tmp}/openapi-drift"
mkdir -p "$DRIFT_ARTIFACTS" 2>/dev/null || DRIFT_ARTIFACTS=""

AUTH=()
if [ -n "${CONTRACT_DRIFT_TOKEN:-}" ]; then
  AUTH=(-H "Authorization: Bearer $CONTRACT_DRIFT_TOKEN")
fi

echo "==> Fetching OpenAPI spec from ${BASE_URL}/openapi.json"
curl -fsS "${AUTH[@]}" "$BASE_URL/openapi.json" -o "$SPEC"
if ! grep -q '"openapi"' "$SPEC"; then
  echo "ERROR: $BASE_URL/openapi.json did not return an OpenAPI document" >&2
  exit 1
fi

echo "==> Regenerating TS types with openapi-typescript"
bunx openapi-typescript "$SPEC" -o "$GENERATED"

# Normalize both sides with prettier (same formatter the lefthook pre-commit
# hook applies to sdk/src/**) so formatting can neither mask nor fake drift.
# Prefer the workspace-installed binary: going through `bunx` here has been
# observed to silently no-op in CI, which would report phantom drift.
echo "==> Normalizing formatting (prettier)"
if [ -x "node_modules/.bin/prettier" ]; then
  PRETTIER=(node_modules/.bin/prettier --no-config --log-level warn)
else
  PRETTIER=(bunx prettier --no-config --log-level warn)
fi
"${PRETTIER[@]}" "$GENERATED" > "$WORK/generated.pretty"
if [ -f "$COMMITTED" ]; then
  "${PRETTIER[@]}" "$COMMITTED" > "$COMMITTED_NORM"
fi

# Canary: the raw generator emits 4-space indent, prettier emits 2-space.
# Path keys ("/...") only sit at depth 1 (2 spaces) in normalized output.
# If this fires, normalization silently failed — abort instead of reporting
# phantom drift.
if grep -q '^    "/' "$WORK/generated.pretty"; then
  echo "INTERNAL ERROR: prettier normalization did not run; cannot check drift." >&2
  echo "Run 'bun install' at the repo root so node_modules/.bin/prettier exists," >&2
  echo "then retry." >&2
  exit 1
fi

if [ "$UPDATE" = "1" ]; then
  mkdir -p "$(dirname "$COMMITTED")"
  cp "$WORK/generated.pretty" "$COMMITTED"
  echo "==> Wrote $COMMITTED ($(wc -l < "$COMMITTED") lines). Review and commit it."
  exit 0
fi

echo "==> Diffing regenerated types against $COMMITTED"
if [ -n "$DRIFT_ARTIFACTS" ]; then
  cp -f "$SPEC" "$WORK/generated.pretty" "$DRIFT_ARTIFACTS/" 2>/dev/null || true
  [ -f "$COMMITTED_NORM" ] && cp -f "$COMMITTED_NORM" "$DRIFT_ARTIFACTS/" 2>/dev/null || true
fi
if [ ! -f "$COMMITTED" ]; then
  echo "ERROR: $COMMITTED does not exist." >&2
  echo "       Generate it against a running server:" >&2
  echo "         CONTRACT_DRIFT_TOKEN=<admin-jwt> scripts/check-contract-drift.sh --update <base-url>" >&2
  exit 1
fi

if grep -q "$PLACEHOLDER_MARKER" "$COMMITTED"; then
  echo "WARNING: $COMMITTED still contains the initial placeholder ($PLACEHOLDER_MARKER)." >&2
  echo "         No baseline exists yet, so drift cannot be detected. Generate a" >&2
  echo "         real baseline: CONTRACT_DRIFT_TOKEN=<admin-jwt> scripts/check-contract-drift.sh --update <base-url>" >&2
  exit 0
fi

if diff -u "$COMMITTED_NORM" "$WORK/generated.pretty"; then
  echo "OK: sdk/src/types-generated.ts matches the served OpenAPI spec."
else
  echo "" >&2
  echo "CONTRACT DRIFT DETECTED: sdk/src/types-generated.ts is out of date with" >&2
  echo "the OpenAPI spec served by $BASE_URL." >&2
  echo "" >&2
  echo "The Go API surface and the TypeScript SDK types moved apart. Fix:" >&2
  echo "  1. Update the SDK types (or the Go handlers) in the same PR:" >&2
  echo "       CONTRACT_DRIFT_TOKEN=<admin-jwt> scripts/check-contract-drift.sh --update <base-url>" >&2
  echo "     then review and commit the regenerated sdk/src/types-generated.ts." >&2
  echo "  2. Never hand-edit sdk/src/types-generated.ts — it is generated." >&2
  exit 1
fi
