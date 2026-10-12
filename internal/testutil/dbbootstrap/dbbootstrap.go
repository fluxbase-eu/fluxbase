// Package dbbootstrap prepares a Fluxbase test database: reset, bootstrap
// roles + extensions, then the full declarative internal schema.
//
// The internal schema (auth.*, platform.*, app.*, ...) is managed declaratively
// and is NOT created by Connection.Migrate() (which only handles user
// filesystem migrations). Package-internal integration tests — e.g.
// internal/auth — that talk to the schema directly therefore need the same
// provisioning that test/e2e's TestMain performs. This package holds that
// logic so internal packages can reuse it without creating import cycles: it
// deliberately depends only on low-level packages (config, database,
// bootstrap, schema) and applies the embedded schema files directly (the same
// SQL the declarative direct-fallback engine and the CI psql provisioning
// steps run) instead of pulling in internal/migrations, which transitively
// imports half the application.
package dbbootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/database/bootstrap"
	dbschema "github.com/nimbleflux/fluxbase/internal/database/schema"
)

// Config describes the database to bootstrap and the admin credentials to do
// it with. User/Password identify the runtime (app) user the schema is
// provisioned for; AdminUser/AdminPassword must be able to create roles and
// schemas.
type Config struct {
	Host          string
	Port          int
	Database      string
	User          string
	Password      string
	AdminUser     string
	AdminPassword string
}

// Ensure provisions the internal schema on the configured database. To stay
// deterministic on databases that are only partially provisioned (or hold
// leftovers from another suite), it first drops all non-system schemas — the
// same reset test/e2e's TestMain performs — then runs the bootstrap SQL and
// applies the declarative schema files.
func Ensure(ctx context.Context, cfg Config) error {
	adminDB, err := database.NewConnection(connectionConfig(cfg, cfg.AdminUser, cfg.AdminPassword))
	if err != nil {
		return fmt.Errorf("dbbootstrap: failed to connect as admin user: %w", err)
	}
	defer adminDB.Close()

	if err := resetSchemas(ctx, adminDB, cfg.Database); err != nil {
		return err
	}

	// Bootstrap creates schemas, extensions and roles. It grants the created
	// roles to CURRENT_USER (the admin), so it must run through the admin pool.
	bootstrapSvc := bootstrap.NewServiceWithConfig(adminDB.Pool(), bootstrap.Config{
		Host:          cfg.Host,
		Port:          cfg.Port,
		Database:      cfg.Database,
		User:          cfg.User,
		Password:      cfg.Password,
		AdminUser:     cfg.AdminUser,
		AdminPassword: cfg.AdminPassword,
	})
	if err := bootstrapSvc.EnsureBootstrap(ctx); err != nil {
		return fmt.Errorf("dbbootstrap: bootstrap SQL failed: %w", err)
	}
	bootstrapSvc.Close()

	// Apply the declarative schema files (tables, indexes, functions,
	// policies) on the freshly reset schema landscape.
	schemaDir, err := dbschema.ExtractSchemas()
	if err != nil {
		return fmt.Errorf("dbbootstrap: failed to extract embedded schemas: %w", err)
	}
	defer func() { _ = os.RemoveAll(schemaDir) }()

	for _, schema := range schemaApplyOrder {
		if err := applySchemaFile(ctx, adminDB, cfg, schemaDir, schema+".sql"); err != nil {
			return err
		}
	}
	for _, extra := range []string{"post-schema-fks.sql", "post-schema.sql"} {
		if err := applySchemaFile(ctx, adminDB, cfg, schemaDir, extra); err != nil {
			return err
		}
	}

	// Bootstrap granted the roles to the admin; the app user needs them too
	// for SET ROLE support. tenant_migration_role is included for parity with
	// the e2e provisioning even though only some packages use it.
	_, err = adminDB.Exec(ctx, fmt.Sprintf(
		`GRANT anon, authenticated, service_role, tenant_service, tenant_migration_role TO %s`,
		quoteIdentifier(cfg.User),
	))
	if err != nil {
		// Non-fatal: the role membership may already exist, or the role may be
		// missing on stripped-down databases that don't exercise SET ROLE.
		log.Warn().Err(err).Str("user", cfg.User).Msg("dbbootstrap: failed to grant role memberships")
	}

	return nil
}

