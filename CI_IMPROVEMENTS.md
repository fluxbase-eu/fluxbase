# CI Process Improvements — Analysis & Recommendations

**Date:** 2026-10-11 · **Measured on:** main run `38178542404` (15 jobs, post-remediation merges) · Local file, not for publication.

---

## 1. Measured current state

**Wall clock: 21.6 min · Total compute: 77 job-minutes across 15 jobs.**

| Job | Duration | Starts at | Gated on |
|---|---|---|---|
| Test Go | 12.8m | +0.1m | `changes.go` |
| Test E2E UI (Playwright) | 9.2m | +0.1m | `changes.ts/admin` |
| Docker Build Test (lite) | 8.6m | **+13.0m** | `needs: [changes, lint-go, lint-typescript, test-sdk, test-go, test-e2e-ui]` |
| Docker Build Test (full) | 8.3m | **+13.0m** | same |
| Docker Docs Build Test | 5.8m | +1.1m | `needs: [changes, test-sdk]` (already well-gated) |
| Lint Go | 5.8m | +0.1m | `changes.go` |
| Go Tests (ocr vips tags) | 5.4m | **+13.0m** | same as Docker |
| Setup Smoke (lite/full) | 5.3m / 4.6m | +0.1m | `changes` |
| Test Go (DB-backed, non-short) | 4.3m | +0.1m | `changes.go` |
| Build matrix (amd64/arm64) | 4.2m / 1.3m | +13.0m | same as Docker |
| Test SDK | 0.9m | +0.1m | `changes.ts` |
| Lint TypeScript | 0.6m | +0.1m | `changes.ts` |

**Critical path = Test Go (12.8m) → second wave (8.6m) = 21.6m.** The second wave (Docker builds, cross-compile build matrix, ocr-tags tests) waits for **all five** test/lint jobs even though it validates artifacts the compiler already built.

**Inside Test Go (12.8m):** ~90s setup (services: Postgres, SeaweedFS, Dex, Deno) + **338s unit tests** (`go test -short -race` over all packages, with coverage) + **312s Go-integration e2e** (`-tags=integration -race -parallel=1` — fully serial, domain-selected) + gate. `-race` roughly doubles CPU cost on both halves.

**Existing strengths worth keeping:** dorny/paths-filter stage selection already scopes go/ts/docker/docs jobs; buildx GHA cache is configured per flavor; setup-go/setup-node caches are on; the docs image job is gated on `test-sdk` only and starts at +1.1m — the model the second wave should follow.

**Gaps observed:** no `concurrency:` group anywhere (superseded pushes run to completion — this session alone burned several duplicate ~20-minute runs); no Playwright sharding/browser cache; Lint Go installs golangci-lint raw (no cache layer); the ocr-tags job runs on every Go change though it only validates cgo/vips builds.

---

## 2. Recommendations (ranked by impact/effort)

### Quick wins (config-only, no structural change)

1. **Add concurrency cancellation** — `concurrency: { group: ci-${{ github.ref }}, cancel-in-progress: true }` to ci.yml (and fresh-boot.yml). Saves 10–20 min of wasted compute per superseded push. Zero risk.
2. **De-gate the second wave** — change `needs` for `docker`, `build`, `test-go-build-tags` to `needs: [changes, lint-go]` (drop test-go/test-e2e-ui/test-sdk). Branch protection already requires those tests for merging, so the gate still holds; the builds validate compilation in parallel. **Wall clock 21.6m → ~13m.** Trade-off: up to ~20 wasted compute-min on red PRs.
3. **Path-filter `test-go-build-tags` (ocr/vips)** to `internal/storage/**`, `internal/ai/**`, `Dockerfile*`, and the workflow file — it re-runs the Go suite with cgo tags and is only relevant when vips-dependent code or the image changes. Saves 5.4m compute on most runs.
4. **Cache Playwright browsers** (`actions/cache` on `~/.cache/ms-playwright`) and **shard the suite** (`--shard=1/3` on 3 runners via matrix). E2E UI 9.2m → ~3.5–4m wall.

