# Fluxbase Scripts

Helper scripts for development, testing, and operations. Each script
documents its own flags in its header comment.

## setup-dev-user.sql

This script grants all necessary permissions to the `fluxbase_app` database user for local development and testing.

### Why is this needed?

For proper Row Level Security (RLS) enforcement, the application must use a **non-superuser** database connection. PostgreSQL superusers (like `postgres`) bypass RLS policies, which would make RLS tests ineffective.

The `fluxbase_app` user is a regular (non-superuser) user that:
- Has all necessary permissions to run migrations and operate the application
- **Cannot** bypass RLS policies, ensuring they work correctly
- Owns all tables and sequences (required for operations like upsert)

### When to run this script

Run this script after setting up your local development database:

```bash
# Run bootstrap (server will apply declarative schema on startup)
PGPASSWORD=postgres psql -h localhost -U postgres -d fluxbase_dev -f internal/database/bootstrap/bootstrap.sql

# Then grant permissions to fluxbase_app
PGPASSWORD=postgres psql -h localhost -U postgres -d fluxbase_dev -f scripts/setup-dev-user.sql
```

**Note:** Internal Fluxbase schema (auth, storage, functions, etc.) is managed declaratively and applied automatically when the server starts. For user-provided migrations in the `public` schema, you can still use imperative migrations with `make migrate-up`.

### What it does

1. Grants CREATE permission on the database
2. Grants ALL permissions on all schemas (public, auth, storage, etc.)
3. Grants ALL PRIVILEGES on all tables, sequences, and functions
4. Changes ownership of all tables and sequences to `fluxbase_app`
5. Changes ownership of all schemas to `fluxbase_app`

### Configuration

After running this script, update your `.env` file to use the `fluxbase_app` user:

```env
FLUXBASE_DATABASE_USER=fluxbase_app
FLUXBASE_DATABASE_PASSWORD=fluxbase_app_password
FLUXBASE_DATABASE_DATABASE=fluxbase_dev
```

This ensures RLS works correctly in both development and tests.

---

## Git Hooks

### Pre-commit Hook

The pre-commit hook automatically runs before each commit to catch issues early and reduce CI failures.

**Checks performed:**
1. **Go Formatting** - Runs `go fmt ./...` on all staged Go files and auto-stages formatting changes
2. **Admin TypeScript** - Type checks admin UI code (`cd admin && npm run type-check`)
3. **SDK TypeScript** - Type checks SDK code (`cd sdk && npm run type-check`)
4. **SDK React TypeScript** - Type checks React SDK code (`cd sdk-react && npm run type-check`)

**Installation:**
```bash
# Automatically installed with:
make setup-dev

# Or install manually:
make install-hooks
```

**Usage:**
The hook runs automatically on `git commit`. If any checks fail, the commit is blocked.

To bypass the hook (not recommended):
```bash
git commit --no-verify
```

**Uninstall:**
```bash
make uninstall-hooks
```

**Benefits:**
- Catches formatting and type errors before they reach CI
- Auto-formats Go code
- Reduces failed CI pipeline runs
- Provides fast feedback during development

## Operational Scripts

### backup.sh

Creates backups of both the PostgreSQL database (`pg_dump`, custom format) and
storage files. See the header of the script for the full flag list and the
[backup & restore guide](../docs/src/content/docs/guides/backup-restore.md).

### restore.sh

Restores Fluxbase database and storage from backups produced by `backup.sh`.
Pair with the backup & restore guide above.

### quick-start.sh

Generates required secrets (JWT, encryption key) and optionally starts
Fluxbase against the local Docker PostgreSQL. The fastest path to a running
instance outside Docker Compose.

### smoke-test.sh

Verifies a running Fluxbase stack actually works (health, auth, REST CRUD,
storage). Brings nothing up itself — pair it with
`deploy/docker-compose.smoke.yml`; CI uses the same pairing.

### start-e2e-ui.sh

Starts the Playwright admin-UI E2E environment: a Go backend on `:8082`
against the `fluxbase_playwright` database plus the admin UI dev server.

### test-runner.sh

Detailed test-summary runner used by the Makefile `test` targets. Prefer
`make test`; call this directly only when you want the richer output format.

### check-contract-drift.sh

Compares the OpenAPI spec served by a running Fluxbase server against the
committed `sdk/src/types-generated.ts` and regenerates the file with
`--update`. CI's `contract-drift` job runs it against a fresh boot; see the
script header for the admin-token environment variables.

### generate-migration-service-key.sh + create-migration-service-key.sql

Generates a secure random service key and inserts it into the database for
schema-migration use. The SQL file is the standalone variant; the shell
script wires generation + insertion together.

### viz-architecture.sh / viz-workflows.sh

Generate dependency-graph diagrams of the codebase (Go call graph via
`go mod`/package analysis) for architecture review. Output lands in the
working directory; not part of CI.

### loadtest/

`main.go` is a minimal stdlib-only HTTP load generator used for memory and
throughput baselines (see the memory-baseline notes in AGENTS.md):

```bash
go run ./scripts/loadtest -url http://localhost:8080     -duration 60s -concurrency 50 -path /health
```
