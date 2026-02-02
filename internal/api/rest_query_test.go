package api

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// PostQueryRequest Struct Tests
// =============================================================================

func TestPostQueryRequest_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		limit := 50
		offset := 10

		req := PostQueryRequest{
			Select: "id,name,email",
			Filters: []PostQueryFilter{
				{Column: "status", Operator: "eq", Value: "active"},
			},
			OrFilters:  []string{"type.eq.admin,type.eq.superuser"},
			AndFilters: []string{"active.eq.true"},
			BetweenFilters: []PostQueryBetweenFilter{
				{Column: "age", Min: 18, Max: 65, Negated: false},
			},
			Order: []PostQueryOrderBy{
				{Column: "created_at", Direction: "desc", Nulls: "last"},
			},
			Limit:   &limit,
			Offset:  &offset,
			Count:   "exact",
			GroupBy: []string{"category"},
		}

		assert.Equal(t, "id,name,email", req.Select)
		assert.Len(t, req.Filters, 1)
		assert.Len(t, req.OrFilters, 1)
		assert.Len(t, req.AndFilters, 1)
		assert.Len(t, req.BetweenFilters, 1)
		assert.Len(t, req.Order, 1)
		assert.Equal(t, &limit, req.Limit)
		assert.Equal(t, &offset, req.Offset)
		assert.Equal(t, "exact", req.Count)
		assert.Equal(t, []string{"category"}, req.GroupBy)
	})

	t.Run("handles empty/nil fields", func(t *testing.T) {
		req := PostQueryRequest{}

		assert.Empty(t, req.Select)
		assert.Nil(t, req.Filters)
		assert.Nil(t, req.OrFilters)
		assert.Nil(t, req.AndFilters)
		assert.Nil(t, req.BetweenFilters)
		assert.Nil(t, req.Order)
		assert.Nil(t, req.Limit)
		assert.Nil(t, req.Offset)
		assert.Empty(t, req.Count)
		assert.Nil(t, req.GroupBy)
	})
}

// =============================================================================
// PostQueryFilter Struct Tests
// =============================================================================

func TestPostQueryFilter_Struct(t *testing.T) {
	t.Run("stores column, operator, and value", func(t *testing.T) {
		filter := PostQueryFilter{
			Column:   "status",
			Operator: "eq",
			Value:    "active",
		}

		assert.Equal(t, "status", filter.Column)
		assert.Equal(t, "eq", filter.Operator)
		assert.Equal(t, "active", filter.Value)
	})

	t.Run("handles nil value", func(t *testing.T) {
		filter := PostQueryFilter{
			Column:   "deleted_at",
			Operator: "is",
			Value:    nil,
		}

		assert.Nil(t, filter.Value)
	})

	t.Run("handles numeric value", func(t *testing.T) {
		filter := PostQueryFilter{
			Column:   "count",
			Operator: "gt",
			Value:    100,
		}

		assert.Equal(t, 100, filter.Value)
	})

	t.Run("handles slice value", func(t *testing.T) {
		filter := PostQueryFilter{
			Column:   "status",
			Operator: "in",
			Value:    []string{"active", "pending"},
		}

		assert.IsType(t, []string{}, filter.Value)
	})
}

// =============================================================================
// PostQueryBetweenFilter Struct Tests
// =============================================================================

func TestPostQueryBetweenFilter_Struct(t *testing.T) {
	t.Run("stores between filter fields", func(t *testing.T) {
		filter := PostQueryBetweenFilter{
			Column:  "age",
			Min:     18,
			Max:     65,
			Negated: false,
		}

		assert.Equal(t, "age", filter.Column)
		assert.Equal(t, 18, filter.Min)
		assert.Equal(t, 65, filter.Max)
		assert.False(t, filter.Negated)
	})

	t.Run("handles negated between filter", func(t *testing.T) {
		filter := PostQueryBetweenFilter{
			Column:  "price",
			Min:     0,
			Max:     10,
			Negated: true, // NOT BETWEEN
		}

		assert.True(t, filter.Negated)
	})

	t.Run("handles date values", func(t *testing.T) {
		filter := PostQueryBetweenFilter{
			Column:  "created_at",
			Min:     "2026-01-01",
			Max:     "2026-12-31",
			Negated: false,
		}

		assert.Equal(t, "2026-01-01", filter.Min)
		assert.Equal(t, "2026-12-31", filter.Max)
	})
}