### Structural (worth a small PR each)

5. **Split `Test Go` into two parallel jobs**: `test-go-unit` (338s unit suite) and `test-go-integration` (312s serial e2e). Both need the same service containers (~90s setup each), but running in parallel cuts the 12.8m job to ~6.5m wall. Then shard the integration job across 2 runners using the **domain bucket selection that already exists** in the workflow (auth/ai/api/tenantdb/webhook FILES lists) → ~3–4m per shard.
6. **golangci-lint via `golangci/golangci-lint-action`** instead of raw install — its built-in cache typically cuts a cold 5.8m lint to ~2–3m. (Note the known v2.10.0/Go-1.27 incompatibility: bump the pinned linter version at the same time, or the action gains nothing.)
7. **Enable GitHub merge queue** once branch protection is tuned: PRs merge in batches, each batch validated against the true post-merge state — eliminates "main broke after merge" class and lets you drop the sequential update-branch dance we just did manually.
8. **Required-checks hygiene** after step 2: keep `test-go`, `test-e2e-ui`, `lint-go` required; do *not* require the de-gated Docker/build jobs (or require them and accept the same gate — decide by whether an image-build breakage on main matters to you; fresh-boot.yml on main already catches that).

**Projected result:** wall clock ~21.6m → **~10–12m**, compute 77 → ~55–60 job-minutes per run, plus the duplicate-run savings from concurrency cancellation.

---

## 3. Test-pyramid gaps → what to add (from the audit + this remediation)

| Gap | Evidence | What to add |
|---|---|---|
| **admin/ has zero unit tests** (Playwright-only, 67k LOC) | audit TS-06 | vitest + Testing Library targeting `lib/api/client.ts` (the 401-refresh queue — pure logic, highest value), `stores/auth-store.ts`, `lib/auth.ts` session restore. These are exactly the paths the SEC-06/TS-04 changes touched. |
| **Cross-tenant guard only tested against a fake** | SEC-01 regression tests use a mocked checker | One DB-backed integration test in `test-go-db`: seed two tenants + a tenant-admin JWT, hit `POST /admin/tenants/{other}/admins`, assert 403. Catches SQL/policy-level bypasses the fake cannot. |
| **`internal/secrets` at 1.6%** | INF-02 | Unit tests for the AES-GCM round-trip + dynamic SET-clause builder (pure functions, no DB). |
| **`internal/database/bootstrap` at 9.2%** | the CI failure we just fixed | Identifier/quote validation functions are pure — trivial table tests; ratchet the new floor 8 → 10+ after. |
| **Contract drift has no CI net** | API-01 (four runtime SDK bugs) | CI job: start server → `GET /openapi.json` → generate types (`openapi-typescript`) → `git diff --exit-code` against a committed generated file. Turns the whole drift class into a red build. |
| **SDK Go↔TS integration test** | — | One e2e: run the server, drive it with the real SDK (`jobs.list`, `storage.list`, `upload(metadata)`, `listUsers`) asserting the just-fixed wire shapes. Cheaper than it sounds — the server fixture already exists in the Playwright job. |
| **E2E: UI-level cross-tenant denial + logout** | SEC-01, SEC-06 | Playwright: tenant-admin logs in → attempts tenant-B admin page → forbidden state; logout clears the session cookie. |
| Race/cost balance | — | Keep `-race` on PR unit tests; consider dropping `-parallel=1` only after tenant-isolation tests are verified shard-safe (they exist precisely because of shared-DB state). |

---

## 4. Suggested execution order

1. concurrency + second-wave de-gating + ocr path-filter (one small PR, ~13m pipelines same day)
2. Playwright sharding + browser cache; golangci-lint-action (+ linter version bump)
3. Test Go split + integration sharding
4. merge queue + branch-protection retune
5. The test additions above, interleaved with feature work (admin unit tests first)
