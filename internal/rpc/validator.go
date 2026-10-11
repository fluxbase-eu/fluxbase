package rpc

import (
	"encoding/json"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
	"github.com/rs/zerolog/log"
)

// sqlParamToken is a named `$name` placeholder occurrence in SQL text.
type sqlParamToken struct {
	start int    // byte offset of the leading '$'
	name  string // placeholder name without the leading '$'
}

// scanNamedParamTokens returns every `$name` placeholder token in sql in order
// of appearance, skipping anything inside single-quoted strings (with doubled
// quote escapes), double-quoted identifiers (with doubled quote escapes),
// dollar-quoted blocks (`$tag$ ... $tag$`, including the empty-tag `$$` form),
// `--` line comments and `/* */` block comments. Ordinal placeholders such as
// `$1` are left alone and not reported.
//
// SECURITY (SEC-12): this context-aware scan is shared by the executor's
// parameter rewriting and the validator's preprocessing so both agree on what
// constitutes a parameter placeholder. A context-free regex would rewrite
// `$name` tokens inside string literals or procedure bodies, corrupting the
// SQL and shifting parameter indices.
func scanNamedParamTokens(sql string) []sqlParamToken {
	var tokens []sqlParamToken
	i, n := 0, len(sql)
	for i < n {
		switch c := sql[i]; {
		case c == '\'':
			i = skipSingleQuoted(sql, i)
		case c == '"':
			i = skipDoubleQuoted(sql, i)
		case c == '-' && i+1 < n && sql[i+1] == '-':
			i = skipLineComment(sql, i)
		case c == '/' && i+1 < n && sql[i+1] == '*':
			i = skipBlockComment(sql, i)
		case c == '$':
			if end, ok := dollarQuoteEnd(sql, i); ok {
				// Dollar-quoted block (`$$...$$`, `$tag$...$tag$`): skip to
				// the matching closing tag. An unterminated block consumes
				// the remainder of the input.
				i = end
				continue
			}
			// A named placeholder must start with a letter or underscore;
			// `$1`-style ordinal placeholders stay untouched.
			if i+1 < n && isParamIdentStart(sql[i+1]) {
				j := i + 2
				for j < n && isParamIdentChar(sql[j]) {
					j++
				}
				tokens = append(tokens, sqlParamToken{start: i, name: sql[i+1 : j]})
				i = j
				continue
			}
			i++
		default:
			i++
		}
	}
	return tokens
}

// dollarQuoteEnd reports whether a dollar-quoted block starts at sql[i] (which
// must be '$'). The opening tag is '$' + optional identifier + '$'; the empty
// identifier gives the `$$` form. If a block starts at i it returns the index
// just past the matching closing tag; an unterminated block returns len(sql).
// A trailing identifier not followed by '$' (e.g. a `$name` placeholder) is not
// a tag, so ok is false.
func dollarQuoteEnd(sql string, i int) (int, bool) {
	j := i + 1
	for j < len(sql) && isParamIdentChar(sql[j]) {
		j++
	}
	if j >= len(sql) || sql[j] != '$' {
		return 0, false
	}
	tag := sql[i : j+1] // "$tag$" or "$$"
	if rest := strings.Index(sql[j+1:], tag); rest >= 0 {
		return j + 1 + rest + len(tag), true
	}
	return len(sql), true
}

