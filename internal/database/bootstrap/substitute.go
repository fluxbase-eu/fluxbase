package bootstrap

import (
	"fmt"
	"strings"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// AppUserPlaceholder is the placeholder string used in SQL files for the
// runtime database user. It is replaced at execution time with the actual
// configured user (cfg.Database.User).
const AppUserPlaceholder = "{{APP_USER}}"

// SubstituteAppUser replaces all {{APP_USER}} placeholders in the SQL string
// with the provided appUser role name. Returns an error if the username is
// not a valid PostgreSQL identifier to prevent SQL injection via bare-string
// substitution into GRANT and ALTER DEFAULT PRIVILEGES statements.
func SubstituteAppUser(sql string, appUser string) (string, error) {
	if appUser == "" || !strings.Contains(sql, AppUserPlaceholder) {
		return sql, nil
	}
	if !database.ValidIdentifier(appUser) {
		return "", fmt.Errorf("invalid database username %q: must start with a letter or underscore, contain only letters, digits and underscores, and be at most %d characters", appUser, database.MaxIdentifierLen)
	}
	return strings.ReplaceAll(sql, AppUserPlaceholder, appUser), nil
}