// =============================================================================
// PostQueryOrderBy Struct Tests
// =============================================================================

func TestPostQueryOrderBy_Struct(t *testing.T) {
	t.Run("stores order by fields", func(t *testing.T) {
		order := PostQueryOrderBy{
			Column:    "created_at",
			Direction: "desc",
			Nulls:     "last",
		}

		assert.Equal(t, "created_at", order.Column)
		assert.Equal(t, "desc", order.Direction)
		assert.Equal(t, "last", order.Nulls)
	})

	t.Run("handles ascending order", func(t *testing.T) {
		order := PostQueryOrderBy{
			Column:    "name",
			Direction: "asc",
		}

		assert.Equal(t, "asc", order.Direction)
		assert.Empty(t, order.Nulls)
	})

	t.Run("handles nulls first", func(t *testing.T) {
		order := PostQueryOrderBy{
			Column:    "priority",
			Direction: "asc",
			Nulls:     "first",
		}

		assert.Equal(t, "first", order.Nulls)
	})
}

// =============================================================================
// Select Parsing Tests
// =============================================================================

func TestSelectParsing(t *testing.T) {
	t.Run("parses comma-separated columns", func(t *testing.T) {
		selectStr := "id,name,email"

		columns := strings.Split(selectStr, ",")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}

		assert.Equal(t, []string{"id", "name", "email"}, columns)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		selectStr := " id , name , email "

		columns := strings.Split(selectStr, ",")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}

		assert.Equal(t, []string{"id", "name", "email"}, columns)
	})

	t.Run("handles single column", func(t *testing.T) {
		selectStr := "id"

		columns := strings.Split(selectStr, ",")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}

		assert.Equal(t, []string{"id"}, columns)
	})
}

// =============================================================================
// OR Filter Parsing Tests
// =============================================================================

func TestOrFilterParsing(t *testing.T) {
	t.Run("parses valid OR filter format", func(t *testing.T) {
		orFilter := "status.eq.active,status.eq.pending"

		parts := strings.Split(orFilter, ",")
		var filters []Filter

		for _, part := range parts {
			filterParts := strings.SplitN(strings.TrimSpace(part), ".", 3)
			if len(filterParts) == 3 {
				filters = append(filters, Filter{
					Column:   filterParts[0],
					Operator: FilterOperator(filterParts[1]),
					Value:    filterParts[2],
					IsOr:     true,
				})
			}
		}

		require.Len(t, filters, 2)
		assert.Equal(t, "status", filters[0].Column)
		assert.Equal(t, OpEqual, filters[0].Operator)
		assert.Equal(t, "active", filters[0].Value)
		assert.True(t, filters[0].IsOr)

		assert.Equal(t, "status", filters[1].Column)
		assert.Equal(t, OpEqual, filters[1].Operator)
		assert.Equal(t, "pending", filters[1].Value)
	})

	t.Run("detects invalid OR filter format", func(t *testing.T) {
		invalidFilters := []string{
			"status.eq",           // Missing value
			"status",              // Missing operator and value
			"status.eq.val.extra", // Too many parts is OK (uses first 3)
		}

		for _, filter := range invalidFilters[:2] {
			parts := strings.SplitN(strings.TrimSpace(filter), ".", 3)
			isValid := len(parts) == 3
			assert.False(t, isValid, "Expected filter to be invalid: %s", filter)
		}

		// Check that extra parts are handled
		parts := strings.SplitN("status.eq.val.extra", ".", 3)
		assert.Len(t, parts, 3)
		assert.Equal(t, "val.extra", parts[2]) // Extra parts become part of value
	})
}

// =============================================================================
// AND Filter Parsing Tests
// =============================================================================

