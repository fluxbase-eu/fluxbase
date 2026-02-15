package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// CollectionStorage handles collection and collection-KB operations
// Collections are shared - access is controlled via collection_members table
type CollectionStorage struct {
	db *database.Connection
}

func NewCollectionStorage(db *database.Connection) *CollectionStorage {
	return &CollectionStorage{db: db}
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Simple slug generation - lowercase and replace spaces with hyphens
	// In production, you might want more sophisticated slug generation
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove non-alphanumeric characters (except hyphens)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ============================================================================
// Collection CRUD with Membership Access Control
// ============================================================================

// GetCollection retrieves a collection by ID with membership check
// Returns the collection with the user's role if they're a member
func (s *CollectionStorage) GetCollection(ctx context.Context, collectionID, userID string) (*Collection, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.description, c.created_by,
		       c.created_at, c.updated_at,
		       COALESCE(cm.role, NULL::text) as user_role,
		       COUNT(DISTINCT cm2.user_id) as member_count,
		       COUNT(DISTINCT ckb.knowledge_base_id) as kb_count
		FROM ai.collections c
		LEFT JOIN ai.collection_members cm ON cm.collection_id = c.id AND cm.user_id = $2
		LEFT JOIN ai.collection_members cm2 ON cm2.collection_id = c.id
		LEFT JOIN ai.collection_knowledge_bases ckb ON ckb.collection_id = c.id
		WHERE c.id = $1
		GROUP BY c.id, c.name, c.slug, c.description, c.created_by, c.created_at, c.updated_at, cm.role
	`

	var c Collection
	var description *string
	var createdBy *string
	var userRole *string
	if err := s.db.QueryRow(ctx, query, collectionID, userID).Scan(
		&c.ID, &c.Name, &c.Slug, &description, &createdBy,
		&c.CreatedAt, &c.UpdatedAt,
		&userRole,
		&c.MemberCount, &c.KBCount,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	c.Description = description
	c.CreatedBy = createdBy
	if userRole != nil {
		role := CollectionRole(*userRole)
		c.CurrentUserRole = &role
	}

	return &c, nil
}

// ListCollections returns all collections the user has access to (is a member of)
func (s *CollectionStorage) ListCollections(ctx context.Context, userID string) ([]CollectionSummary, error) {
	query := `
		SELECT c.id, c.name, c.slug,
		       COALESCE(cm.role, NULL::text) as user_role,
		       COUNT(DISTINCT cm2.user_id) as member_count,
		       COUNT(DISTINCT ckb.knowledge_base_id) as kb_count
		FROM ai.collections c
		INNER JOIN ai.collection_members cm ON cm.collection_id = c.id AND cm.user_id = $1
		LEFT JOIN ai.collection_members cm2 ON cm2.collection_id = c.id
		LEFT JOIN ai.collection_knowledge_bases ckb ON ckb.collection_id = c.id
		GROUP BY c.id, c.name, c.slug, cm.role
		ORDER BY c.name ASC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer rows.Close()

	var collections []CollectionSummary
	for rows.Next() {
		var c CollectionSummary
		var userRole *string
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Slug,
			&userRole,
			&c.MemberCount, &c.KBCount,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan collection row")
			continue
		}

		if userRole != nil {
			role := CollectionRole(*userRole)
			c.UserRole = &role
		}

		collections = append(collections, c)
	}

	return collections, nil
}

