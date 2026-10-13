#!/usr/bin/env bash
# Single source of truth for propagating a release version into every file
# that carries one. The release workflow (update-version-files job) and the
# local `make bump-*` targets both call this script — one code path, so CI
# and local can't drift (they did once; that's why this script exists).
#
# Usage:
#   scripts/set-version.sh <version>          Apply the version to all files
#   scripts/set-version.sh <version> --check  Exit 1 if any file would change
#
# The --check mode is what the `version-drift` CI job runs against the
# latest release tag: on a clean tree, re-applying the last released version
# must be a no-op.
#
# Semantics (matching long-standing release behavior):
#   - Always updated: Chart.yaml version+appVersion, SDK package.json
#     versions + peerDeps, admin/package.json, docs CLI install pin,
#     Helm values.yaml postgres image tag (from postgres/versions.json).
#   - Stable releases only: VERSION file, docker-compose fallback versions,
#     deploy/.env.example (RCs must not move the "current stable" pointers).
#   - bun.lock is refreshed via `bun install` when bun is available.
set -euo pipefail

NEW_VERSION="${1:-}"
CHECK=0
[[ "${2:-}" == "--check" ]] && CHECK=1

if [[ -z "$NEW_VERSION" ]]; then
  echo "usage: set-version.sh <version> [--check]" >&2
  exit 2
fi
if [[ ! "$NEW_VERSION" =~ ^[0-9]{4}\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ ]]; then
  echo "error: '$NEW_VERSION' is not a CalVer version (expected YYYY.M.N[-rc.N])" >&2
  exit 2
fi

TAG="v${NEW_VERSION}"
STABLE_VERSION="${NEW_VERSION%%-rc*}"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

DRIFTED=()
UPDATED=()

# apply_sed <file> <sed-script> — rewrite via temp file so --check never
# touches the tree and drift is detectable.
apply_sed() {
  local file="$1" script="$2"
  if [[ ! -f "$file" ]]; then
    echo "warning: $file not found, skipping" >&2
    return
  fi
  local tmp
  tmp="$(mktemp)"
  sed -E "$script" "$file" > "$tmp"
  if ! cmp -s "$tmp" "$file"; then
    DRIFTED+=("$file")
    if [[ $CHECK -eq 0 ]]; then
      mv "$tmp" "$file"
      UPDATED+=("$file")
    else
      rm -f "$tmp"
    fi
  else
    rm -f "$tmp"
  fi
}

# apply_jq <file> <jq-filter> — same contract as apply_sed.
apply_jq() {
  local file="$1" filter="$2"
  if [[ ! -f "$file" ]]; then
    echo "warning: $file not found, skipping" >&2
    return
  fi
  local tmp
  tmp="$(mktemp)"
  jq --arg version "$NEW_VERSION" "$filter" "$file" > "$tmp"
  if ! cmp -s "$tmp" "$file"; then
    DRIFTED+=("$file")
    if [[ $CHECK -eq 0 ]]; then
      mv "$tmp" "$file"
      UPDATED+=("$file")
    else
      rm -f "$tmp"
    fi
  else
    rm -f "$tmp"
  fi
}

# --- always updated ---------------------------------------------------------

apply_sed deploy/helm/fluxbase/Chart.yaml \
  "s/^version:.*/version: ${NEW_VERSION}/"
apply_sed deploy/helm/fluxbase/Chart.yaml \
  "s/^appVersion:.*/appVersion: \"${NEW_VERSION}\"/"

# Postgres image tag pinned by the Helm chart, always kept in sync with the
# latest immutable tag from versions.json (this drifted silently once).
apply_sed deploy/helm/fluxbase/values.yaml \
  "/fluxbase-postgres/,/pullPolicy/{s/tag: \"[^\"]*\"/tag: \"$(jq -r '.versions[] | select(.latest == true) | "\(.version)-v\(.imageVersion)"' postgres/versions.json)\"/}"

for dir in sdk sdk-react sdk-svelte sdk-next sdk-vue admin; do
  apply_jq "${dir}/package.json" '.version = $version'
done
# Framework SDKs peer-depend on the core SDK with a compatible range.
for dir in sdk-react sdk-svelte sdk-next sdk-vue; do
  apply_jq "${dir}/package.json" \
    '.peerDependencies["@nimbleflux/fluxbase-sdk"] = "^\($version)"'
done

apply_sed docs/src/content/docs/cli/installation.md \
  "s#install-cli\.sh \| bash -s -- v[0-9.]*-*[a-z0-9.]*#install-cli.sh | bash -s -- ${TAG}#g"

# --- stable releases only ---------------------------------------------------

if [[ "$NEW_VERSION" == "$STABLE_VERSION" ]]; then
  if [[ ! -f VERSION || "$(cat VERSION)" != "$NEW_VERSION" ]]; then
    if [[ $CHECK -eq 0 ]]; then
      echo "$NEW_VERSION" > VERSION
      UPDATED+=("VERSION")
    else
      DRIFTED+=("VERSION")
    fi
  fi

  apply_sed deploy/docker-compose.yml \
    "s/\\\$\{FLUXBASE_VERSION:-[^}]*\}/\\\$\{FLUXBASE_VERSION:-${NEW_VERSION}\}/g"
  apply_sed deploy/docker-compose.minimal.yaml \
    "s/\\\$\{FLUXBASE_VERSION:-[^}]*\}/\\\$\{FLUXBASE_VERSION:-${NEW_VERSION}\}/g"
  apply_sed deploy/.env.example \
    "s/^# FLUXBASE_VERSION=.*/# FLUXBASE_VERSION=${NEW_VERSION}/g"
fi

# --- bun.lock ---------------------------------------------------------------

if [[ $CHECK -eq 0 ]] && command -v bun > /dev/null 2>&1; then
  # Refresh the workspace lockfile for the new package.json versions.
  bun install > /dev/null
  echo "refreshed: bun.lock"
elif [[ $CHECK -eq 0 ]]; then
  echo "warning: bun not found — bun.lock was NOT refreshed; run 'bun install' before committing" >&2
fi

# --- report -----------------------------------------------------------------

if [[ ${#DRIFTED[@]} -gt 0 ]]; then
  if [[ $CHECK -eq 1 ]]; then
    echo "Version drift detected in: ${DRIFTED[*]}" >&2
    echo "Run 'scripts/set-version.sh <version>' to fix, then commit." >&2
    exit 1
  fi
  echo "Applied ${NEW_VERSION} to: ${UPDATED[*]}"
else
  if [[ $CHECK -eq 1 ]]; then
    echo "No version drift (all files already match ${NEW_VERSION})"
  else
    echo "No changes needed (already at ${NEW_VERSION})"
  fi
fi
