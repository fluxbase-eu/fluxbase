package ai

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/testutil/dbbootstrap"
)

// DB-backed integration tests for the ai package. These run only outside
// -short mode and require a PostgreSQL server (the pgvector image) reachable
// via FLUXBASE_DATABASE_* — the same contract as the internal/auth tests and
// CI's test-go-db job. When the server is unreachable (plain dev machine),
// the DB-backed tests skip instead of failing.

var (
	aiTestDB     *database.Connection
	aiTestDBDown bool // set when the database is unreachable; DB tests skip
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	cfg := &config.DatabaseConfig{
		Host:          aiTestEnv("FLUXBASE_DATABASE_HOST", "postgres"),
		Port:          aiTestEnvInt("FLUXBASE_DATABASE_PORT", 5432),
		User:          aiTestEnv("FLUXBASE_DATABASE_USER", "postgres"),
		Password:      aiTestEnv("FLUXBASE_DATABASE_PASSWORD", "postgres"),
		AdminUser:     aiTestEnv("FLUXBASE_DATABASE_ADMIN_USER", "postgres"),
		AdminPassword: aiTestEnv("FLUXBASE_DATABASE_ADMIN_PASSWORD", "postgres"),
		Database:      aiTestEnv("FLUXBASE_TEST_DATABASE", "fluxbase_test"),
		SSLMode:       aiTestEnv("FLUXBASE_DATABASE_SSLMODE", "disable"),

		MaxConnections:  10,
		MinConnections:  2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
		HealthCheck:     time.Minute,
	}

	// Same contract as the internal/auth tests: SKIP_MIGRATIONS=true (CI's
	// test-go-db job) means the declarative schema is already applied and the
	// bootstrap must not drop schemas out from under the other packages.
	if os.Getenv("SKIP_MIGRATIONS") == "" {
		if err := dbbootstrap.Ensure(context.Background(), dbbootstrap.Config{
			Host:          cfg.Host,
			Port:          cfg.Port,
			Database:      cfg.Database,
			User:          cfg.User,
			Password:      cfg.Password,
			AdminUser:     cfg.AdminUser,
			AdminPassword: cfg.AdminPassword,
		}); err != nil {
			fmt.Printf("SKIP ai DB-backed tests: database unreachable (%v)\n", err)
			aiTestDBDown = true
			os.Exit(m.Run())
		}
	}

	conn, err := database.NewConnection(*cfg)
	if err != nil {
		fmt.Printf("SKIP ai DB-backed tests: connect as app user failed (%v)\n", err)
		aiTestDBDown = true
		os.Exit(m.Run())
	}
	aiTestDB = conn

	code := m.Run()
	conn.Close()
	os.Exit(code)
}

// requireDB skips a test when the integration database is unavailable.
func requireDB(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("DB-backed test skipped in -short mode")
	}
	if aiTestDBDown || aiTestDB == nil {
		t.Skip("integration database unavailable")
	}
}

func aiTestEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func aiTestEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}

// aiTestCtx returns a fresh context with a generous timeout for one test.
func aiTestCtx(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// unitVector3 builds a 1536-dim embedding whose direction is the 3-dim unit
// vector (x, y, z) placed in the first three components, so cosine
// similarities between test vectors are exact dot products of the inputs.
func unitVector3(x, y, z float64) []float32 {
	norm := math.Sqrt(x*x + y*y + z*z)
	v := make([]float32, SupportedEmbeddingDimension)
	v[0] = float32(x / norm)
	v[1] = float32(y / norm)
	v[2] = float32(z / norm)
	return v
}

// createTestKB provisions an enabled knowledge base with a unique name and
// registers cleanup.
func createTestKB(t *testing.T, ctx context.Context, storage *KnowledgeBaseStorage) *KnowledgeBase {
	t.Helper()
	kb := &KnowledgeBase{
		Name:          "itest-" + uuid.New().String()[:8],
		Namespace:     "test",
		Enabled:       true,
		Source:        "api",
		Visibility:    KBVisibilityPrivate,
		RerankEnabled: true, // match the column default (explicit INSERT would otherwise write false)
	}
	require.NoError(t, storage.CreateKnowledgeBase(ctx, kb))
	t.Cleanup(func() {
		_ = storage.DeleteKnowledgeBase(context.Background(), kb.ID)
	})
	return kb
}

// createTestDocument inserts a document owned by userID (may be "" for a
// global document) in the KB and registers cleanup.
func createTestDocument(t *testing.T, ctx context.Context, storage *KnowledgeBaseStorage, kbID, content, userID string) *Document {
	t.Helper()
	doc := &Document{
		KnowledgeBaseID: kbID,
		Title:           "itest doc",
		Content:         content,
		SourceType:      "manual",
		ContentHash:     hashContent(content),
	}
	if userID != "" {
		doc.Metadata = []byte(fmt.Sprintf(`{"user_id": %q}`, userID))
	}
	require.NoError(t, storage.CreateDocument(ctx, doc))
	t.Cleanup(func() {
		_ = storage.DeleteDocument(context.Background(), doc.ID)
	})
	return doc
}

// createTestChunk inserts one chunk with the given content and embedding.
func createTestChunk(t *testing.T, ctx context.Context, storage *KnowledgeBaseStorage, kbID, docID, content string, index int, embedding []float32) Chunk {
	t.Helper()
	chunk := Chunk{
		DocumentID:      docID,
		KnowledgeBaseID: kbID,
		Content:         content,
		ChunkIndex:      index,
		Embedding:       embedding,
	}
	require.NoError(t, storage.CreateChunks(ctx, []Chunk{chunk}))
	return chunk
}