func TestAndFilterParsing(t *testing.T) {
	t.Run("parses valid AND filter format", func(t *testing.T) {
		andFilter := "active.eq.true,verified.eq.true"

		parts := strings.Split(andFilter, ",")
		var filters []Filter

		for _, part := range parts {
			filterParts := strings.SplitN(strings.TrimSpace(part), ".", 3)
			if len(filterParts) == 3 {
				filters = append(filters, Filter{
					Column:   filterParts[0],
					Operator: FilterOperator(filterParts[1]),
					Value:    filterParts[2],
					IsOr:     false,
				})
			}
		}

		require.Len(t, filters, 2)
		assert.Equal(t, "active", filters[0].Column)
		assert.False(t, filters[0].IsOr)

		assert.Equal(t, "verified", filters[1].Column)
		assert.False(t, filters[1].IsOr)
	})
}

// =============================================================================
// Between Filter Conversion Tests
// =============================================================================

func TestBetweenFilterConversion(t *testing.T) {
	t.Run("between filter creates two AND filters", func(t *testing.T) {
		bf := PostQueryBetweenFilter{
			Column:  "age",
			Min:     18,
			Max:     65,
			Negated: false,
		}

		// Between: (column >= min AND column <= max)
		var filters []Filter

		if !bf.Negated {
			filters = append(filters, Filter{
				Column:   bf.Column,
				Operator: OpGreaterOrEqual,
				Value:    bf.Min,
				IsOr:     false,
			})
			filters = append(filters, Filter{
				Column:   bf.Column,
				Operator: OpLessOrEqual,
				Value:    bf.Max,
				IsOr:     false,
			})
		}

		require.Len(t, filters, 2)
		assert.Equal(t, OpGreaterOrEqual, filters[0].Operator)
		assert.Equal(t, 18, filters[0].Value)
		assert.Equal(t, OpLessOrEqual, filters[1].Operator)
		assert.Equal(t, 65, filters[1].Value)
	})

	t.Run("negated between filter creates two OR filters", func(t *testing.T) {
		bf := PostQueryBetweenFilter{
			Column:  "age",
			Min:     18,
			Max:     65,
			Negated: true,
		}

		// Not between: (column < min OR column > max)
		var filters []Filter
		groupID := 1

		if bf.Negated {
			filters = append(filters, Filter{
				Column:    bf.Column,
				Operator:  OpLessThan,
				Value:     bf.Min,
				IsOr:      true,
				OrGroupID: groupID,
			})
			filters = append(filters, Filter{
				Column:    bf.Column,
				Operator:  OpGreaterThan,
				Value:     bf.Max,
				IsOr:      true,
				OrGroupID: groupID,
			})
		}

		require.Len(t, filters, 2)
		assert.Equal(t, OpLessThan, filters[0].Operator)
		assert.Equal(t, 18, filters[0].Value)
		assert.True(t, filters[0].IsOr)
		assert.Equal(t, 1, filters[0].OrGroupID)

		assert.Equal(t, OpGreaterThan, filters[1].Operator)
		assert.Equal(t, 65, filters[1].Value)
		assert.True(t, filters[1].IsOr)
		assert.Equal(t, 1, filters[1].OrGroupID)
	})
}

// =============================================================================
// Order Conversion Tests
// =============================================================================

func TestOrderConversion(t *testing.T) {
	t.Run("converts desc direction", func(t *testing.T) {
		o := PostQueryOrderBy{
			Column:    "created_at",
			Direction: "desc",
			Nulls:     "last",
		}

		orderBy := OrderBy{
			Column: o.Column,
			Desc:   strings.ToLower(o.Direction) == "desc",
			Nulls:  strings.ToLower(o.Nulls),
		}

		assert.Equal(t, "created_at", orderBy.Column)
		assert.True(t, orderBy.Desc)
		assert.Equal(t, "last", orderBy.Nulls)
	})

	t.Run("converts asc direction", func(t *testing.T) {
		o := PostQueryOrderBy{
			Column:    "name",
			Direction: "asc",
		}

		orderBy := OrderBy{
			Column: o.Column,
			Desc:   strings.ToLower(o.Direction) == "desc",
			Nulls:  strings.ToLower(o.Nulls),
		}

		assert.False(t, orderBy.Desc)
	})

	t.Run("handles case-insensitive direction", func(t *testing.T) {
		directions := []string{"DESC", "Desc", "desc"}

		for _, dir := range directions {
			isDesc := strings.ToLower(dir) == "desc"
			assert.True(t, isDesc, "Expected %s to be recognized as desc", dir)
		}
	})
}

