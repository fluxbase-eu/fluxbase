package api

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nestedAndQuery builds an and= filter nested n levels deep:
// and=(and(...(col.eq.1)...)) with n-1 nested and( wrappers.
func nestedAndQuery(n int) string {
	if n < 1 {
		n = 1
	}
	return "and=(" + strings.Repeat("and(", n-1) + "col.eq.1" + strings.Repeat(")", n-1) + ")"
}

// TestParseFilter_DepthLimit verifies nested and()/or() expressions are
// rejected beyond MaxFilterDepth and accepted at the boundary.
func TestParseFilter_DepthLimit(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("nesting at MaxFilterDepth is accepted", func(t *testing.T) {
		values, err := url.ParseQuery(nestedAndQuery(MaxFilterDepth))
		require.NoError(t, err)

		params, err := parser.Parse(values)
		require.NoError(t, err)
		require.Len(t, params.Filters, 1)
		assert.Equal(t, "col", params.Filters[0].Column)
	})

	t.Run("nesting beyond MaxFilterDepth is rejected", func(t *testing.T) {
		values, err := url.ParseQuery(nestedAndQuery(MaxFilterDepth + 1))
		require.NoError(t, err)

		_, err = parser.Parse(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("maximum nesting depth of %d", MaxFilterDepth))
	})

	t.Run("error mentions the configured depth", func(t *testing.T) {
		values, err := url.ParseQuery(nestedAndQuery(1000))
		require.NoError(t, err)

		_, err = parser.Parse(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nesting depth")
	})
}

// TestParseFilter_CountLimit verifies the MaxFilterCount cap.
func TestParseFilter_CountLimit(t *testing.T) {
	parser := NewQueryParser(testConfig())

	buildValues := func(n int) url.Values {
		vals := url.Values{}
		for i := 0; i < n; i++ {
			vals.Set(fmt.Sprintf("col%d.eq", i), fmt.Sprintf("val%d", i))
		}
		return vals
	}

	t.Run("MaxFilterCount filters are accepted", func(t *testing.T) {
		params, err := parser.Parse(buildValues(MaxFilterCount))
		require.NoError(t, err)
		assert.Len(t, params.Filters, MaxFilterCount)
	})

	t.Run("more than MaxFilterCount filters are rejected", func(t *testing.T) {
		_, err := parser.Parse(buildValues(MaxFilterCount + 1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("maximum of %d filter", MaxFilterCount))
	})

	t.Run("count cap applies inside logical groups", func(t *testing.T) {
		filters := make([]string, 0, MaxFilterCount+1)
		for i := 0; i < MaxFilterCount+1; i++ {
			filters = append(filters, fmt.Sprintf("col%d.eq.v%d", i, i))
		}
		values, err := url.ParseQuery("or=(" + strings.Join(filters, ",") + ")")
		require.NoError(t, err)

		_, err = parser.Parse(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many filters")
	})

	t.Run("count cap applies inside nested OR groups", func(t *testing.T) {
		filters := make([]string, 0, MaxFilterCount+1)
		for i := 0; i < MaxFilterCount+1; i++ {
			filters = append(filters, fmt.Sprintf("col%d.eq.v%d", i, i))
		}
		values, err := url.ParseQuery("and=(or(" + strings.Join(filters, ",") + "))")
		require.NoError(t, err)

		_, err = parser.Parse(values)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too many filters")
	})
}

// TestParseFilterValue_Parity proves the shared parseFilterValue helper parses
// identically across all four call sites that previously held copy-pasted
// switches: classic format (column.op=value), PostgREST format
// (column=op.value), logical groups (or=(col.op.value,...)) and nested OR
// groups (or=(or(col.op.value,...))).
func TestParseFilterValue_Parity(t *testing.T) {
	parser := NewQueryParser(testConfig())

	type parsed struct {
		value    interface{}
		isOr     bool
		groupSet bool
	}

	// collect filters from all four syntactic forms for one operator/value pair.
	parseAllForms := func(t *testing.T, opExpr, raw string) map[string]parsed {
		t.Helper()

		queries := map[string]string{
			"classic":   fmt.Sprintf("col.%s=%s", opExpr, url.QueryEscape(raw)),
			"postgrest": fmt.Sprintf("col=%s", url.QueryEscape(opExpr+"."+raw)),
			"logical":   fmt.Sprintf("or=(%s)", url.QueryEscape("col."+opExpr+"."+raw)),
			"nested-or": fmt.Sprintf("and=(or(%s))", url.QueryEscape("col."+opExpr+"."+raw)),
		}

		out := make(map[string]parsed, len(queries))
		for form, q := range queries {
			values, err := url.ParseQuery(q)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			require.NoError(t, err, "form %s: op %s raw %q", form, opExpr, raw)
			require.Len(t, params.Filters, 1, "form %s: op %s raw %q", form, opExpr, raw)

			f := params.Filters[0]
			assert.Equal(t, "col", f.Column, "form %s", form)
			out[form] = parsed{value: f.Value, isOr: f.IsOr, groupSet: f.OrGroupID != 0}
		}
		return out
	}

	cases := []struct {
		name   string
		opExpr string
		raw    string
		expect interface{}
	}{
		{"in array parens", "in", "(1,2,3)", []string{"1", "2", "3"}},
		// Note: bracket arrays containing commas (e.g. ["a","b"]) are not valid
		// inside or=(...) groups — the logical-group splitter only respects
		// parentheses — so parity for that input is out of scope here.
		{"in single-element brackets", "in", `["a"]`, []string{"a"}},
		{"between two bounds", "between", "(1,10)", []interface{}{"1", "10"}},
		{"is null", "is", "null", nil},
		{"is true", "is", "true", true},
		{"is false", "is", "false", false},
		{"plain eq value", "eq", "some-value", "some-value"},
		{"value with dots", "eq", "1.2.3", "1.2.3"},
		{"gte timestamp", "gte", "2025-01-01T00:00:00Z", "2025-01-01T00:00:00Z"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAllForms(t, tc.opExpr, tc.raw)

			// All four forms must produce the identical parsed value.
			for form, p := range got {
				assert.Equal(t, tc.expect, p.value, "form %s: value mismatch", form)
			}

			// Structural expectations per form.
			assert.False(t, got["classic"].isOr)
			assert.False(t, got["postgrest"].isOr)
			assert.True(t, got["logical"].isOr)
			assert.True(t, got["nested-or"].isOr)
			assert.False(t, got["logical"].groupSet)
			assert.True(t, got["nested-or"].groupSet)
		})
	}
}

// TestParseFilterValue_ParityErrors verifies operator-specific value errors
// surface identically from all four parsing paths.
func TestParseFilterValue_ParityErrors(t *testing.T) {
	parser := NewQueryParser(testConfig())

	queries := []string{
		"col.is=" + url.QueryEscape("maybe"),                // classic
		"col=" + url.QueryEscape("is.maybe"),                // PostgREST
		"or=(col.is." + url.QueryEscape("maybe") + ")",      // logical group
		"and=(or(col.is." + url.QueryEscape("maybe") + "))", // nested OR group
	}

	// The core error produced by parseFilterValue; the wrapping prefix
	// ("invalid filter parameter <key>:") differs per syntactic path, but the
	// underlying error must be identical everywhere.
	const coreErr = "invalid value for OpIs operator: maybe (must be null, true, or false)"

	for i, q := range queries {
		values, err := url.ParseQuery(q)
		require.NoError(t, err)

		_, err = parser.Parse(values)
		require.Error(t, err, "query %d should be rejected", i)
		assert.Contains(t, err.Error(), coreErr, "query %d", i)
	}
}
