package api

import (
	"fmt"
	"strings"
)

const (
	// MaxFilterDepth bounds recursion for nested and()/or() filter expressions.
	// Deeper expressions are rejected before they can compile to pathological SQL.
	MaxFilterDepth = 10

	// MaxFilterCount caps the number of filters parsed from a single request.
	MaxFilterCount = 100
)

// parseFilter parses filter parameters
func (qp *QueryParser) parseFilter(key, value string, params *QueryParams) error {
	// Handle logical operators (depth 1: top-level or()/and() group)
	if key == "or" {
		return qp.parseLogicalFilter(value, params, true, 1)
	}
	if key == "and" {
		return qp.parseLogicalFilter(value, params, false, 1)
	}

	// Check for classic format first: column.operator=value
	// This takes precedence over PostgREST format
	if strings.Contains(key, ".") {
		parts := strings.SplitN(key, ".", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid filter format: %s", key)
		}

		column := parts[0]
		operator := FilterOperator(parts[1])

		// Validate column reference before it is interpolated into SQL
		if !isValidColumnReference(column) {
			return fmt.Errorf("invalid filter column name: %s", column)
		}

		// Parse value based on operator
		filterValue, err := qp.parseFilterValue(operator, value)
		if err != nil {
			return err
		}

		if err := addFilter(params, Filter{
			Column:   column,
			Operator: operator,
			Value:    filterValue,
			IsOr:     false,
		}); err != nil {
			return err
		}

		return nil
	}

	// Try PostgREST format: column=operator.value
	// Split value by first dot to extract operator
	dotIndex := strings.Index(value, ".")
	if dotIndex > 0 {
		// PostgREST format: column=operator.value
		column := key
		operatorStr := value[:dotIndex]
		filterValue := value[dotIndex+1:]

		// Validate column reference before it is interpolated into SQL
		if !isValidColumnReference(column) {
			return fmt.Errorf("invalid filter column name: %s", column)
		}

		operator := FilterOperator(operatorStr)

		// Parse value based on operator
		parsedValue, err := qp.parseFilterValue(operator, filterValue)
		if err != nil {
			return err
		}

		if err := addFilter(params, Filter{
			Column:   column,
			Operator: operator,
			Value:    parsedValue,
			IsOr:     false,
		}); err != nil {
			return err
		}

		return nil
	}

	// If neither format matched, return an error
	return fmt.Errorf("invalid filter format: %s", key)
}

// parseLogicalFilter parses or/and grouped filters with support for nested expressions
// Supports formats like:
//   - or=(name.eq.John,age.gt.30)
//   - and=(or(col.lt.min1,col.gt.max1),or(col.lt.min2,col.gt.max2))
//
// depth tracks and()/or() nesting and is capped at MaxFilterDepth to prevent
// unbounded recursion and pathological SQL generation.
func (qp *QueryParser) parseLogicalFilter(value string, params *QueryParams, isOr bool, depth int) error {
	if depth > MaxFilterDepth {
		return fmt.Errorf("filter expression exceeds maximum nesting depth of %d", MaxFilterDepth)
	}

	// Parse format: or=(name.eq.John,age.gt.30)
	// Only remove one pair of outer parentheses (not all leading/trailing parens)
	if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
		value = value[1 : len(value)-1]
	}

	// Use parentheses-aware splitting to handle nested expressions
	filters, err := qp.parseNestedFilters(value)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}

		// Check for nested or() expression
		if strings.HasPrefix(filter, "or(") && strings.HasSuffix(filter, ")") {
			// Nested OR expression - parse recursively with new group ID
			innerValue := strings.TrimPrefix(filter, "or(")
			innerValue = strings.TrimSuffix(innerValue, ")")
			if err := qp.parseNestedOrGroup(innerValue, params); err != nil {
				return err
			}
			continue
		}

		// Check for nested and() expression
		if strings.HasPrefix(filter, "and(") && strings.HasSuffix(filter, ")") {
			// Nested AND expression - parse recursively
			innerValue := strings.TrimPrefix(filter, "and(")
			innerValue = strings.TrimSuffix(innerValue, ")")
			if err := qp.parseLogicalFilter(innerValue, params, false, depth+1); err != nil {
				return err
			}
			continue
		}

		// Regular filter: column.operator.value
		parts := strings.SplitN(filter, ".", 3)
		if len(parts) != 3 {
			return fmt.Errorf("invalid filter format in logical group: %s", filter)
		}

		column := parts[0]
		operator := FilterOperator(parts[1])
		rawValue := parts[2]

		// Validate column reference before it is interpolated into SQL
		if !isValidColumnReference(column) {
			return fmt.Errorf("invalid filter column name: %s", column)
		}

		// Parse value based on operator (same logic as regular filter parsing)
		parsedValue, err := qp.parseFilterValue(operator, rawValue)
		if err != nil {
			return err
		}

		if err := addFilter(params, Filter{
			Column:   column,
			Operator: operator,
			Value:    parsedValue,
			IsOr:     isOr,
		}); err != nil {
			return err
		}
	}

	return nil
}

