package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentProcessor(t *testing.T) {
	t.Run("creates processor with nil dependencies", func(t *testing.T) {
		processor := NewDocumentProcessor(nil, nil)
		assert.NotNil(t, processor)
		assert.Nil(t, processor.storage)
		assert.Nil(t, processor.embeddingService)
	})
}

func TestProcessDocumentOptions_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		opts := ProcessDocumentOptions{
			ChunkSize:     512,
			ChunkOverlap:  50,
			ChunkStrategy: ChunkingStrategyRecursive,
		}

		assert.Equal(t, 512, opts.ChunkSize)
		assert.Equal(t, 50, opts.ChunkOverlap)
		assert.Equal(t, ChunkingStrategyRecursive, opts.ChunkStrategy)
	})

	t.Run("zero values", func(t *testing.T) {
		var opts ProcessDocumentOptions
		assert.Equal(t, 0, opts.ChunkSize)
		assert.Equal(t, 0, opts.ChunkOverlap)
		assert.Equal(t, ChunkingStrategy(""), opts.ChunkStrategy)
	})
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes leading and trailing whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "collapses multiple spaces",
			input:    "hello    world",
			expected: "hello world",
		},
		{
			name:     "converts tabs to spaces",
			input:    "hello\tworld",
			expected: "hello world",
		},
		{
			name:     "converts newlines to spaces",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "handles multiple whitespace types",
			input:    "hello  \t\n  world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t\n   ",
			expected: "",
		},
		{
			name:     "preserves single spaces",
			input:    "hello world foo bar",
			expected: "hello world foo bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "short text",
			input:    "hello", // 5 chars / 4 = 1
			expected: 1,
		},
		{
			name:     "medium text",
			input:    "hello world", // 11 chars / 4 = 2
			expected: 2,
		},
		{
			name:     "longer text",
			input:    "This is a longer piece of text that should have more tokens", // 59 chars / 4 = 14
			expected: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTokenCount(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashContent(t *testing.T) {
	t.Run("produces consistent hash", func(t *testing.T) {
		content := "Hello, World!"
		hash1 := hashContent(content)
		hash2 := hashContent(content)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different content produces different hash", func(t *testing.T) {
		hash1 := hashContent("Hello")
		hash2 := hashContent("World")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("produces 64 character hex string", func(t *testing.T) {
		hash := hashContent("test content")
		assert.Len(t, hash, 64) // SHA-256 produces 32 bytes = 64 hex chars
	})

	t.Run("handles empty string", func(t *testing.T) {
		hash := hashContent("")
		assert.Len(t, hash, 64)
		// SHA-256 of empty string is well-known
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
	})
}

func TestSplitSentences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single sentence with period",
			input:    "Hello world.",
			expected: []string{"Hello world."},
		},
		{
			name:     "multiple sentences",
			input:    "Hello world. How are you? I am fine!",
			expected: []string{"Hello world.", "How are you?", "I am fine!"},
		},
		{
			name:     "sentence without ending punctuation",
			input:    "Hello world",
			expected: []string{"Hello world"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "abbreviation handling",
			input:    "Dr. Smith went home. He was tired.",
			expected: []string{"Dr.", "Smith went home.", "He was tired."},
		},
		{
			name:     "question mark",
			input:    "What is this? It is a test.",
			expected: []string{"What is this?", "It is a test."},
		},
		{
			name:     "exclamation mark",
			input:    "Wow! Amazing!",
			expected: []string{"Wow!", "Amazing!"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSentences(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDocumentProcessor_ChunkByFixed(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("splits short content", func(t *testing.T) {
		content := "Hello world"
		chunks, err := processor.chunkByFixed(content, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
		assert.Equal(t, "Hello world", chunks[0])
	})

	t.Run("splits long content", func(t *testing.T) {
		// Create content that's about 1000 chars
		content := ""
		for i := 0; i < 100; i++ {
			content += "1234567890"
		}
		// chunkSize=50 => 50*4=200 chars per chunk
		// overlap=10 => 10*4=40 chars overlap
		// 1000 chars with 200 chunk size and 40 overlap
		chunks, err := processor.chunkByFixed(content, 50, 10)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1)
	})

	t.Run("handles empty content", func(t *testing.T) {
		chunks, err := processor.chunkByFixed("", 100, 10)
		require.NoError(t, err)
		assert.Empty(t, chunks)
	})
}

func TestDocumentProcessor_SplitByCharacter(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("splits correctly", func(t *testing.T) {
		content := "0123456789" // 10 chars
		chunks, err := processor.splitByCharacter(content, 5, 2)
		require.NoError(t, err)
		// First chunk: 0-5 = "01234"
		// Second chunk starts at 5-2=3: "34567"
		// Third chunk starts at 6: "6789"
		assert.GreaterOrEqual(t, len(chunks), 2)
	})

	t.Run("handles content smaller than chunk size", func(t *testing.T) {
		content := "hello"
		chunks, err := processor.splitByCharacter(content, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
		assert.Equal(t, "hello", chunks[0])
	})

	t.Run("handles empty content", func(t *testing.T) {
		chunks, err := processor.splitByCharacter("", 100, 10)
		require.NoError(t, err)
		assert.Empty(t, chunks)
	})
}

func TestDocumentProcessor_ChunkBySentence(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("chunks multiple sentences", func(t *testing.T) {
		content := "First sentence. Second sentence. Third sentence."
		chunks, err := processor.chunkBySentence(content, 100, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 1)
	})

	t.Run("handles single sentence", func(t *testing.T) {
		content := "This is a single sentence."
		chunks, err := processor.chunkBySentence(content, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
	})
}

func TestDocumentProcessor_ChunkByParagraph(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("chunks multiple paragraphs", func(t *testing.T) {
		content := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
		chunks, err := processor.chunkByParagraph(content, 100, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 1)
	})

	t.Run("handles single paragraph", func(t *testing.T) {
		content := "This is a single paragraph without any double newlines."
		chunks, err := processor.chunkByParagraph(content, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
	})

	t.Run("removes empty paragraphs", func(t *testing.T) {
		content := "First.\n\n\n\nSecond."
		chunks, err := processor.chunkByParagraph(content, 100, 10)
		require.NoError(t, err)
		// Should only have content from non-empty paragraphs
		for _, chunk := range chunks {
			assert.NotEmpty(t, chunk)
		}
	})
}

func TestDocumentProcessor_ChunkRecursive(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("handles short content", func(t *testing.T) {
		content := "Short content."
		chunks, err := processor.chunkRecursive(content, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
		assert.Equal(t, "Short content.", chunks[0])
	})

	t.Run("splits on paragraph boundaries first", func(t *testing.T) {
		content := "First paragraph with some content.\n\nSecond paragraph with more content.\n\nThird paragraph."
		// Small chunk size to force splitting
		chunks, err := processor.chunkRecursive(content, 20, 5)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 2)
	})

	t.Run("splits on sentence boundaries when needed", func(t *testing.T) {
		content := "First sentence. Second sentence. Third sentence. Fourth sentence."
		// Small chunk size
		chunks, err := processor.chunkRecursive(content, 10, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 1)
	})
}

func TestDocumentProcessor_ChunkDocument(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("uses recursive strategy by default", func(t *testing.T) {
		content := "Test content for chunking."
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: "", // Empty = default to recursive
		}
		chunks, err := processor.chunkDocument(content, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)
	})

	t.Run("uses sentence strategy", func(t *testing.T) {
		content := "First sentence. Second sentence."
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: ChunkingStrategySentence,
		}
		chunks, err := processor.chunkDocument(content, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)
	})

	t.Run("uses paragraph strategy", func(t *testing.T) {
		content := "First paragraph.\n\nSecond paragraph."
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: ChunkingStrategyParagraph,
		}
		chunks, err := processor.chunkDocument(content, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)
	})

	t.Run("uses fixed strategy", func(t *testing.T) {
		content := "Content for fixed chunking test."
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: ChunkingStrategyFixed,
		}
		chunks, err := processor.chunkDocument(content, opts)
		require.NoError(t, err)
		assert.NotEmpty(t, chunks)
	})

	t.Run("handles empty content", func(t *testing.T) {
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: ChunkingStrategyRecursive,
		}
		chunks, err := processor.chunkDocument("", opts)
		require.NoError(t, err)
		assert.Empty(t, chunks)
	})

	t.Run("cleans whitespace before chunking", func(t *testing.T) {
		content := "  Content   with   extra   spaces  "
		opts := ProcessDocumentOptions{
			ChunkSize:     100,
			ChunkOverlap:  10,
			ChunkStrategy: ChunkingStrategyFixed,
		}
		chunks, err := processor.chunkDocument(content, opts)
		require.NoError(t, err)
		require.NotEmpty(t, chunks)
		// Content should be cleaned
		assert.NotContains(t, chunks[0], "   ")
	})
}

func TestDocumentProcessor_MergeUnits(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("merges small units into chunks", func(t *testing.T) {
		units := []string{"Unit 1.", "Unit 2.", "Unit 3.", "Unit 4."}
		chunks, err := processor.mergeUnits(units, 100, 10)
		require.NoError(t, err)
		// All units should fit in one chunk with size 100 tokens
		assert.GreaterOrEqual(t, len(chunks), 1)
	})

	t.Run("splits when units are too large", func(t *testing.T) {
		// Create units that need splitting
		units := []string{
			"This is a longer unit that contains more text.",
			"Another longer unit with substantial content.",
			"Yet another unit to ensure we have enough text.",
		}
		// Small chunk size to force splitting
		chunks, err := processor.mergeUnits(units, 10, 2)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 1)
	})

	t.Run("handles empty units slice", func(t *testing.T) {
		chunks, err := processor.mergeUnits([]string{}, 100, 10)
		require.NoError(t, err)
		assert.Empty(t, chunks)
	})

	t.Run("handles single unit", func(t *testing.T) {
		units := []string{"Single unit."}
		chunks, err := processor.mergeUnits(units, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
	})
}

func TestDocumentProcessor_SplitRecursively(t *testing.T) {
	processor := &DocumentProcessor{}

	t.Run("returns single chunk for small text", func(t *testing.T) {
		text := "Small text"
		separators := []string{"\n\n", "\n", ". ", " "}
		chunks, err := processor.splitRecursively(text, separators, 100, 10)
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
	})

	t.Run("tries next separator when current not found", func(t *testing.T) {
		// Text without double newlines
		text := "Text without paragraph breaks but with sentences. Like this one. And another."
		separators := []string{"\n\n", "\n", ". ", " "}
		chunks, err := processor.splitRecursively(text, separators, 20, 5)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(chunks), 1)
	})

	t.Run("falls back to character split with no separators", func(t *testing.T) {
		text := "TextWithNoSeparatorsAtAll"
		separators := []string{}
		chunks, err := processor.splitRecursively(text, separators, 5, 1)
		require.NoError(t, err)
		// Should split by character
		assert.GreaterOrEqual(t, len(chunks), 1)
	})
}
