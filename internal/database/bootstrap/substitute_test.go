package bootstrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteAppUser(t *testing.T) {
	t.Run("valid identifier postgres", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "postgres")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO postgres", result)
	})

	t.Run("valid identifier my_user", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my_user")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO my_user", result)
	})

	t.Run("valid identifier _underscore", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "_underscore")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO _underscore", result)
	})

	t.Run("empty appUser returns sql unchanged", func(t *testing.T) {
		sql := "GRANT ALL TO {{APP_USER}}"
		result, err := SubstituteAppUser(sql, "")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("no placeholder returns sql unchanged", func(t *testing.T) {
		sql := "SELECT 1"
		result, err := SubstituteAppUser(sql, "postgres")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("rejected hyphen in identifier", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my-user")
		assert.Error(t, err)
	})

	t.Run("rejected semicolon injection", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "user; DROP TABLE")
		assert.Error(t, err)
	})

	t.Run("rejected quote in identifier", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "user'name")
		assert.Error(t, err)
	})

	t.Run("rejected starts with digit", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "1user")
		assert.Error(t, err)
	})

	t.Run("rejected contains space", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my user")
		assert.Error(t, err)
	})

	t.Run("empty string with placeholder returns unchanged", func(t *testing.T) {
		sql := "GRANT ALL TO {{APP_USER}}"
		result, err := SubstituteAppUser(sql, "")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})
}

func TestSubstituteAppUser_MultiplePlaceholders(t *testing.T) {
	sql := "GRANT anon TO {{APP_USER}}; ALTER DEFAULT PRIVILEGES GRANT ALL ON TABLES TO {{APP_USER}};"
	result, err := SubstituteAppUser(sql, "fluxbase_app")
	require.NoError(t, err)
	assert.Equal(t, "GRANT anon TO fluxbase_app; ALTER DEFAULT PRIVILEGES GRANT ALL ON TABLES TO fluxbase_app;", result)
	assert.NotContains(t, result, AppUserPlaceholder)
}

func TestSubstituteAppUser_LongIdentifier(t *testing.T) {
	// PostgreSQL allows identifiers up to 63 bytes; SubstituteAppUser enforces
	// that limit up front via database.ValidIdentifier.
	long := strings.Repeat("a", 63)
	result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", long)
	require.NoError(t, err)
	assert.Equal(t, "GRANT ALL TO "+long, result)

	// 64+ chars exceed MaxIdentifierLen and are rejected before reaching the
	// database (defense in depth — PostgreSQL would reject them anyway).
	long64 := strings.Repeat("a", 64)
	_, err = SubstituteAppUser("GRANT ALL TO {{APP_USER}}", long64)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database username")
}

func TestSubstituteAppUser_RejectsInjectionVectors(t *testing.T) {
	cases := []struct {
		name    string
		appUser string
	}{
		{"dollar-quoted string", "$$; DROP SCHEMA auth CASCADE; $$"},
		{"double-dash comment", "app-- stolen"},
		{"newline escape", "app\nGRANT ALL ON SCHEMA public TO attacker"},
		{"parenthesis", "app) TO attacker; --"},
		{"quoted identifier with spaces", "\"app user\""},
		{"backslash", "app\\"},
		{"unicode homoglyph dash", "app\u2013user"},
		{"percent wildcard", "app%"},
		{"empty after trim", " "},
	}
	sql := "GRANT ALL TO {{APP_USER}}"
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SubstituteAppUser(sql, tc.appUser)
			assert.Error(t, err, "identifier %q must be rejected", tc.appUser)
			assert.Empty(t, result)
		})
	}
}

func TestSubstituteAppUser_EmptySQLWithUser(t *testing.T) {
	// An empty SQL string containing no placeholder is returned unchanged.
	result, err := SubstituteAppUser("", "postgres")
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestAppUserPlaceholderConstant(t *testing.T) {
	// The placeholder is part of the SQL-authoring contract (bootstrap.sql and
	// the per-schema files embed it); silently changing it would break
	// provisioning everywhere.
	assert.Equal(t, "{{APP_USER}}", AppUserPlaceholder)
}
