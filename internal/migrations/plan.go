package migrations

import "time"

// Plan represents a migration plan from pgschema
type Plan struct {
	Changes     []Change      `json:"changes"`
	DDL         string        `json:"ddl"`
	Transaction bool          `json:"transaction"`
	Summary     *PlanSummary  `json:"summary,omitempty"`
	Duration    time.Duration `json:"-"`
}

// PlanSummary provides a summary of the plan
type PlanSummary struct {
	TotalChanges     int `json:"total_changes"`
	CreateCount      int `json:"create_count"`
	AlterCount       int `json:"alter_count"`
	DropCount        int `json:"drop_count"`
	DestructiveCount int `json:"destructive_count"`
}

// Change represents a single schema change
type Change struct {
	Type        ChangeType `json:"type"`
	ObjectType  string     `json:"object_type"`
	Schema      string     `json:"schema"`
	Name        string     `json:"name"`
	SQL         string     `json:"sql"`
	Destructive bool       `json:"destructive"`
	DependsOn   []string   `json:"depends_on,omitempty"`
}

// ChangeType represents the type of schema change
type ChangeType string

const (
	ChangeCreate ChangeType = "CREATE"
	ChangeAlter  ChangeType = "ALTER"
	ChangeDrop   ChangeType = "DROP"
)

// ApplyResult represents the result of applying a plan
type ApplyResult struct {
	Applied  []Change
	Duration time.Duration
	Error    error
}

// ValidationResult represents the result of schema validation
type ValidationResult struct {
	Valid  bool
	Drifts []Drift
	Error  error
}

// Drift represents a schema drift between declared and actual state
type Drift struct {
	Type        string `json:"type"`
	ObjectType  string `json:"object_type"`
	Schema      string `json:"schema"`
	Name        string `json:"name"`
	SQL         string `json:"sql"`
	Destructive bool   `json:"destructive"`
}

// MigrationState represents the current state of migrations
type MigrationState struct {
	HasImperativeMigrations bool
	HasDeclarativeState     bool
	LastAppliedVersion      int64
	SchemaFingerprint       string
	HasDirtyMigrations      bool // For informational logging only - not blocking
}

// DeclarativeState represents the state record in migrations.declarative_state
type DeclarativeState struct {
	ID                int       `json:"id"`
	SchemaFingerprint string    `json:"schema_fingerprint"`
	AppliedAt         time.Time `json:"applied_at"`
	AppliedBy         string    `json:"applied_by"`
	Source            string    `json:"source"` // 'fresh_install' | 'transitioned' | 'schema_apply'
}