// =============================================================================
// Count Type Tests
// =============================================================================

func TestCountType(t *testing.T) {
	t.Run("exact count type", func(t *testing.T) {
		countStr := "exact"
		countType := CountType(countStr)

		assert.Equal(t, CountType("exact"), countType)
	})

	t.Run("planned count type", func(t *testing.T) {
		countStr := "planned"
		countType := CountType(countStr)

		assert.Equal(t, CountType("planned"), countType)
	})

	t.Run("estimated count type", func(t *testing.T) {
		countStr := "estimated"
		countType := CountType(countStr)

		assert.Equal(t, CountType("estimated"), countType)
	})
}

// =============================================================================
// Filter Operator Tests
// =============================================================================

func TestFilterOperatorConversion(t *testing.T) {
	t.Run("converts string to FilterOperator", func(t *testing.T) {
		operators := map[string]FilterOperator{
			"eq":    OpEqual,
			"neq":   OpNotEqual,
			"gt":    OpGreaterThan,
			"gte":   OpGreaterOrEqual,
			"lt":    OpLessThan,
			"lte":   OpLessOrEqual,
			"like":  OpLike,
			"ilike": OpILike,
			"in":    OpIn,
			"is":    OpIs,
		}

		for str, expected := range operators {
			actual := FilterOperator(str)
			assert.Equal(t, expected, actual, "Operator %s should convert to %v", str, expected)
		}
	})
}

// =============================================================================
// Content-Range Header Tests
// =============================================================================

func TestContentRangeHeaderQuery(t *testing.T) {
	t.Run("calculates range with offset", func(t *testing.T) {
		offset := 10
		resultsLen := 25
		totalCount := 100

		start := 0
		if &offset != nil {
			start = offset
		}
		end := start + resultsLen - 1
		if end < start {
			end = start
		}

		// Format: start-end/total
		assert.Equal(t, 10, start)
		assert.Equal(t, 34, end)
		assert.Equal(t, 100, totalCount)
	})

	t.Run("handles zero results", func(t *testing.T) {
		offset := 0
		resultsLen := 0

		start := offset
		end := start + resultsLen - 1
		if end < start {
			end = start
		}

		assert.Equal(t, 0, start)
		assert.Equal(t, 0, end) // end = start when no results
	})

	t.Run("handles nil offset", func(t *testing.T) {
		var offset *int
		resultsLen := 10

		start := 0
		if offset != nil {
			start = *offset
		}
		end := start + resultsLen - 1

		assert.Equal(t, 0, start)
		assert.Equal(t, 9, end)
	})
}

// =============================================================================
// Conversion to QueryParams Tests
// =============================================================================

