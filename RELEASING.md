# Releasing Fluxbase

Everything about cutting a release happens through the **Release** workflow
(`.github/workflows/release.yml`). Tags are created by CI, never by hand —
pushing a `v*` tag yourself does nothing, and there is no local release path
that publishes artifacts.

## Quick start

```bash
# Pre-flight: CI green, no release blockers, version files in sync
make release-check

# Rehearse the whole release without publishing anything
make release DRY_RUN=true

# Cut a release candidate
make release              # TYPE defaults to rc

# Promote to stable (or cut a stable directly)
make release TYPE=stable
```

Each of these dispatches the Release workflow and streams its progress. The
equivalent UI path is Actions → Release → Run workflow.

## Versioning

CalVer `vYYYY.M.N[-rc.N]`, derived automatically from the tag history — you
never pick a number:

| Last release | You run          | Result                     |
|--------------|------------------|----------------------------|
| `v2026.9.10` | `rc`             | `v2026.9.11-rc.1`          |
| `v2026.9.10` | `stable`         | `v2026.9.11`               |
| `…-rc.3`     | `rc`             | `…-rc.4` (continue series) |
| `…-rc.3`     | `stable`         | `v2026.9.11` (promote)     |

The version-bump commit (VERSION, Helm chart, SDK `package.json`s, docs pin,
compose fallbacks — all via `scripts/set-version.sh`) is pushed back to the
branch you dispatched from, and the fully regenerated `CHANGELOG.md` comes
with it. On a clean tree, `scripts/set-version.sh <latest-tag> --check` is a
no-op; CI's `version-drift` job fails anything else.

## Release notes

Generated deterministically from conventional commits with
[git-cliff](https://git-cliff.org) (`cliff.toml`). Preview what the next
release will say with `make release-notes`. RC notes are incremental per RC
series; stable notes cover everything since the last published stable
release. An optional AI summary (`generate_ai_notes`, default off) is
garnish only — it can never change or block the release.

## What gets published

Component toggles (all default on, except postgres images):

| Toggle              | Produces                                                                 |
|---------------------|--------------------------------------------------------------------------|
| `publish_binaries`  | `fluxbase-server` tarballs (linux amd64/arm64, full OCR/vips feature set, SLSA-provenance attested) + `fluxbase` CLI (linux/darwin/windows) |
| `publish_docker`    | `ghcr.io/nimbleflux/fluxbase[:ver|-lite]` multi-arch, cosign-signed, SBOM + provenance |
| `publish_docs`      | `ghcr.io/nimbleflux/fluxbase-docs`                                       |
| `publish_npm_sdk`   | `@nimbleflux/fluxbase-sdk{,-react,-svelte,-next,-vue}`                   |
| `publish_kotlin_sdk`| `io.github.nimbleflux:fluxbase-kotlin` to GitHub Packages                |
| `publish_go_module` | pkg.go.dev indexing ping                                                 |
| `publish_helm`      | Chart to `oci://ghcr.io/nimbleflux/charts` + gh-pages index              |
| `publish_postgres`  | `ghcr.io/nimbleflux/fluxbase-postgres` from `postgres/versions.json`     |

Stable releases additionally move the rolling tags (`latest`, `major`,
`major.minor`, npm `latest` dist-tag); RCs get `rc` tags/dist-tags only.

## Failure recovery

Every publish step is **idempotent**: after a partial failure, re-run the
failed jobs from the run page (Re-run failed jobs) and already-published
pieces are skipped.

| Step                    | Safe to re-run because                                   |
|-------------------------|----------------------------------------------------------|
| create-release          | Skips if the tag or release already exists               |
| publish-binaries        | `gh release upload --clobber`                            |
| Docker arch / manifests | Digest-stable pushes; re-`imagetools create` is a no-op  |
| npm SDKs                | Skips versions already on the registry                   |
| Kotlin SDK              | Skips if the Maven POM already exists                    |
| Helm chart              | Skips if the chart version is already in the OCI repo    |
| update-version-files    | Commits nothing when the tree is already at the version  |

If post-release smoke tests fail, the run goes red **and a
`release-blocker` issue is opened automatically** — the next release refuses
to start while it's open. Fix, re-run, close the issue.

## Policies

- **RC first for risky changes.** Stable releases run full validation
  (green CI on the exact commit, no open `release-blocker`s); validation can
  only be skipped for RCs. `skip_validation=true` on a stable is rejected.
- **Rolling tags are immutable once released.** Never re-point `latest` at
  an old version by hand; cut a new release instead. Version-specific tags
  (`v…` / `:2026.9.10`) are permanently immutable.
- **Postgres images** are rebuilt only when `publish_postgres` is on
  (bump `postgres/versions.json` first). The Helm chart's pinned image tag
  always follows `versions.json`'s latest immutable tag, on every release.
- **One at a time.** The workflow has a `release` concurrency group: a
  second dispatch queues until the first finishes.
- **Who can release:** anyone who can dispatch the workflow on `main`;
  branch protection on `main` (green CI required) is the actual gate.

## Version references — single source of truth

`scripts/set-version.sh` is the only thing that writes version numbers into
files (VERSION, Chart.yaml `version`+`appVersion`, Helm postgres pin,
compose/.env defaults, SDK `package.json`s + peerDeps, admin
`package.json`, docs install pin). Locally via `make bump-patch|minor|major`,
in CI via the release workflow. Don't hand-edit those copies; if CI's
`version-drift` job complains, run the script.