// parseNestedOrGroup parses an OR group and assigns a unique group ID to all filters
func (qp *QueryParser) parseNestedOrGroup(value string, params *QueryParams) error {
	// Increment group counter for this OR group
	params.orGroupCounter++
	groupID := params.orGroupCounter

	// Split by comma (respecting parentheses)
	filters, err := qp.parseNestedFilters(value)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}

		// Parse each filter: column.operator.value
		parts := strings.SplitN(filter, ".", 3)
		if len(parts) != 3 {
			return fmt.Errorf("invalid filter format in OR group: %s", filter)
		}

		column := parts[0]
		operator := FilterOperator(parts[1])
		rawValue := parts[2]

		// Validate column reference before it is interpolated into SQL
		if !isValidColumnReference(column) {
			return fmt.Errorf("invalid filter column name: %s", column)
		}

		// Parse value based on operator (same logic as regular filter parsing)
		parsedValue, err := qp.parseFilterValue(operator, rawValue)
		if err != nil {
			return err
		}

		if err := addFilter(params, Filter{
			Column:    column,
			Operator:  operator,
			Value:     parsedValue,
			IsOr:      true,
			OrGroupID: groupID,
		}); err != nil {
			return err
		}
	}

	return nil
}

// parseNestedFilters splits a filter string by commas while respecting parentheses nesting
func (qp *QueryParser) parseNestedFilters(value string) ([]string, error) {
	var filters []string
	var current strings.Builder
	depth := 0

	for _, ch := range value {
		switch ch {
		case '(':
			depth++
			current.WriteRune(ch)
		case ')':
			depth--
			current.WriteRune(ch)
			if depth < 0 {
				return nil, fmt.Errorf("unbalanced parentheses in filter expression")
			}
		case ',':
			if depth == 0 {
				if s := strings.TrimSpace(current.String()); s != "" {
					filters = append(filters, s)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if depth != 0 {
		return nil, fmt.Errorf("unbalanced parentheses in filter expression")
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		filters = append(filters, s)
	}

	return filters, nil
}

// addFilter appends filter to params.Filters, enforcing MaxFilterCount.
func addFilter(params *QueryParams, filter Filter) error {
	if len(params.Filters) >= MaxFilterCount {
		return fmt.Errorf("too many filters: request exceeds the maximum of %d filter expressions", MaxFilterCount)
	}
	params.Filters = append(params.Filters, filter)
	return nil
}

// parseFilterValue parses a raw filter value based on its operator.
// It is the single value-parsing path for all filter forms (classic
// column.op=value, PostgREST column=op.value, logical groups, and nested
// OR groups) so they cannot drift apart.
func (qp *QueryParser) parseFilterValue(op FilterOperator, raw string) (interface{}, error) {
	switch op {
	case OpIn:
		// Parse array values: (1,2,3) or ["a","b","c"]
		return qp.parseArrayValue(raw), nil
	case OpBetween:
		// Parse range values: (1,10)
		return qp.parseBetweenValue(raw)
	case OpIs:
		// Parse null/true/false - H-14: Validate boolean values
		switch raw {
		case "null":
			return nil, nil
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid value for OpIs operator: %s (must be null, true, or false)", raw)
		}
	default:
		return raw, nil
	}
}

// parseArrayValue parses array values from string
func (qp *QueryParser) parseArrayValue(value string) []string {
	// Remove parentheses or brackets
	value = strings.Trim(value, "()[]")

	// Split by comma
	items := strings.Split(value, ",")
	result := make([]string, len(items))

	for i, item := range items {
		// Remove quotes if present
		result[i] = strings.Trim(strings.TrimSpace(item), "\"'")
	}

	return result
}

// parseBetweenValue parses a between range value: (1,10) or ["a","z"].
// Exactly two bounds are required; they keep their original order.
func (qp *QueryParser) parseBetweenValue(value string) ([]interface{}, error) {
	items := qp.parseArrayValue(value)
	if len(items) != 2 {
		return nil, fmt.Errorf("invalid value for between operator: %s (exactly two bounds required, e.g. (1,10))", value)
	}
	return []interface{}{items[0], items[1]}, nil
}
