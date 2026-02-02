package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePaginationParams(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		defaultLimit  int
		maxLimit      int
		expectedLimit int
		expectedOffset int
	}{
		{
			name:          "valid limit and offset",
			limit:         50,
			offset:        100,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 50,
			expectedOffset: 100,
		},
		{
			name:          "zero limit uses default",
			limit:         0,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 20,
			expectedOffset: 0,
		},
		{
			name:          "negative limit uses default",
			limit:         -5,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 20,
			expectedOffset: 0,
		},
		{
			name:          "limit exceeds max uses default",
			limit:         150,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 20,
			expectedOffset: 0,
		},
		{
			name:          "limit at max is valid",
			limit:         100,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 100,
			expectedOffset: 0,
		},
		{
			name:          "negative offset becomes zero",
			limit:         50,
			offset:        -10,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 50,
			expectedOffset: 0,
		},
		{
			name:          "large offset is preserved",
			limit:         50,
			offset:        10000,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 50,
			expectedOffset: 10000,
		},
		{
			name:          "minimum valid limit (1)",
			limit:         1,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 1,
			expectedOffset: 0,
		},
		{
			name:          "zero offset is valid",
			limit:         50,
			offset:        0,
			defaultLimit:  20,
			maxLimit:      100,
			expectedLimit: 50,
			expectedOffset: 0,
		},
		{
			name:          "both invalid - uses defaults",
			limit:         -1,
			offset:        -1,
			defaultLimit:  25,
			maxLimit:      50,
			expectedLimit: 25,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, offset := NormalizePaginationParams(tt.limit, tt.offset, tt.defaultLimit, tt.maxLimit)
			assert.Equal(t, tt.expectedLimit, limit, "limit mismatch")
			assert.Equal(t, tt.expectedOffset, offset, "offset mismatch")
		})
	}
}

func TestNormalizePaginationParams_EdgeCases(t *testing.T) {
	t.Run("max limit of 1", func(t *testing.T) {
		limit, offset := NormalizePaginationParams(5, 0, 1, 1)
		assert.Equal(t, 1, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("default equals max", func(t *testing.T) {
		limit, offset := NormalizePaginationParams(0, 0, 100, 100)
		assert.Equal(t, 100, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("very large max limit", func(t *testing.T) {
		limit, offset := NormalizePaginationParams(1000000, 0, 100, 10000000)
		assert.Equal(t, 1000000, limit)
		assert.Equal(t, 0, offset)
	})
}
