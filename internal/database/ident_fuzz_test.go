package database

import (
	"strings"
	"testing"
)

// FuzzQuoteIdentifier checks the invariants that make identifier quoting safe
// to interpolate into SQL:
//  1. the result is wrapped in double quotes;
//  2. every double quote inside the body is escaped (doubled) — a lone
//     unescaped quote would terminate the identifier;
//  3. anything ValidIdentifier accepts quotes to itself verbatim.
func FuzzQuoteIdentifier(f *testing.F) {
	f.Add("users")
	f.Add("_private_2024")
	f.Add(`has"quote`)
	f.Add(`has""two`)
	f.Add(`"lead`)
	f.Add("trail\"")
	f.Add("semi;colons")
	f.Add("sp ace")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		q := QuoteIdentifier(s)

		if len(q) < 2 || !strings.HasPrefix(q, `"`) || !strings.HasSuffix(q, `"`) {
			t.Fatalf("QuoteIdentifier(%q) = %q, not wrapped in double quotes", s, q)
		}

		body := q[1 : len(q)-1]
		for i := 0; i < len(body); i++ {
			if body[i] != '"' {
				continue
			}
			if i+1 >= len(body) || body[i+1] != '"' {
				t.Fatalf("QuoteIdentifier(%q) = %q: unescaped quote at byte %d", s, q, i)
			}
			i++ // skip the escape pair
		}

		if ValidIdentifier(s) && q != `"`+s+`"` {
			t.Fatalf("valid identifier %q must quote verbatim, got %q", s, q)
		}
	})
}
