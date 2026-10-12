package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// ExecutionTenantResolver resolves the owning tenant of a stored execution
// record. Implementations must run the lookup under the caller's tenant
// context: a tenant-scoped caller can only resolve executions they own, while
// an empty tenant context (instance admin) resolves any execution.
type ExecutionTenantResolver interface {
	// ExecutionTenant returns the tenant that owns the execution and whether
	// an execution record exists and is visible to the caller's tenant scope.
	ExecutionTenant(ctx context.Context, executionID string) (ownerTenant string, found bool, err error)
}

// dbExecutionTenantResolver resolves execution ownership from the tenant-scoped
// metadata tables that own execution records: functions.edge_executions for
// edge-function executions and rpc.executions for RPC executions.
type dbExecutionTenantResolver struct {
	db *database.Connection
}

// NewDBExecutionTenantResolver returns an ExecutionTenantResolver backed by the
// execution metadata tables, or nil when db is nil (resolution unavailable —
// tenant-scoped reads then fail closed).
func NewDBExecutionTenantResolver(db *database.Connection) ExecutionTenantResolver {
	if db == nil {
		return nil
	}
	return &dbExecutionTenantResolver{db: db}
}

// ExecutionTenant resolves the owning tenant of an execution. The lookup uses
// the same tenant-aware role as the execution writers: under a tenant context
// the tenant_service role (plus RLS) only exposes rows owned by that tenant,
// while an empty tenant context uses service_role and can see every row.
func (s *dbExecutionTenantResolver) ExecutionTenant(ctx context.Context, executionID string) (string, bool, error) {
	ta := database.TenantAware{DB: s.db}

	var owner *string
	err := ta.WithTenant(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT tenant_id FROM functions.edge_executions WHERE id = $1
			UNION ALL
			SELECT tenant_id FROM rpc.executions WHERE id = $1
			LIMIT 1
		`, executionID).Scan(&owner)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("failed to resolve execution tenant: %w", err)
	}
	if owner == nil {
		// Platform-level record without a tenant owner: never visible to
		// tenant-scoped callers.
		return "", false, nil
	}
	return *owner, true, nil
}
