package realtime

import (
	"testing"
)

// FuzzParseFilterMatches feeds arbitrary subscription-filter strings through
// parsing and record matching, including JSONB path traversal and wildcard
// pattern compilation. Oracle: no panic on either path.
func FuzzParseFilterMatches(f *testing.F) {
	f.Add("created_by=eq.user123", "created_by", "user123")
	f.Add("priority=gt.5", "priority", "3")
	f.Add("status=in.(queued,running)", "status", "queued")
	f.Add("data->>key=eq.v", "data", "")
	f.Add("name=like.rec*", "name", "record")
	f.Add("n=lte.10", "", "")
	f.Add(")))garbage", "", "")
	f.Add("=", "", "")

	f.Fuzz(func(t *testing.T, filterStr, col, val string) {
		filter, err := ParseFilter(filterStr)
		if err != nil {
			return
		}
		if filterStr == "" && filter != nil {
			t.Errorf("empty filter string must parse to nil, got %+v", filter)
		}

		record := map[string]interface{}{
			"name":   "record",
			"n":      3.14,
			"b":      true,
			"nilkey": nil,
			"data":   map[string]interface{}{"key": "v", "n": 2, "deep": map[string]interface{}{"x": 1}},
		}
		if col != "" {
			record[col] = val
		}

		// Must not panic on any record shape.
		_ = filter.Matches(record)
	})
}
