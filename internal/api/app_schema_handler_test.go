package api

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// =============================================================================
// SEC-07 (#372) regression guards: raw database error text must never reach
// API clients through the admin schema/auth handler error paths.
// =============================================================================

// TestSchemaHandlerErrorSanitization_NoRawSQL verifies that a pgx-style error
// surfaced through database.SanitizeErrorMessage (what the guarded handlers
// now interpolate) carries no raw SQL text or server message.
func TestSchemaHandlerErrorSanitization_NoRawSQL(t *testing.T) {
	rawSQL := `SELECT api_key FROM platform.service_keys WHERE key_hash = 'supersecret'`
	pgErr := &pgconn.PgError{
		Code:    "42601",
		Message: `syntax error at or near "` + rawSQL + `"`,
	}
	err := fmt.Errorf("query failed: %w", pgErr)

	msg := database.SanitizeErrorMessage(err)

	assert.NotContains(t, msg, "platform.service_keys")
	assert.NotContains(t, msg, "api_key")
	assert.NotContains(t, msg, "supersecret")
	assert.NotContains(t, msg, "SELECT")
	assert.Contains(t, msg, "42601", "the SQLSTATE code is preserved for debugging")
}

// TestSchemaHandlerSendInternalError_NoRawErrInterpolation is a static guard
// over the admin handler sources: every SendInternalError call that formats an
// error must interpolate database.SanitizeErrorMessage(err), never the raw err
// (SEC-07, #372).
func TestSchemaHandlerSendInternalError_NoRawErrInterpolation(t *testing.T) {
	files := []string{
		"admin_auth_handler.go",
		"app_schema_handler.go",
		"ddl_handler_column.go",
		"internal_schema_handler.go",
	}

	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(name)
			require.NoError(t, err)

			for i, line := range strings.Split(string(src), "\n") {
				if !strings.Contains(line, "SendInternalError(c, fmt.Sprintf(") {
					continue
				}
				assert.Contains(t, line, "database.SanitizeErrorMessage(err)",
					"%s:%d interpolates a raw error into the client response: %s",
					name, i+1, strings.TrimSpace(line))
			}
		})
	}
}