// CreateCollection creates a new collection and automatically adds the creator as an owner
func (s *CollectionStorage) CreateCollection(ctx context.Context, userID string, req CreateCollectionRequest) (*Collection, error) {
	// Auto-generate slug from name if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Create the collection
	collectionID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO ai.collections (id, name, slug, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, slug, description, created_by, created_at, updated_at
	`

	var c Collection
	var description *string
	if err := s.db.QueryRow(ctx, query,
		collectionID, req.Name, slug, req.Description,
		userID, now, now,
	).Scan(
		&c.ID, &c.Name, &c.Slug, &description,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}
	c.Description = description

	// Automatically add creator as owner
	if err := s.AddCollectionMember(ctx, collectionID, userID, string(CollectionRoleOwner), userID); err != nil {
		// Rollback collection creation if member addition fails
		_ = s.DeleteCollection(ctx, collectionID, userID)
		return nil, fmt.Errorf("failed to add creator as owner: %w", err)
	}

	// Set derived fields
	c.MemberCount = 1
	c.KBCount = 0
	ownerRole := CollectionRoleOwner
	c.CurrentUserRole = &ownerRole

	return &c, nil
}

// UpdateCollection updates a collection (requires editor or owner role)
func (s *CollectionStorage) UpdateCollection(ctx context.Context, collectionID, userID string, req UpdateCollectionRequest) (*Collection, error) {
	// Check user has editor or owner role
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection role: %w", err)
	}
	if role != string(CollectionRoleEditor) && role != string(CollectionRoleOwner) {
		return nil, fmt.Errorf("insufficient permissions: requires editor or owner role")
	}

	// Build dynamic UPDATE query based on non-nil fields
	updates := []string{}
	args := []interface{}{collectionID}
	argNum := 2

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *req.Name)
		argNum++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argNum))
		args = append(args, *req.Description)
		argNum++
	}

	if len(updates) == 0 {
		// No updates to apply, just return the collection
		return s.GetCollection(ctx, collectionID, userID)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argNum))
	args = append(args, time.Now())

	query := fmt.Sprintf(`UPDATE ai.collections SET %s WHERE id = $1 RETURNING id, name, slug, description, created_by, created_at, updated_at`,
		strings.Join(updates, ", "))

	var c Collection
	var description *string
	var createdBy *string
	if err := s.db.QueryRow(ctx, query, args...).Scan(
		&c.ID, &c.Name, &c.Slug, &description,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}
	c.Description = description
	c.CreatedBy = createdBy

	// Get derived fields
	rolePtr := CollectionRole(role)
	c.CurrentUserRole = &rolePtr
	c.MemberCount, c.KBCount, err = s.getCollectionCounts(ctx, collectionID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get collection counts after update")
	}

	return &c, nil
}

// DeleteCollection deletes a collection (requires owner role)
func (s *CollectionStorage) DeleteCollection(ctx context.Context, collectionID, userID string) error {
	// Check user has owner role
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return fmt.Errorf("failed to check collection role: %w", err)
	}
	if role != string(CollectionRoleOwner) {
		return fmt.Errorf("insufficient permissions: requires owner role")
	}

	// Delete collection (cascade will delete members and KB links)
	query := `DELETE FROM ai.collections WHERE id = $1`
	_, err = s.db.Exec(ctx, query, collectionID)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}

// getCollectionCounts retrieves member and KB counts for a collection
func (s *CollectionStorage) getCollectionCounts(ctx context.Context, collectionID string) (memberCount, kbCount int, err error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM ai.collection_members WHERE collection_id = $1) as member_count,
			(SELECT COUNT(*) FROM ai.collection_knowledge_bases WHERE collection_id = $1) as kb_count
	`
	err = s.db.QueryRow(ctx, query, collectionID).Scan(&memberCount, &kbCount)
	return
}

// ============================================================================
// Collection Membership Management
// ============================================================================

// AddCollectionMember adds a user to a collection (requires owner role)
func (s *CollectionStorage) AddCollectionMember(ctx context.Context, collectionID, userID, role, addedBy string) error {
	// Check adder has owner role
	adderRole, err := s.GetCollectionRole(ctx, collectionID, addedBy)
	if err != nil {
		return fmt.Errorf("failed to check collection role: %w", err)
	}
	if adderRole != string(CollectionRoleOwner) {
		return fmt.Errorf("insufficient permissions: requires owner role")
	}

	// Validate role
	if role != string(CollectionRoleViewer) && role != string(CollectionRoleEditor) && role != string(CollectionRoleOwner) {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Add member (use ON CONFLICT to update if exists)
	query := `
		INSERT INTO ai.collection_members (collection_id, user_id, role, added_by, added_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (collection_id, user_id)
		DO UPDATE SET role = $3, added_by = $4, added_at = NOW()
	`
	_, err = s.db.Exec(ctx, query, collectionID, userID, role, addedBy)
	if err != nil {
		return fmt.Errorf("failed to add collection member: %w", err)
	}

	return nil
}

// RemoveCollectionMember removes a user from a collection (requires owner role)
func (s *CollectionStorage) RemoveCollectionMember(ctx context.Context, collectionID, userID string) error {
	// Check requester has owner role (we'll need the userID of the requester)
	// For now, we'll check in the handler layer - this method will be called after auth check

	query := `DELETE FROM ai.collection_members WHERE collection_id = $1 AND user_id = $2`
	_, err := s.db.Exec(ctx, query, collectionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove collection member: %w", err)
	}

	return nil
}

// ListCollectionMembers lists all members of a collection (requires membership)
func (s *CollectionStorage) ListCollectionMembers(ctx context.Context, collectionID, requestingUserID string) ([]CollectionMember, error) {
	// Check requesting user is a member
	role, err := s.GetCollectionRole(ctx, collectionID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection membership: %w", err)
	}
	if role == "" {
		return nil, fmt.Errorf("user is not a member of this collection")
	}

	query := `
		SELECT collection_id, user_id, role, added_by, added_at
		FROM ai.collection_members
		WHERE collection_id = $1
		ORDER BY
			CASE role
				WHEN 'owner' THEN 1
				WHEN 'editor' THEN 2
				WHEN 'viewer' THEN 3
			END,
			added_at DESC
	`

	rows, err := s.db.Query(ctx, query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list collection members: %w", err)
	}
	defer rows.Close()

	var members []CollectionMember
	for rows.Next() {
		var m CollectionMember
		var addedBy *string
		if err := rows.Scan(&m.CollectionID, &m.UserID, &m.Role, &addedBy, &m.AddedAt); err != nil {
			log.Warn().Err(err).Msg("Failed to scan collection member row")
			continue
		}
		m.AddedBy = addedBy
		members = append(members, m)
	}

	return members, nil
}

// UpdateCollectionMemberRole updates a member's role (requires owner role)
func (s *CollectionStorage) UpdateCollectionMemberRole(ctx context.Context, collectionID, userID, newRole, updatedBy string) error {
	// Check updater has owner role
	updaterRole, err := s.GetCollectionRole(ctx, collectionID, updatedBy)
	if err != nil {
		return fmt.Errorf("failed to check collection role: %w", err)
	}
	if updaterRole != string(CollectionRoleOwner) {
		return fmt.Errorf("insufficient permissions: requires owner role")
	}

	// Validate new role
	if newRole != string(CollectionRoleViewer) && newRole != string(CollectionRoleEditor) && newRole != string(CollectionRoleOwner) {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	query := `
		UPDATE ai.collection_members
		SET role = $1, added_by = $2, added_at = NOW()
		WHERE collection_id = $3 AND user_id = $4
	`
	_, err = s.db.Exec(ctx, query, newRole, updatedBy, collectionID, userID)
	if err != nil {
		return fmt.Errorf("failed to update collection member role: %w", err)
	}

	return nil
}

// GetCollectionRole retrieves a user's role in a collection
// Returns empty string if user is not a member
func (s *CollectionStorage) GetCollectionRole(ctx context.Context, collectionID, userID string) (string, error) {
	query := `
		SELECT role
		FROM ai.collection_members
		WHERE collection_id = $1 AND user_id = $2
	`

	var role string
	err := s.db.QueryRow(ctx, query, collectionID, userID).Scan(&role)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get collection role: %w", err)
	}

	return role, nil
}

// IsCollectionMember checks if a user is a member of a collection
func (s *CollectionStorage) IsCollectionMember(ctx context.Context, collectionID, userID string) (bool, error) {
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return false, err
	}
	return role != "", nil
}

// CanUserEditCollection checks if a user can edit a collection (editor or owner role)
func (s *CollectionStorage) CanUserEditCollection(ctx context.Context, collectionID, userID string) (bool, error) {
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return false, err
	}
	return role == string(CollectionRoleEditor) || role == string(CollectionRoleOwner), nil
}

