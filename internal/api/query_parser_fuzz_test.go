package api

import (
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// placeholderRe matches bound-parameter tokens ($1, $2, ...) in compiled SQL.
var placeholderRe = regexp.MustCompile(`\$\d+`)

// quotedIdentRe matches double-quoted identifiers ("col", "ta""ble"), which
// legitimately contain arbitrary column-name text.
var quotedIdentRe = regexp.MustCompile(`"[^"]*(?:""[^"]*)*"`)

// FuzzQueryParser drives the full query-parameter surface (filters, order,
// select, pagination, cursors) with arbitrary input. Oracle: parsing and SQL
// compilation must never panic — every failure has to be an ordinary error.
func FuzzQueryParser(f *testing.F) {
	// Seeds lifted from the table-driven parser tests, plus regression
	// inputs from past hardening fixes (#394).
	f.Add("name", "eq.John", "order", "created_at.desc")
	f.Add("age", "gt.30", "select", "id,name")
	f.Add("status", "in.(queued,running)", "limit", "10")
	f.Add("or", "(name.eq.John,age.gt.30)", "and", "(or(a.lt.1,a.gt.9),b.eq.x)")
	f.Add("name", "eq.'; DROP TABLE users; --", "cursor", "bm90LWJhc2U2NA==")
	f.Add("data->>key", "eq.value", "count", "exact")
	f.Add("col", "cs.{1,2,3}", "offset", "-1")
	f.Add("name", "is.null", "truncate", "abc")
	f.Add("or", "((((((((((((1))))))))))))", "and", "((((((((((((1))))))))))))")

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		values := url.Values{}
		values.Set(k1, v1)
		values.Set(k2, v2)

		qp := NewQueryParser(testConfig())
		params, err := qp.Parse(values)
		if err != nil {
			return
		}
		// A successful parse must also survive SQL compilation.
		_, _, _ = params.ToSQL("fuzz_table")
	})
}

// FuzzParseFilter drives the filter parser directly and enforces the security
// invariant: a hostile filter value may reach the query only as a bound
// parameter — it must never be interpolated into the SQL text.
func FuzzParseFilter(f *testing.F) {
	f.Add("name", "eq", "John")
	f.Add("age", "gt", "30")
	f.Add("status", "in", "(queued,running)")
	f.Add("created_at", "between", "(2024-01-01,2024-12-31)")
	f.Add("name", "is", "null")
	f.Add("name", "eq", "'; DROP TABLE users; --")
	f.Add("name", "eq", `"\"); DROP TABLE users; --`)
	f.Add("data->>key", "eq", "v")
	f.Add("name", "eq", "")
	f.Add("or", "eq", "((((((((((((1))))))))))))")
	f.Add("age", "eq", "1).(or(")

	f.Fuzz(func(t *testing.T, column, op, value string) {
		qp := NewQueryParser(testConfig())
		params := &QueryParams{Filters: []Filter{}, Order: []OrderBy{}}

		if err := qp.parseFilter(column+"."+op, value, params); err != nil {
			return
		}
		sql, _, err := params.ToSQL("fuzz_table")
		if err != nil {
			return
		}

		// Mask everything allowed to carry text (placeholders, quoted
		// identifiers, casts) before looking for the raw value.
		masked := placeholderRe.ReplaceAllString(sql, "?")
		masked = quotedIdentRe.ReplaceAllString(masked, "?")
		masked = strings.ReplaceAll(masked, "::numeric", "")

		// Only values with injection-significant characters are checked, so
		// benign collisions with keywords are not reported. The "is"
		// operator emits whitelisted keywords only.
		if op != "is" && len(value) >= 3 &&
			strings.ContainsAny(value, "\"'`;\\-") &&
			strings.Contains(masked, value) {
			t.Errorf("raw filter value %q leaked into SQL: %s", value, sql)
		}
	})
}

// FuzzDecodeCursor feeds arbitrary base64 to the keyset-pagination cursor
// decoder. Oracle: no panic; a nil error implies a usable cursor.
func FuzzDecodeCursor(f *testing.F) {
	valid, err := EncodeCursor("created_at", "2024-01-01T00:00:00Z", false)
	if err != nil {
		f.Fatal(err)
	}
	f.Add(valid)
	f.Add("")
	f.Add("!!!not base64!!!")
	f.Add("eyJjb2x1bW4iOiIifQ==") // {"column":""}

	f.Fuzz(func(t *testing.T, cursor string) {
		data, err := DecodeCursor(cursor)
		if err != nil {
			return
		}
		if data == nil || data.Column == "" {
			t.Errorf("DecodeCursor succeeded but returned unusable cursor")
		}
	})
}
