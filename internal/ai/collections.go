package ai

import (
	"time"
)

// CollectionRole defines access level for collections
type CollectionRole string

const (
	CollectionRoleViewer CollectionRole = "viewer" // Can view only
	CollectionRoleEditor CollectionRole = "editor" // Can view and add/remove KBs
	CollectionRoleOwner  CollectionRole = "owner"  // Full control including managing members
)

// Collection represents a shared collection for organizing knowledge bases
// Collections are not owned by a single user - access is controlled via collection_members
type Collection struct {
	ID              string          `json:"id" db:"id"`
	Name            string          `json:"name" db:"name"`
	Slug            string          `json:"slug" db:"slug"`
	Description     *string         `json:"description,omitempty" db:"description"`
	CreatedBy       *string         `json:"created_by,omitempty" db:"created_by"`               // User who created the collection
	MemberCount     int             `json:"member_count,omitempty" db:"member_count"`           // Derived field
	KBCount         int             `json:"kb_count,omitempty" db:"kb_count"`                   // Derived field
	CurrentUserRole *CollectionRole `json:"current_user_role,omitempty" db:"current_user_role"` // Derived field for current user
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// CollectionMember represents a user's membership in a collection
type CollectionMember struct {
	CollectionID string         `json:"collection_id" db:"collection_id"`
	UserID       string         `json:"user_id" db:"user_id"`
	Role         CollectionRole `json:"role" db:"role"`
	AddedBy      *string        `json:"added_by,omitempty" db:"added_by"` // User who added this member
	AddedAt      time.Time      `json:"added_at" db:"added_at"`
}

// CollectionSummary for list views
type CollectionSummary struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	Slug        string          `json:"slug" db:"slug"`
	KBCount     int             `json:"kb_count" db:"kb_count"`
	MemberCount int             `json:"member_count" db:"member_count"`
	UserRole    *CollectionRole `json:"user_role,omitempty" db:"user_role"` // Current user's role
}

// CreateCollectionRequest creates new collection
type CreateCollectionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug,omitempty" validate:"omitempty,max=100"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// UpdateCollectionRequest updates collection
type UpdateCollectionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// CollectionKBLink represents collection-KB relationship
type CollectionKBLink struct {
	CollectionID    string    `json:"collection_id" db:"collection_id"`
	KnowledgeBaseID string    `json:"knowledge_base_id" db:"knowledge_base_id"`
	AddedAt         time.Time `json:"added_at" db:"added_at"`
}

// LinkKBToCollectionRequest links KB to collection
type LinkKBToCollectionRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id" validate:"required"`
}

// UnlinkKBFromCollectionRequest unlinks KB
type UnlinkKBFromCollectionRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id" validate:"required"`
}

// ChatbotKnowledgeSources defines KB sources for chatbot
type ChatbotKnowledgeSources struct {
	CollectionIDs    []string `json:"collection_ids,omitempty"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids,omitempty"`
}