func TestConvertPostQueryToParams(t *testing.T) {
	h := &RESTHandler{}

	t.Run("converts basic request", func(t *testing.T) {
		limit := 50
		offset := 10
		req := &PostQueryRequest{
			Select: "id, name",
			Limit:  &limit,
			Offset: &offset,
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.NotNil(t, params)
		assert.Equal(t, []string{"id", "name"}, params.Select)
		assert.Equal(t, &limit, params.Limit)
		assert.Equal(t, &offset, params.Offset)
	})

	t.Run("converts filters", func(t *testing.T) {
		req := &PostQueryRequest{
			Filters: []PostQueryFilter{
				{Column: "status", Operator: "eq", Value: "active"},
				{Column: "type", Operator: "neq", Value: "deleted"},
			},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Filters, 2)
		assert.Equal(t, "status", params.Filters[0].Column)
		assert.Equal(t, OpEqual, params.Filters[0].Operator)
		assert.False(t, params.Filters[0].IsOr)
	})

	t.Run("converts between filters", func(t *testing.T) {
		req := &PostQueryRequest{
			BetweenFilters: []PostQueryBetweenFilter{
				{Column: "price", Min: 10, Max: 100, Negated: false},
			},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Filters, 2) // Between creates 2 filters
		assert.Equal(t, OpGreaterOrEqual, params.Filters[0].Operator)
		assert.Equal(t, OpLessOrEqual, params.Filters[1].Operator)
	})

	t.Run("converts negated between filters", func(t *testing.T) {
		req := &PostQueryRequest{
			BetweenFilters: []PostQueryBetweenFilter{
				{Column: "age", Min: 18, Max: 65, Negated: true},
			},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Filters, 2) // NOT BETWEEN creates 2 OR filters
		assert.Equal(t, OpLessThan, params.Filters[0].Operator)
		assert.True(t, params.Filters[0].IsOr)
		assert.Equal(t, OpGreaterThan, params.Filters[1].Operator)
		assert.True(t, params.Filters[1].IsOr)
	})

	t.Run("converts OR filters", func(t *testing.T) {
		req := &PostQueryRequest{
			OrFilters: []string{"status.eq.active,status.eq.pending"},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Filters, 2)
		assert.True(t, params.Filters[0].IsOr)
		assert.True(t, params.Filters[1].IsOr)
		assert.Equal(t, params.Filters[0].OrGroupID, params.Filters[1].OrGroupID)
	})

	t.Run("returns error for invalid OR filter", func(t *testing.T) {
		req := &PostQueryRequest{
			OrFilters: []string{"invalid.format"},
		}

		_, err := h.convertPostQueryToParams(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid OR filter format")
	})

	t.Run("converts AND filters", func(t *testing.T) {
		req := &PostQueryRequest{
			AndFilters: []string{"active.eq.true,verified.eq.true"},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Filters, 2)
		assert.False(t, params.Filters[0].IsOr)
		assert.False(t, params.Filters[1].IsOr)
	})

	t.Run("returns error for invalid AND filter", func(t *testing.T) {
		req := &PostQueryRequest{
			AndFilters: []string{"invalid"},
		}

		_, err := h.convertPostQueryToParams(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid AND filter format")
	})

	t.Run("converts order", func(t *testing.T) {
		req := &PostQueryRequest{
			Order: []PostQueryOrderBy{
				{Column: "created_at", Direction: "desc", Nulls: "last"},
				{Column: "name", Direction: "asc"},
			},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.Len(t, params.Order, 2)
		assert.Equal(t, "created_at", params.Order[0].Column)
		assert.True(t, params.Order[0].Desc)
		assert.Equal(t, "last", params.Order[0].Nulls)
		assert.Equal(t, "name", params.Order[1].Column)
		assert.False(t, params.Order[1].Desc)
	})

	t.Run("converts count type", func(t *testing.T) {
		req := &PostQueryRequest{
			Count: "exact",
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		assert.Equal(t, CountType("exact"), params.Count)
	})

	t.Run("converts group by", func(t *testing.T) {
		req := &PostQueryRequest{
			GroupBy: []string{"category", "status"},
		}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		assert.Equal(t, []string{"category", "status"}, params.GroupBy)
	})

	t.Run("handles empty request", func(t *testing.T) {
		req := &PostQueryRequest{}

		params, err := h.convertPostQueryToParams(req)

		require.NoError(t, err)
		require.NotNil(t, params)
		assert.Empty(t, params.Select)
		assert.Empty(t, params.Filters)
		assert.Empty(t, params.Order)
		assert.Nil(t, params.Limit)
		assert.Nil(t, params.Offset)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkSelectParsing(b *testing.B) {
	selectStr := "id, name, email, created_at, updated_at, status, type"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		columns := strings.Split(selectStr, ",")
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
		}
		_ = columns
	}
}

func BenchmarkOrFilterParsing(b *testing.B) {
	orFilter := "status.eq.active,status.eq.pending,status.eq.completed"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parts := strings.Split(orFilter, ",")
		for _, part := range parts {
			_ = strings.SplitN(strings.TrimSpace(part), ".", 3)
		}
	}
}

func BenchmarkConvertPostQueryToParams(b *testing.B) {
	h := &RESTHandler{}
	limit := 50
	offset := 10

	req := &PostQueryRequest{
		Select: "id, name, email",
		Filters: []PostQueryFilter{
			{Column: "status", Operator: "eq", Value: "active"},
		},
		Order: []PostQueryOrderBy{
			{Column: "created_at", Direction: "desc"},
		},
		Limit:  &limit,
		Offset: &offset,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h.convertPostQueryToParams(req)
	}
}
