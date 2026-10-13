package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// ResponseCacheSemanticThreshold is the minimum cosine similarity between a
// stored question embedding and an incoming question for a semantic cache
// hit. High on purpose: a wrong cache hit is worse than a cache miss.
const ResponseCacheSemanticThreshold = 0.95

// CachedResponse is one cached chatbot answer.
type CachedResponse struct {
	ID               string    `json:"id"`
	ChatbotID        string    `json:"chatbot_id"`
	Model            string    `json:"model"`
	Question         string    `json:"question"`
	QuestionHash     string    `json:"question_hash"`
	Response         string    `json:"response"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	HitCount         int       `json:"hit_count"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// ResponseCacheStore persists and looks up cached chatbot responses.
// Entries are scoped per (chatbot, model) and expire by TTL; they are
// invalidated when a linked knowledge base document changes.
type ResponseCacheStore struct {
	database.TenantAware
}

// NewResponseCacheStore creates a new response cache store.
func NewResponseCacheStore(db *database.Connection) *ResponseCacheStore {
	return &ResponseCacheStore{TenantAware: database.TenantAware{DB: db}}
}

// HashQuestion normalizes a question and returns its SHA-256 hash, the key
// for exact-match lookups.
func HashQuestion(question string) string {
	normalized := strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(question))), " ")
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// LookupExact returns the cached response for an exact question match, or
// nil on a miss.
func (s *ResponseCacheStore) LookupExact(ctx context.Context, chatbotID, model, questionHash string) (*CachedResponse, error) {
	var entry CachedResponse
	err := s.WithTenant(ctx, func(tx pgx.Tx) error {
		query := `
			SELECT id, chatbot_id, model, question, question_hash, response,
				prompt_tokens, completion_tokens, hit_count, created_at, expires_at
			FROM ai.response_cache
			WHERE chatbot_id = $1 AND model = $2 AND question_hash = $3
			  AND expires_at > NOW()
			LIMIT 1
		`
		return tx.QueryRow(ctx, query, chatbotID, model, questionHash).Scan(
			&entry.ID, &entry.ChatbotID, &entry.Model, &entry.Question, &entry.QuestionHash,
			&entry.Response, &entry.PromptTokens, &entry.CompletionTokens,
			&entry.HitCount, &entry.CreatedAt, &entry.ExpiresAt,
		)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("response cache exact lookup failed: %w", err)
	}
	return &entry, nil
}

// LookupSemantic returns the most similar cached response above
// ResponseCacheSemanticThreshold, or nil on a miss.
func (s *ResponseCacheStore) LookupSemantic(ctx context.Context, chatbotID, model string, embedding []float32) (*CachedResponse, error) {
	if len(embedding) == 0 {
		return nil, nil
	}
	embeddingStr := formatEmbeddingLiteral(embedding)

	var entry CachedResponse
	err := s.WithTenant(ctx, func(tx pgx.Tx) error {
		query := fmt.Sprintf(`
			SELECT id, chatbot_id, model, question, question_hash, response,
				prompt_tokens, completion_tokens, hit_count, created_at, expires_at
			FROM ai.response_cache
			WHERE chatbot_id = $1 AND model = $2
			  AND embedding IS NOT NULL
			  AND expires_at > NOW()
			  AND 1 - (embedding <=> '%s'::vector) >= $3
			ORDER BY embedding <=> '%s'::vector
			LIMIT 1
		`, embeddingStr, embeddingStr)
		return tx.QueryRow(ctx, query, chatbotID, model, ResponseCacheSemanticThreshold).Scan(
			&entry.ID, &entry.ChatbotID, &entry.Model, &entry.Question, &entry.QuestionHash,
			&entry.Response, &entry.PromptTokens, &entry.CompletionTokens,
			&entry.HitCount, &entry.CreatedAt, &entry.ExpiresAt,
		)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("response cache semantic lookup failed: %w", err)
	}
	return &entry, nil
}

// Store caches a response for a chatbot. Upsert on
// (chatbot_id, model, question_hash) so a repeat question refreshes the
// entry instead of erroring on the unique constraint. Best-effort by design:
// callers log and continue on error.
func (s *ResponseCacheStore) Store(ctx context.Context, entry *CachedResponse, embedding []float32, ttl time.Duration) error {
	if entry.QuestionHash == "" {
		entry.QuestionHash = HashQuestion(entry.Question)
	}
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	now := time.Now()
	entry.CreatedAt = now
	entry.ExpiresAt = now.Add(ttl)

	return s.WithTenant(ctx, func(tx pgx.Tx) error {
		query := fmt.Sprintf(`
			INSERT INTO ai.response_cache (
				id, chatbot_id, model, question, question_hash, embedding,
				response, prompt_tokens, completion_tokens, hit_count,
				created_at, expires_at
			) VALUES ($1, $2, $3, $4, $5, '%s'::vector, $6, $7, $8, 0, $9, $10)
			ON CONFLICT (chatbot_id, model, question_hash) DO UPDATE SET
				question = EXCLUDED.question,
				embedding = EXCLUDED.embedding,
				response = EXCLUDED.response,
				prompt_tokens = EXCLUDED.prompt_tokens,
				completion_tokens = EXCLUDED.completion_tokens,
				hit_count = 0,
				created_at = EXCLUDED.created_at,
				expires_at = EXCLUDED.expires_at
		`, formatEmbeddingLiteral(embedding))

		_, err := tx.Exec(ctx, query,
			entry.ID, entry.ChatbotID, entry.Model, entry.Question, entry.QuestionHash,
			entry.Response, entry.PromptTokens, entry.CompletionTokens,
			entry.CreatedAt, entry.ExpiresAt,
		)
		return err
	})
}

// RecordHit bumps the hit counter and last_hit_at for a served entry.
func (s *ResponseCacheStore) RecordHit(ctx context.Context, id string) {
	_ = s.WithTenant(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE ai.response_cache SET hit_count = hit_count + 1, last_hit_at = NOW() WHERE id = $1`, id)
		return err
	})
}

// InvalidateForChatbot drops all cache entries for a chatbot.
func (s *ResponseCacheStore) InvalidateForChatbot(ctx context.Context, chatbotID string) error {
	return s.WithTenant(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM ai.response_cache WHERE chatbot_id = $1`, chatbotID)
		return err
	})
}

