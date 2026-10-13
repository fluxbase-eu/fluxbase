# Contributing to Fluxbase

Thanks for contributing! This file covers the contributor workflow. The
full codebase guide (stack, module map, architecture, security patterns)
lives in [AGENTS.md](AGENTS.md) — start there before your first PR.

## Getting started

```bash
make setup-dev     # first-time dev environment setup
make dev           # fast dev: backend + frontend (Vite HMR)
make test          # Go test suite
make docs          # serve the docs site locally
```

The dev environment expects PostgreSQL with pgvector; `make setup-dev` and
the [devcontainer](.devcontainer/README.md) handle this for you.

## Commit messages

Commits follow the [Conventional Commits](https://www.conventionalcommits.org/)
spec, enforced by the lefthook `commit-msg` hook:

```
feat(ai): LLM-based entity extraction
fix(sdk): correct toggleChatbot verb
docs: add reranking guide
chore(ci): retry transient link-check failures
```

Types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `perf`, `ci`.
Scope with the affected area (`ai`, `sdk`, `mcp`, `api`, `storage`, `ci`, …).

## Pull requests

- Keep PRs focused; one logical change per PR.
- Update documentation for user-facing changes (see the "Updating
  Documentation" policy in [AGENTS.md](AGENTS.md)).
- API-surface changes must update the hand-curated OpenAPI spec
  (`internal/api/openapi*.go`) and regenerate `sdk/src/types-generated.ts`
  via `scripts/check-contract-drift.sh --update` against a running server —
  CI's `contract-drift` job enforces this.
- The PR template includes a documentation checkbox; use it.

## Testing

- Go: `make test`; AI/MCP packages: `go test ./internal/ai/... ./internal/mcp/...`
- SDKs: `npm run test` in `sdk/`, `sdk-react/`, `sdk-svelte/`, `sdk-vue/`;
  `./gradlew test` in `sdk-kotlin/`.
- E2E: see `test/README.md` for the full suite and environment requirements.

## Changelog & releases

CHANGELOG.md and GitHub release notes are **generated from the
conventional-commit history** by [git-cliff](https://git-cliff.org)
(`cliff.toml`) at release time — do not hand-edit release sections, and do
not add "Unreleased" entries. Your job is the commit message: follow
[Conventional Commits](https://www.conventionalcommits.org) (enforced by
lefthook) and mark breaking changes with `!` or a `BREAKING CHANGE:` footer;
they surface automatically in the notes.

Releases are cut from `main` via the Release workflow — `make release`
(TYPE=rc|stable, DRY_RUN=true to rehearse). The workflow derives the next
CalVer tag, generates the notes, and publishes everything. The full runbook
is in [RELEASING.md](RELEASING.md); preview notes locally with
`make release-notes`.

## Code style

- Go: `gofmt` + `golangci-lint run ./...` (enforced pre-commit)
- TypeScript: ESLint + Prettier per package
- Match the comment density and naming of the surrounding code

## Reporting issues

Open a GitHub issue with reproduction steps. Security issues: see
[SECURITY](docs/src/content/docs/security/overview.md) guidance before
posting details publicly.
