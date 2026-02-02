package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// extractTableName Tests
// =============================================================================

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		// SELECT queries
		{
			name:     "simple select",
			sql:      "SELECT * FROM users",
			expected: "users",
		},
		{
			name:     "select with columns",
			sql:      "SELECT id, name, email FROM users WHERE active = true",
			expected: "users",
		},
		{
			name:     "select with schema",
			sql:      "SELECT * FROM public.users",
			expected: "public",
		},
		{
			name:     "select lowercase",
			sql:      "select * from products",
			expected: "products",
		},
		{
			name:     "select with quoted table",
			sql:      `SELECT * FROM "users"`,
			expected: "users",
		},
		{
			name:     "select with single quoted table",
			sql:      "SELECT * FROM 'users'",
			expected: "users",
		},

		// INSERT queries
		{
			name:     "simple insert",
			sql:      "INSERT INTO users (name) VALUES ('John')",
			expected: "users",
		},
		{
			name:     "insert with schema",
			sql:      "INSERT INTO auth.users (name) VALUES ('John')",
			expected: "auth",
		},
		{
			name:     "insert lowercase",
			sql:      "insert into products (name) values ('Widget')",
			expected: "products",
		},

		// UPDATE queries
		{
			name:     "simple update",
			sql:      "UPDATE users SET name = 'Jane' WHERE id = 1",
			expected: "users",
		},
		{
			name:     "update with schema",
			sql:      "UPDATE public.users SET name = 'Jane'",
			expected: "public",
		},
		{
			name:     "update lowercase",
			sql:      "update orders set status = 'shipped'",
			expected: "orders",
		},

		// DELETE queries
		{
			name:     "simple delete",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: "users",
		},
		{
			name:     "delete with schema",
			sql:      "DELETE FROM auth.sessions WHERE expired = true",
			expected: "auth",
		},
		{
			name:     "delete lowercase",
			sql:      "delete from temp_data",
			expected: "temp_data",
		},

		// Edge cases
		{
			name:     "unknown statement type",
			sql:      "CREATE TABLE users (id INT)",
			expected: "unknown",
		},
		{
			name:     "truncate statement",
			sql:      "TRUNCATE TABLE users",
			expected: "unknown",
		},
		{
			name:     "empty string",
			sql:      "",
			expected: "unknown",
		},
		{
			name:     "whitespace only",
			sql:      "   ",
			expected: "unknown",
		},
		{
			name:     "select with join",
			sql:      "SELECT u.* FROM users u JOIN orders o ON u.id = o.user_id",
			expected: "users",
		},
		{
			name:     "select with subquery",
			sql:      "SELECT * FROM (SELECT * FROM users) as subq",
			expected: "users", // extractTableName uses simple regex that finds first FROM
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTableName(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTableName_CaseInsensitive(t *testing.T) {
	// All variations should work
	variations := []string{
		"SELECT * FROM users",
		"select * from users",
		"Select * From users",
		"SELECT * FROM USERS",
		"sElEcT * fRoM users",
	}

	for _, sql := range variations {
		result := extractTableName(sql)
		assert.Equal(t, "users", result, "Failed for SQL: %s", sql)
	}
}

// =============================================================================
// extractOperation Tests
// =============================================================================

func TestExtractOperation(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		// SELECT
		{
			name:     "select uppercase",
			sql:      "SELECT * FROM users",
			expected: "select",
		},
		{
			name:     "select lowercase",
			sql:      "select * from users",
			expected: "select",
		},
		{
			name:     "select mixed case",
			sql:      "Select * From users",
			expected: "select",
		},
		{
			name:     "select with leading whitespace",
			sql:      "   SELECT * FROM users",
			expected: "select",
		},

		// INSERT
		{
			name:     "insert uppercase",
			sql:      "INSERT INTO users VALUES (1)",
			expected: "insert",
		},
		{
			name:     "insert lowercase",
			sql:      "insert into users values (1)",
			expected: "insert",
		},

		// UPDATE
		{
			name:     "update uppercase",
			sql:      "UPDATE users SET name = 'John'",
			expected: "update",
		},
		{
			name:     "update lowercase",
			sql:      "update users set name = 'John'",
			expected: "update",
		},

		// DELETE
		{
			name:     "delete uppercase",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: "delete",
		},
		{
			name:     "delete lowercase",
			sql:      "delete from users where id = 1",
			expected: "delete",
		},

		// Other operations
		{
			name:     "create table",
			sql:      "CREATE TABLE users (id INT)",
			expected: "other",
		},
		{
			name:     "drop table",
			sql:      "DROP TABLE users",
			expected: "other",
		},
		{
			name:     "alter table",
			sql:      "ALTER TABLE users ADD COLUMN email TEXT",
			expected: "other",
		},
		{
			name:     "truncate",
			sql:      "TRUNCATE TABLE users",
			expected: "other",
		},
		{
			name:     "begin transaction",
			sql:      "BEGIN",
			expected: "other",
		},
		{
			name:     "commit",
			sql:      "COMMIT",
			expected: "other",
		},
		{
			name:     "rollback",
			sql:      "ROLLBACK",
			expected: "other",
		},
		{
			name:     "set statement",
			sql:      "SET search_path TO public",
			expected: "other",
		},

		// Edge cases
		{
			name:     "empty string",
			sql:      "",
			expected: "other",
		},
		{
			name:     "whitespace only",
			sql:      "   ",
			expected: "other",
		},
		{
			name:     "comment only",
			sql:      "-- this is a comment",
			expected: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOperation(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// truncateQuery Tests
// =============================================================================

func TestTruncateQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		maxLen   int
		expected string
	}{
		{
			name:     "short query under limit",
			query:    "SELECT * FROM users",
			maxLen:   100,
			expected: "SELECT * FROM users",
		},
		{
			name:     "query exactly at limit",
			query:    "SELECT * FROM users",
			maxLen:   19,
			expected: "SELECT * FROM users",
		},
		{
			name:     "query over limit",
			query:    "SELECT * FROM users WHERE active = true",
			maxLen:   20,
			expected: "SELECT * FROM users ... (truncated)",
		},
		{
			name:     "very short limit",
			query:    "SELECT * FROM users",
			maxLen:   5,
			expected: "SELEC... (truncated)",
		},
		{
			name:     "empty query",
			query:    "",
			maxLen:   100,
			expected: "",
		},
		{
			name:     "zero max length",
			query:    "SELECT",
			maxLen:   0,
			expected: "... (truncated)",
		},
		{
			name:     "long query",
			query:    "SELECT id, name, email, created_at, updated_at, status, role, metadata FROM users WHERE active = true AND verified = true ORDER BY created_at DESC LIMIT 100",
			maxLen:   50,
			expected: "SELECT id, name, email, created_at, updated_at, st... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateQuery(tt.query, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateQuery_Length(t *testing.T) {
	query := "SELECT * FROM users WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10)"
	maxLen := 30

	result := truncateQuery(query, maxLen)

	// Result should contain the truncated marker
	assert.Contains(t, result, "... (truncated)")
	// The prefix should be exactly maxLen characters
	prefix := result[:maxLen]
	assert.Len(t, prefix, maxLen)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkExtractTableName_SELECT(b *testing.B) {
	sql := "SELECT id, name, email FROM users WHERE active = true ORDER BY created_at"
	for i := 0; i < b.N; i++ {
		_ = extractTableName(sql)
	}
}

func BenchmarkExtractTableName_INSERT(b *testing.B) {
	sql := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"
	for i := 0; i < b.N; i++ {
		_ = extractTableName(sql)
	}
}

func BenchmarkExtractTableName_UPDATE(b *testing.B) {
	sql := "UPDATE users SET name = 'Jane', email = 'jane@example.com' WHERE id = 123"
	for i := 0; i < b.N; i++ {
		_ = extractTableName(sql)
	}
}

func BenchmarkExtractTableName_DELETE(b *testing.B) {
	sql := "DELETE FROM users WHERE id = 123 AND active = false"
	for i := 0; i < b.N; i++ {
		_ = extractTableName(sql)
	}
}

func BenchmarkExtractOperation(b *testing.B) {
	sql := "SELECT * FROM users WHERE active = true"
	for i := 0; i < b.N; i++ {
		_ = extractOperation(sql)
	}
}

func BenchmarkTruncateQuery_Short(b *testing.B) {
	query := "SELECT * FROM users"
	for i := 0; i < b.N; i++ {
		_ = truncateQuery(query, 200)
	}
}

func BenchmarkTruncateQuery_Long(b *testing.B) {
	query := "SELECT id, name, email, phone, address, city, state, zip, country, created_at, updated_at FROM users WHERE active = true AND verified = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 100 OFFSET 0"
	for i := 0; i < b.N; i++ {
		_ = truncateQuery(query, 100)
	}
}

// =============================================================================
// quoteIdentifier Tests
// =============================================================================

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "simple table name",
			identifier: "users",
			expected:   `"users"`,
		},
		{
			name:       "table name with underscore",
			identifier: "user_profiles",
			expected:   `"user_profiles"`,
		},
		{
			name:       "schema qualified name",
			identifier: "public.users",
			expected:   `"public.users"`,
		},
		{
			name:       "identifier with embedded quote",
			identifier: `my"table`,
			expected:   `"my""table"`,
		},
		{
			name:       "identifier with multiple quotes",
			identifier: `test"with"quotes`,
			expected:   `"test""with""quotes"`,
		},
		{
			name:       "empty identifier",
			identifier: "",
			expected:   `""`,
		},
		{
			name:       "identifier with spaces",
			identifier: "my table",
			expected:   `"my table"`,
		},
		{
			name:       "reserved keyword",
			identifier: "select",
			expected:   `"select"`,
		},
		{
			name:       "mixed case identifier",
			identifier: "MyTable",
			expected:   `"MyTable"`,
		},
		{
			name:       "identifier with special characters",
			identifier: "user@data",
			expected:   `"user@data"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.identifier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteIdentifier_SQLInjectionPrevention(t *testing.T) {
	t.Run("prevents basic injection", func(t *testing.T) {
		// Attempt SQL injection via identifier
		malicious := `users"; DROP TABLE users; --`
		result := quoteIdentifier(malicious)

		// The result should be safely quoted, making injection impossible
		assert.Equal(t, `"users""; DROP TABLE users; --"`, result)
		assert.Contains(t, result, `""`)
	})

	t.Run("handles multiple injection attempts", func(t *testing.T) {
		malicious := `"";--`
		result := quoteIdentifier(malicious)

		// Embedded quotes should be doubled
		assert.Equal(t, `"""""";--"`, result)
	})
}

func BenchmarkQuoteIdentifier_Simple(b *testing.B) {
	identifier := "users"
	for i := 0; i < b.N; i++ {
		_ = quoteIdentifier(identifier)
	}
}

func BenchmarkQuoteIdentifier_WithQuotes(b *testing.B) {
	identifier := `table"with"quotes`
	for i := 0; i < b.N; i++ {
		_ = quoteIdentifier(identifier)
	}
}