// CanUserManageCollection checks if a user can manage a collection (owner role only)
func (s *CollectionStorage) CanUserManageCollection(ctx context.Context, collectionID, userID string) (bool, error) {
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return false, err
	}
	return role == string(CollectionRoleOwner), nil
}

// ============================================================================
// Collection-Knowledge Base Links
// ============================================================================

// LinkKnowledgeBaseToCollection links a KB to a collection (requires editor or owner role)
func (s *CollectionStorage) LinkKnowledgeBaseToCollection(ctx context.Context, collectionID, kbID, userID string) error {
	// Check user has editor or owner role
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return fmt.Errorf("failed to check collection role: %w", err)
	}
	if role != string(CollectionRoleEditor) && role != string(CollectionRoleOwner) {
		return fmt.Errorf("insufficient permissions: requires editor or owner role")
	}

	// Verify KB exists
	// Note: KB access control should be checked separately based on visibility

	query := `
		INSERT INTO ai.collection_knowledge_bases (collection_id, knowledge_base_id, added_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (collection_id, knowledge_base_id) DO NOTHING
	`
	_, err = s.db.Exec(ctx, query, collectionID, kbID)
	if err != nil {
		return fmt.Errorf("failed to link knowledge base to collection: %w", err)
	}

	return nil
}

