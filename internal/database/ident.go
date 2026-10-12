package database

import (
	"regexp"
	"strings"
)

// MaxIdentifierLen is the maximum length of a PostgreSQL identifier in bytes:
// NAMEDATALEN (64) minus the terminating null byte.
const MaxIdentifierLen = 63

// identifierRegex matches PostgreSQL identifiers that do not require quoting:
// they start with a letter or underscore and contain only letters, digits and
// underscores.
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidIdentifier reports whether s is a valid, unquoted PostgreSQL identifier:
// it starts with a letter or underscore, contains only letters, digits and
// underscores, and is at most MaxIdentifierLen (63) bytes long. Call this
// before interpolating user-supplied names into SQL, or use QuoteIdentifier.
func ValidIdentifier(s string) bool {
	return len(s) <= MaxIdentifierLen && identifierRegex.MatchString(s)
}

// QuoteIdentifier returns s as a safely quoted SQL identifier, escaping any
// embedded double quotes by doubling them. It does not validate s; callers
// dealing with user input should check ValidIdentifier first.
func QuoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
