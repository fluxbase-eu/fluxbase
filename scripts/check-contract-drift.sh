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

# The committed file is openapi-typescript's RAW output, and it is diffed
# raw-vs-raw: the generator is a pure function of the spec JSON, so no
# formatter may touch the file — a prettier pass here silently no-op'd in CI
# (exit 0, unformatted output) and produced phantom drift. lefthook's
# pre-commit prettier command excludes this file for the same reason.
echo "==> Sanity-checking generated output"
if ! grep -q '^export interface paths' "$GENERATED"; then
  echo "INTERNAL ERROR: generator output does not look like openapi-typescript output." >&2
  exit 1
fi

if [ "$UPDATE" = "1" ]; then
  mkdir -p "$(dirname "$COMMITTED")"
  cp "$GENERATED" "$COMMITTED"
  echo "==> Wrote $COMMITTED ($(wc -l < "$COMMITTED") lines). Review and commit it."
  exit 0
fi

echo "==> Diffing regenerated types against $COMMITTED"
# Save the check inputs/outputs BEFORE anything can fail below.
if [ -n "$DRIFT_ARTIFACTS" ]; then
  cp -f "$SPEC" "$GENERATED" "$DRIFT_ARTIFACTS/" 2>/dev/null || true
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

if diff -u "$COMMITTED" "$GENERATED"; then
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