// UnlinkKnowledgeBaseFromCollection unlinks a KB from a collection (requires editor or owner role)
func (s *CollectionStorage) UnlinkKnowledgeBaseFromCollection(ctx context.Context, collectionID, kbID, userID string) error {
	// Check user has editor or owner role
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return fmt.Errorf("failed to check collection role: %w", err)
	}
	if role != string(CollectionRoleEditor) && role != string(CollectionRoleOwner) {
		return fmt.Errorf("insufficient permissions: requires editor or owner role")
	}

	query := `DELETE FROM ai.collection_knowledge_bases WHERE collection_id = $1 AND knowledge_base_id = $2`
	_, err = s.db.Exec(ctx, query, collectionID, kbID)
	if err != nil {
		return fmt.Errorf("failed to unlink knowledge base from collection: %w", err)
	}

	return nil
}

// ListCollectionKnowledgeBases lists all KBs in a collection (requires membership)
func (s *CollectionStorage) ListCollectionKnowledgeBases(ctx context.Context, collectionID, userID string) ([]KnowledgeBaseSummary, error) {
	// Check user is a member
	role, err := s.GetCollectionRole(ctx, collectionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check collection membership: %w", err)
	}
	if role == "" {
		return nil, fmt.Errorf("user is not a member of this collection")
	}

	query := `
		SELECT kb.id, kb.name, kb.namespace, kb.description, kb.enabled,
		       kb.document_count, kb.total_chunks, kb.updated_at,
		       kb.visibility
		FROM ai.knowledge_bases kb
		JOIN ai.collection_knowledge_bases ckb ON ckb.knowledge_base_id = kb.id
		WHERE ckb.collection_id = $1
		ORDER BY ckb.added_at DESC
	`

	rows, err := s.db.Query(ctx, query, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list collection knowledge bases: %w", err)
	}
	defer rows.Close()

	var kbs []KnowledgeBaseSummary
	for rows.Next() {
		var kb KnowledgeBaseSummary
		var description *string
		if err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Namespace, &description, &kb.Enabled,
			&kb.DocumentCount, &kb.TotalChunks, &kb.UpdatedAt,
			&kb.Visibility,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan knowledge base row")
			continue
		}
		if description != nil {
			kb.Description = *description
		}
		kbs = append(kbs, kb)
	}

	return kbs, nil
}
