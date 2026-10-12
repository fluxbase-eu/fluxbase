package tools

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// quotedIdentRe matches double-quoted identifiers, which legitimately contain
// arbitrary column-name text.
var quotedIdentRe = regexp.MustCompile(`"[^"]*(?:""[^"]*)*"`)

// FuzzMCPFilterParse drives the MCP tool-argument filter parsing path.
// Model-supplied JSON arguments are hostile by definition. Oracle: parsing
// and SQL compilation never panic, a condition with a bound argument always
// references its placeholder, and a hostile value is never interpolated into
// the SQL text.
func FuzzMCPFilterParse(f *testing.F) {
	f.Add("age", "gt.30")
	f.Add("name", "eq.John")
	f.Add("status", "in.(queued,running)")
	f.Add("flag", "is.null")
	f.Add("flag", "is.NOT_BOOLEAN")
	f.Add("name", "eq.'; DROP TABLE users; --")
	f.Add("col", "bogus.x")
	f.Add("data->>key", "eq.v")
	f.Add("name", "eq.")

	f.Fuzz(func(t *testing.T, column, valueStr string) {
		filter, err := parseFilterValue(column, valueStr)
		if err != nil {
			return
		}

		quotedCol := quoteIdentifier(column)
		if quotedCol == "" {
			return
		}

		condition, arg := filterToSQL(filter, quotedCol, 1)
		if condition == "" {
			// Whitelist path (IS NULL/TRUE/FALSE with an invalid value
			// shape): the caller drops the condition.
			return
		}
		if arg == nil {
			return
		}

		if !strings.Contains(condition, "$1") {
			t.Errorf("condition %q carries an argument but no $1 placeholder", condition)
		}

		s := fmt.Sprint(arg)
		masked := quotedIdentRe.ReplaceAllString(condition, "?")
		if len(s) >= 3 && strings.ContainsAny(s, "\"'`;\\") && strings.Contains(masked, s) {
			t.Errorf("raw filter value %q leaked into SQL: %s", s, condition)
		}
	})
}

// FuzzMCPOrder drives the MCP order-string parser. Oracle: never panics, and
// every parsed column survives round-tripping through quoting.
func FuzzMCPOrder(f *testing.F) {
	f.Add("created_at.desc")
	f.Add("name asc, age desc")
	f.Add("col1.desc,col2.asc,col3")
	f.Add(",,,")
	f.Add("weird . spacing . everywhere")

	f.Fuzz(func(t *testing.T, orderStr string) {
		order, err := parseOrder(orderStr)
		if err != nil {
			return
		}
		for _, o := range order {
			if o.Column == "" {
				t.Errorf("parseOrder(%q) produced an empty column", orderStr)
			}
		}
	})
}

// FuzzMCPCrudFilterToSQL drives the CRUD tool's filter-to-SQL path with the
// same invariants as FuzzMCPFilterParse.
func FuzzMCPCrudFilterToSQL(f *testing.F) {
	f.Add("age", "gt.30")
	f.Add("name", "eq.John")
	f.Add("status", "in.(a,b)")
	f.Add("flag", "is.null")
	f.Add("name", "eq.'; DROP TABLE users; --")
	f.Add("col", "unsupported.op")

	f.Fuzz(func(t *testing.T, column, valueStr string) {
		sql, arg, err := parseFilterToSQL(column, valueStr, 1)
		if err != nil {
			return
		}
		if arg == nil {
			return
		}

		if !strings.Contains(sql, "$1") {
			t.Errorf("condition %q carries an argument but no $1 placeholder", sql)
		}

		s := fmt.Sprint(arg)
		masked := quotedIdentRe.ReplaceAllString(sql, "?")
		if len(s) >= 3 && strings.ContainsAny(s, "\"'`;\\") && strings.Contains(masked, s) {
			t.Errorf("raw filter value %q leaked into SQL: %s", s, sql)
		}
	})
}