// resetSchemas drops all non-system schemas and recreates public, mirroring
// test/e2e's resetTestDatabase. The schema files are only guaranteed clean on
// a fresh schema landscape (CREATE POLICY has no IF NOT EXISTS), so any
// pre-existing state must go. Extension-managed schemas are preserved.
func resetSchemas(ctx context.Context, adminDB *database.Connection, dbName string) error {
	_, err := adminDB.Exec(ctx, `DO $$ DECLARE r RECORD; BEGIN FOR r IN SELECT nspname FROM pg_namespace WHERE nspname NOT IN (
		'pg_catalog', 'information_schema', 'pg_toast',
		'_timescaledb_cache', '_timescaledb_catalog', '_timescaledb_config',
		'_timescaledb_functions', '_timescaledb_internal',
		'timescaledb_experimental', 'timescaledb_information'
	) LOOP EXECUTE 'DROP SCHEMA IF EXISTS ' || quote_ident(r.nspname) || ' CASCADE'; END LOOP; END $$`)
	if err != nil {
		return fmt.Errorf("dbbootstrap: failed to drop existing schemas: %w", err)
	}
	if _, err := adminDB.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS public; GRANT ALL ON SCHEMA public TO public`); err != nil {
		return fmt.Errorf("dbbootstrap: failed to recreate public schema: %w", err)
	}
	log.Info().Str("database", dbName).Msg("dbbootstrap: reset existing schemas")
	return nil
}

// schemaApplyOrder lists the embedded schema files in dependency order,
// mirroring migrations.DefaultFluxbaseSchemas (schemas without a file —
// system, migrations, api — are skipped by applySchemaFile).
var schemaApplyOrder = []string{
	"platform", "auth", "storage", "jobs", "functions", "realtime",
	"ai", "rpc", "app", "branching", "logging", "mcp",
}

// applySchemaFile executes one schema SQL file through the admin pool.
func applySchemaFile(ctx context.Context, adminDB *database.Connection, cfg Config, dir, name string) error {
	file := filepath.Join(dir, name)
	sql, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		// Not every managed schema has a file.
		return nil
	}
	if err != nil {
		return fmt.Errorf("dbbootstrap: failed to read schema file %s: %w", file, err)
	}
	// No-argument Exec uses the simple query protocol, which can run the
	// whole multi-statement file in one round trip.
	// The files carry {{APP_USER}} placeholders (see the psql provisioning in
	// CI) that must be bound to the runtime user first.
	applySQL := strings.ReplaceAll(string(sql), "{{APP_USER}}", cfg.User)
	if _, err := adminDB.Exec(ctx, applySQL); err != nil {
		return fmt.Errorf("dbbootstrap: failed to apply schema file %s: %w", file, err)
	}
	return nil
}

// connectionConfig builds the connection config for connecting as the given
// user. Tests run with sslmode=disable.
func connectionConfig(cfg Config, user, password string) config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:           cfg.Host,
		Port:           cfg.Port,
		User:           user,
		Password:       password,
		Database:       cfg.Database,
		SSLMode:        "disable",
		AdminUser:      cfg.AdminUser,
		AdminPassword:  cfg.AdminPassword,
		MaxConnections: 2,
		MinConnections: 0,
		HealthCheck:    30 * time.Second,
	}
}

// quoteIdentifier quotes a PostgreSQL identifier (defense in depth; the inputs
// come from test env vars, but role names are identifiers, not strings).
func quoteIdentifier(name string) string {
	quoted := `"`
	for _, r := range name {
		if r == '"' {
			quoted += `""`
		} else {
			quoted += string(r)
		}
	}
	return quoted + `"`
}