// InvalidateForKnowledgeBase drops cache entries for every chatbot linked to
// the knowledge base — called when one of its documents finishes processing
// so grounded answers never go stale.
func (s *ResponseCacheStore) InvalidateForKnowledgeBase(ctx context.Context, kbID string) error {
	return s.WithTenant(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			DELETE FROM ai.response_cache
			WHERE chatbot_id IN (
				SELECT chatbot_id FROM ai.chatbot_knowledge_bases
				WHERE knowledge_base_id = $1
			)
		`, kbID)
		return err
	})
}

// PurgeExpired deletes expired entries and returns how many were removed.
func (s *ResponseCacheStore) PurgeExpired(ctx context.Context) (int64, error) {
	var deleted int64
	err := s.WithTenant(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `DELETE FROM ai.response_cache WHERE expires_at <= NOW()`)
		if err != nil {
			return err
		}
		deleted = tag.RowsAffected()
		return nil
	})
	return deleted, err
}

// StartPurgeSweeper periodically deletes expired cache entries until ctx is
// cancelled. A full pass is cheap (indexed expires_at) even when most entries
// are live.
func (s *ResponseCacheStore) StartPurgeSweeper(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Hour
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				purgeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				if deleted, err := s.PurgeExpired(purgeCtx); err != nil {
					log.Warn().Err(err).Msg("Response cache purge failed")
				} else if deleted > 0 {
					log.Debug().Int64("deleted", deleted).Msg("Purged expired response cache entries")
				}
				cancel()
			}
		}
	}()
}