// skipSingleQuoted returns the index just past a single-quoted string starting
// at sql[i] (which must hold the opening quote byte). A doubled quote inside is
// an escaped quote; an unterminated string consumes the remainder of the input.
func skipSingleQuoted(sql string, i int) int {
	i++
	for i < len(sql) {
		if sql[i] == '\'' {
			if i+1 < len(sql) && sql[i+1] == '\'' {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return i
}

// skipDoubleQuoted returns the index just past a double-quoted identifier
// starting at sql[i] (which must be '"'). A doubled "" is an escaped quote; an
// unterminated identifier consumes the remainder of the input.
func skipDoubleQuoted(sql string, i int) int {
	i++
	for i < len(sql) {
		if sql[i] == '"' {
			if i+1 < len(sql) && sql[i+1] == '"' {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return i
}

// skipLineComment returns the index just past a `--` line comment starting at
// sql[i], consuming the trailing newline if present.
func skipLineComment(sql string, i int) int {
	if nl := strings.IndexByte(sql[i:], '\n'); nl >= 0 {
		return i + nl + 1
	}
	return len(sql)
}

// skipBlockComment returns the index just past a `/*` block comment starting at
// sql[i]; an unterminated comment consumes the remainder of the input.
func skipBlockComment(sql string, i int) int {
	if end := strings.Index(sql[i+2:], "*/"); end >= 0 {
		return i + 2 + end + 2
	}
	return len(sql)
}

// isParamIdentStart reports whether c may start a placeholder name.
func isParamIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isParamIdentChar reports whether c may appear inside a placeholder name or a
// dollar-quote tag.
func isParamIdentChar(c byte) bool {
	return isParamIdentStart(c) || (c >= '0' && c <= '9')
}

// Validator handles validation of RPC inputs and SQL queries
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidationResult represents the result of SQL validation
type ValidationResult struct {
	Valid          bool     `json:"valid"`
	Errors         []string `json:"errors,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	TablesAccessed []string `json:"tables_accessed,omitempty"`
	OperationsUsed []string `json:"operations_used,omitempty"`
}

// ValidateInput validates input parameters against a JSON schema
func (v *Validator) ValidateInput(params map[string]interface{}, schemaBytes json.RawMessage) error {
	if len(schemaBytes) == 0 {
		// No schema defined, accept any input
		return nil
	}

	// Parse the schema
	var schema map[string]string
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return fmt.Errorf("invalid input schema: %w", err)
	}

	// Validate each required field
	for fieldName, fieldType := range schema {
		isOptional := IsOptionalField(fieldName)
		cleanName := CleanFieldName(fieldName)

		value, exists := params[cleanName]

		// Check required fields
		if !exists {
			if !isOptional {
				return fmt.Errorf("missing required parameter: %s", cleanName)
			}
			continue
		}

		// Validate type
		if err := v.validateType(cleanName, value, fieldType); err != nil {
			return err
		}
	}

	return nil
}

// validateType validates a value against an expected type
func (v *Validator) validateType(fieldName string, value interface{}, expectedType string) error {
	if value == nil {
		return nil // Null values are allowed
	}

	switch strings.ToLower(expectedType) {
	case "uuid":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter %s must be a UUID string", fieldName)
		}
	case "string", "text":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter %s must be a string", fieldName)
		}
	case "number", "int", "integer":
		switch value.(type) {
		case int, int32, int64, float32, float64, json.Number:
			// Valid numeric types
		default:
			return fmt.Errorf("parameter %s must be a number", fieldName)
		}
	case "float", "double", "decimal":
		switch value.(type) {
		case float32, float64, json.Number:
			// Valid float types
		default:
			return fmt.Errorf("parameter %s must be a decimal number", fieldName)
		}
	case "boolean", "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter %s must be a boolean", fieldName)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("parameter %s must be an array", fieldName)
		}
	case "object", "json", "jsonb":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("parameter %s must be an object", fieldName)
		}
	}

	return nil
}

// preprocessNamedParams replaces $param_name placeholders with dummy values
// so pg_query can parse the SQL. Returns the preprocessed SQL and extracted param names.
// Placeholders inside string literals, quoted identifiers, dollar-quoted blocks
// and comments are left untouched (see scanNamedParamTokens).
func (v *Validator) preprocessNamedParams(sql string) (string, []string) {
	tokens := scanNamedParamTokens(sql)

	var paramNames []string
	var processed strings.Builder
	processed.Grow(len(sql))
	last := 0
	for i, tok := range tokens {
		processed.WriteString(sql[last:tok.start])
		// Replace with PostgreSQL positional parameter ($1, $2, etc.)
		fmt.Fprintf(&processed, "$%d", i+1)
		paramNames = append(paramNames, tok.name)
		last = tok.start + 1 + len(tok.name)
	}
	processed.WriteString(sql[last:])

	return processed.String(), paramNames
}

// ValidateSQL validates a SQL query for safety and compliance
func (v *Validator) ValidateSQL(sql string, allowedTables, allowedSchemas []string) *ValidationResult {
	result := &ValidationResult{
		Valid:          true,
		TablesAccessed: []string{},
		OperationsUsed: []string{},
	}

	// Build lookup maps
	tablesMap := make(map[string]bool)
	for _, t := range allowedTables {
		tablesMap[strings.ToLower(t)] = true
	}

	schemasMap := make(map[string]bool)
	for _, s := range allowedSchemas {
		schemasMap[strings.ToLower(s)] = true
	}

	// Basic safety checks
	lowerSQL := strings.ToLower(sql)

	// Blocked patterns
	blockedPatterns := []string{
		"pg_catalog",
		"information_schema",
		"pg_temp",
		"pg_toast",
		"xp_",      // MSSQL stored procedures (defense in depth)
		"exec(",    // Execution attempt
		"execute(", // Execution attempt
	}

	for _, pattern := range blockedPatterns {
		if strings.Contains(lowerSQL, strings.ToLower(pattern)) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Query contains blocked pattern: %s", pattern))
		}
	}

	// Preprocess named parameters ($param_name -> $1, $2, etc.) for parsing
	processedSQL, _ := v.preprocessNamedParams(sql)

	// Parse the SQL
	parseResult, err := pg_query.Parse(processedSQL)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse SQL: %s", err.Error()))
		return result
	}

	// Check for multiple statements
	if len(parseResult.Stmts) > 1 {
		result.Valid = false
		result.Errors = append(result.Errors, "Multiple SQL statements not allowed")
		return result
	}

	if len(parseResult.Stmts) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Empty SQL statement")
		return result
	}

	// Get the statement
	stmt := parseResult.Stmts[0].Stmt

	// Determine operation type
	operationType := v.getOperationType(stmt)
	if operationType != "" {
		result.OperationsUsed = append(result.OperationsUsed, operationType)
	}

	// Extract and validate tables
	tables := v.extractTables(stmt)
	result.TablesAccessed = tables

	for _, table := range tables {
		tableLower := strings.ToLower(table)

		// Check schema restrictions
		if strings.Contains(table, ".") {
			parts := strings.SplitN(table, ".", 2)
			schema := strings.ToLower(parts[0])
			tableName := parts[1]

			if len(schemasMap) > 0 && !schemasMap[schema] {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("Schema not allowed: %s", schema))
			}

			// If specific tables are allowed, check table name
			if len(tablesMap) > 0 {
				if !tablesMap[strings.ToLower(tableName)] && !tablesMap[tableLower] {
					result.Valid = false
					result.Errors = append(result.Errors, fmt.Sprintf("Table not allowed: %s", table))
				}
			}
		} else if len(tablesMap) > 0 && !tablesMap[tableLower] {
			// No schema prefix - check against allowed tables
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Table not allowed: %s", table))
		}
	}

	return result
}

// ValidateAccess checks if a user with the given role can access the procedure
//
// SECURITY IMPLICATIONS:
//
// Service role (service_role and instance_admin) bypass all access checks:
// - These roles have unrestricted access to ALL procedures
// - If service role credentials are compromised, an attacker gains full database access
// - Service role tokens should NEVER be exposed to client-side code
// - Service role credentials should be stored securely (environment variables, secret managers)
//
// Recommended safeguards:
// 1. Restrict service role usage to server-side code only
// 2. Implement IP whitelisting for service role endpoints (if possible)
// 3. Enable comprehensive audit logging for all service role operations
// 4. Consider rate limiting specifically for service roles
// 5. Rotate service role secrets regularly
// 6. Never commit service role credentials to version control
//
// The bypass is necessary because service_role tokens don't have a user_id,
// so isAuthenticated will be false even though service_role should have full access.
func (v *Validator) ValidateAccess(proc *Procedure, userRole string, isAuthenticated bool) error {
	// Service roles bypass all checks (check first, before authentication)
	// This is important because service_role tokens don't have a user_id,
	// so isAuthenticated will be false even though service_role should have full access
	if userRole == "service_role" || userRole == "instance_admin" || userRole == "tenant_service" {
		return nil
	}

	// Check public access
	if !isAuthenticated && !proc.IsPublic {
		return fmt.Errorf("procedure requires authentication")
	}

	// Check role requirement (OR semantics - user needs ANY of the required roles)
	if len(proc.RequireRoles) > 0 {
		// Check if user's role satisfies ANY of the required roles
		for _, requiredRole := range proc.RequireRoles {
			switch requiredRole {
			case "anon":
				// Anonymous access allowed
				return nil
			case "authenticated":
				if isAuthenticated {
					return nil
				}
			default:
				// Specific role required - exact match
				if userRole == requiredRole {
					return nil
				}
			}
		}

		// None of the required roles matched
		return fmt.Errorf("procedure requires one of roles: %v", proc.RequireRoles)
	}

	return nil
}

// getOperationType determines the SQL operation type from a parsed statement
func (v *Validator) getOperationType(stmt *pg_query.Node) string {
	if stmt == nil {
		return ""
	}

	switch {
	case stmt.GetSelectStmt() != nil:
		return "SELECT"
	case stmt.GetInsertStmt() != nil:
		return "INSERT"
	case stmt.GetUpdateStmt() != nil:
		return "UPDATE"
	case stmt.GetDeleteStmt() != nil:
		return "DELETE"
	case stmt.GetCreateStmt() != nil:
		return "CREATE"
	case stmt.GetDropStmt() != nil:
		return "DROP"
	case stmt.GetAlterTableStmt() != nil:
		return "ALTER"
	case stmt.GetTruncateStmt() != nil:
		return "TRUNCATE"
	default:
		return "UNKNOWN"
	}
}

// extractTables extracts table names from a parsed statement
func (v *Validator) extractTables(stmt *pg_query.Node) []string {
	if stmt == nil {
		return nil
	}

	tables := make(map[string]bool)
	v.walkNode(stmt, tables)

	var result []string
	for table := range tables {
		result = append(result, table)
	}
	return result
}

// walkNode recursively walks the AST to find table references
func (v *Validator) walkNode(node *pg_query.Node, tables map[string]bool) {
	if node == nil {
		return
	}

	// Check for RangeVar (table reference)
	if rv := node.GetRangeVar(); rv != nil {
		tableName := rv.Relname
		if rv.Schemaname != "" {
			tableName = rv.Schemaname + "." + tableName
		}
		if tableName != "" {
			tables[tableName] = true
		}
	}

	// Check for SelectStmt
	if sel := node.GetSelectStmt(); sel != nil {
		for _, from := range sel.FromClause {
			v.walkNode(from, tables)
		}
		if sel.WhereClause != nil {
			v.walkNode(sel.WhereClause, tables)
		}
		for _, target := range sel.TargetList {
			v.walkNode(target, tables)
		}
	}

	// Check for InsertStmt
	if ins := node.GetInsertStmt(); ins != nil {
		if ins.Relation != nil {
			tableName := ins.Relation.Relname
			if ins.Relation.Schemaname != "" {
				tableName = ins.Relation.Schemaname + "." + tableName
			}
			tables[tableName] = true
		}
		if ins.SelectStmt != nil {
			v.walkNode(ins.SelectStmt, tables)
		}
	}

	// Check for UpdateStmt
	if upd := node.GetUpdateStmt(); upd != nil {
		if upd.Relation != nil {
			tableName := upd.Relation.Relname
			if upd.Relation.Schemaname != "" {
				tableName = upd.Relation.Schemaname + "." + tableName
			}
			tables[tableName] = true
		}
		for _, from := range upd.FromClause {
			v.walkNode(from, tables)
		}
	}

	// Check for DeleteStmt
	if del := node.GetDeleteStmt(); del != nil {
		if del.Relation != nil {
			tableName := del.Relation.Relname
			if del.Relation.Schemaname != "" {
				tableName = del.Relation.Schemaname + "." + tableName
			}
			tables[tableName] = true
		}
	}

	// Check for JoinExpr
	if join := node.GetJoinExpr(); join != nil {
		v.walkNode(join.Larg, tables)
		v.walkNode(join.Rarg, tables)
	}

	// Check for SubLink (subquery)
	if sublink := node.GetSubLink(); sublink != nil {
		v.walkNode(sublink.Subselect, tables)
	}

	// Check for CommonTableExpr (WITH clause)
	if cte := node.GetCommonTableExpr(); cte != nil {
		v.walkNode(cte.Ctequery, tables)
	}

	// Log unhandled node types for debugging
	log.Trace().Interface("node_type", fmt.Sprintf("%T", node.Node)).Msg("Walking AST node")
}
