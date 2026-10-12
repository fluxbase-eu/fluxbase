package database

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIdentifier(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{"simple lowercase", "users", true},
		{"uppercase", "Users", true},
		{"with underscore", "user_accounts", true},
		{"leading underscore", "_internal", true},
		{"with digits", "table1", true},
		{"digits after letter", "a1b2c3", true},
		{"single letter", "a", true},
		{"single underscore", "_", true},
		{"empty string", "", false},
		{"leading digit", "1table", false},
		{"leading hyphen", "-table", false},
		{"embedded hyphen", "my-table", false},
		{"embedded space", "my table", false},
		{"embedded quote", `my"table`, false},
		{"embedded semicolon", "table; DROP TABLE users", false},
		{"embedded single quote", "table'name", false},
		{"dot separated", "schema.table", false},
		{"parenthesis", "table(", false},
		{"newline", "table\n", false},
		{"sql comment", "table--", false},
		{"exactly 63 bytes", strings.Repeat("a", 63), true},
		{"64 bytes exceeds PG limit", strings.Repeat("a", 64), false},
		{"far over limit", strings.Repeat("a", 300), false},
		{"63 bytes with underscore mix", strings.Repeat("_a1", 21), true},
		{"unicode letter rejected", "tablé", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, ValidIdentifier(tt.input), "input %q (len %d)", tt.input, len(tt.input))
		})
	}
}

// TestQuoteIdentifier table coverage (including embedded-quote escaping and
// injection attempts) lives in connection_test.go, retargeted at the shared
// QuoteIdentifier. The round-trip property below guards escaping correctness.

// TestQuoteIdentifier_RoundTrip verifies that quoting neutralizes double
// quotes: un-escaping the quoted identifier yields the original string, so a
// malicious value cannot terminate the quoted region early.
func TestQuoteIdentifier_RoundTrip(t *testing.T) {
	inputs := []string{
		`x"; DROP TABLE users; --`,
		`a""b`,
		`"wrapped"`,
		`pct%_weird`,
	}

	for _, in := range inputs {
		quoted := QuoteIdentifier(in)
		// Exact inverse of QuoteIdentifier's escaping.
		unquoted := strings.ReplaceAll(quoted[1:len(quoted)-1], `""`, `"`)
		assert.Equal(t, in, unquoted, "input %q", in)
		// The quoted form must start and end with exactly one unescaped quote.
		assert.True(t, strings.HasPrefix(quoted, `"`), "input %q", in)
		assert.True(t, strings.HasSuffix(quoted, `"`), "input %q", in)
	}
}
