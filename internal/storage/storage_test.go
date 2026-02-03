package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestObject_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		now := time.Now()
		obj := Object{
			Key:          "path/to/file.txt",
			Bucket:       "my-bucket",
			Size:         1024,
			ContentType:  "text/plain",
			LastModified: now,
			ETag:         "abc123",
			Metadata: map[string]string{
				"custom-key": "custom-value",
			},
		}

		assert.Equal(t, "path/to/file.txt", obj.Key)
		assert.Equal(t, "my-bucket", obj.Bucket)
		assert.Equal(t, int64(1024), obj.Size)
		assert.Equal(t, "text/plain", obj.ContentType)
		assert.Equal(t, now, obj.LastModified)
		assert.Equal(t, "abc123", obj.ETag)
		assert.Equal(t, "custom-value", obj.Metadata["custom-key"])
	})

	t.Run("zero value has expected defaults", func(t *testing.T) {
		var obj Object
		assert.Empty(t, obj.Key)
		assert.Empty(t, obj.Bucket)
		assert.Zero(t, obj.Size)
		assert.Empty(t, obj.ContentType)
		assert.True(t, obj.LastModified.IsZero())
		assert.Nil(t, obj.Metadata)
	})
}

func TestUploadOptions_Struct(t *testing.T) {
	opts := UploadOptions{
		ContentType:     "application/pdf",
		Metadata:        map[string]string{"key": "value"},
		CacheControl:    "max-age=3600",
		ContentEncoding: "gzip",
	}

	assert.Equal(t, "application/pdf", opts.ContentType)
	assert.Equal(t, "value", opts.Metadata["key"])
	assert.Equal(t, "max-age=3600", opts.CacheControl)
	assert.Equal(t, "gzip", opts.ContentEncoding)
}

func TestDownloadOptions_Struct(t *testing.T) {
	now := time.Now()
	opts := DownloadOptions{
		IfModifiedSince:   &now,
		IfUnmodifiedSince: &now,
		IfMatch:           "etag123",
		IfNoneMatch:       "etag456",
		Range:             "bytes=0-1023",
	}

	assert.NotNil(t, opts.IfModifiedSince)
	assert.NotNil(t, opts.IfUnmodifiedSince)
	assert.Equal(t, "etag123", opts.IfMatch)
	assert.Equal(t, "etag456", opts.IfNoneMatch)
	assert.Equal(t, "bytes=0-1023", opts.Range)
}

func TestSignedURLOptions_Struct(t *testing.T) {
	opts := SignedURLOptions{
		ExpiresIn:        15 * time.Minute,
		Method:           "GET",
		ContentType:      "image/jpeg",
		TransformWidth:   800,
		TransformHeight:  600,
		TransformFormat:  "webp",
		TransformQuality: 85,
		TransformFit:     "cover",
	}

	assert.Equal(t, 15*time.Minute, opts.ExpiresIn)
	assert.Equal(t, "GET", opts.Method)
	assert.Equal(t, "image/jpeg", opts.ContentType)
	assert.Equal(t, 800, opts.TransformWidth)
	assert.Equal(t, 600, opts.TransformHeight)
	assert.Equal(t, "webp", opts.TransformFormat)
	assert.Equal(t, 85, opts.TransformQuality)
	assert.Equal(t, "cover", opts.TransformFit)
}

func TestListOptions_Struct(t *testing.T) {
	opts := ListOptions{
		Prefix:     "images/",
		MaxKeys:    100,
		Delimiter:  "/",
		StartAfter: "images/photo99.jpg",
	}

	assert.Equal(t, "images/", opts.Prefix)
	assert.Equal(t, 100, opts.MaxKeys)
	assert.Equal(t, "/", opts.Delimiter)
	assert.Equal(t, "images/photo99.jpg", opts.StartAfter)
}

func TestListResult_Struct(t *testing.T) {
	result := ListResult{
		Objects: []Object{
			{Key: "file1.txt", Size: 100},
			{Key: "file2.txt", Size: 200},
		},
		CommonPrefixes: []string{"folder1/", "folder2/"},
		IsTruncated:    true,
		NextMarker:     "file3.txt",
	}

	assert.Len(t, result.Objects, 2)
	assert.Len(t, result.CommonPrefixes, 2)
	assert.True(t, result.IsTruncated)
	assert.Equal(t, "file3.txt", result.NextMarker)
}

func TestChunkedUploadSession_Struct(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	session := ChunkedUploadSession{
		UploadID:        "upload-123",
		Bucket:          "my-bucket",
		Key:             "large-file.zip",
		TotalSize:       10 * 1024 * 1024, // 10MB
		ChunkSize:       5 * 1024 * 1024,  // 5MB
		TotalChunks:     2,
		CompletedChunks: []int{0},
		ContentType:     "application/zip",
		Metadata:        map[string]string{"uploaded-by": "user123"},
		CacheControl:    "max-age=86400",
		OwnerID:         "user123",
		S3UploadID:      "s3-upload-456",
		S3PartETags:     map[int]string{0: "etag0"},
		Status:          "in-progress",
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
	}

	assert.Equal(t, "upload-123", session.UploadID)
	assert.Equal(t, "my-bucket", session.Bucket)
	assert.Equal(t, "large-file.zip", session.Key)
	assert.Equal(t, int64(10*1024*1024), session.TotalSize)
	assert.Equal(t, int64(5*1024*1024), session.ChunkSize)
	assert.Equal(t, 2, session.TotalChunks)
	assert.Len(t, session.CompletedChunks, 1)
	assert.Equal(t, "application/zip", session.ContentType)
	assert.Equal(t, "user123", session.OwnerID)
	assert.Equal(t, "s3-upload-456", session.S3UploadID)
	assert.Equal(t, "etag0", session.S3PartETags[0])
	assert.Equal(t, "in-progress", session.Status)
}

func TestChunkResult_Struct(t *testing.T) {
	result := ChunkResult{
		ChunkIndex: 0,
		ETag:       "chunk-etag-123",
		Size:       5 * 1024 * 1024,
	}

	assert.Equal(t, 0, result.ChunkIndex)
	assert.Equal(t, "chunk-etag-123", result.ETag)
	assert.Equal(t, int64(5*1024*1024), result.Size)
}

func TestChunkedUploadSession_CompletionTracking(t *testing.T) {
	session := ChunkedUploadSession{
		TotalChunks:     5,
		CompletedChunks: []int{},
	}

	t.Run("tracks completed chunks", func(t *testing.T) {
		// Simulate completing chunks
		session.CompletedChunks = append(session.CompletedChunks, 0)
		session.CompletedChunks = append(session.CompletedChunks, 1)
		session.CompletedChunks = append(session.CompletedChunks, 2)

		assert.Len(t, session.CompletedChunks, 3)
	})

	t.Run("checks completion status", func(t *testing.T) {
		isComplete := len(session.CompletedChunks) == session.TotalChunks
		assert.False(t, isComplete)

		// Complete remaining chunks
		session.CompletedChunks = append(session.CompletedChunks, 3)
		session.CompletedChunks = append(session.CompletedChunks, 4)

		isComplete = len(session.CompletedChunks) == session.TotalChunks
		assert.True(t, isComplete)
	})
}

func TestSignedURLOptions_TransformOptions(t *testing.T) {
	t.Run("image resize options", func(t *testing.T) {
		opts := SignedURLOptions{
			ExpiresIn:       1 * time.Hour,
			Method:          "GET",
			TransformWidth:  400,
			TransformHeight: 300,
			TransformFit:    "contain",
		}

		assert.Equal(t, 400, opts.TransformWidth)
		assert.Equal(t, 300, opts.TransformHeight)
		assert.Equal(t, "contain", opts.TransformFit)
	})

	t.Run("format conversion options", func(t *testing.T) {
		opts := SignedURLOptions{
			ExpiresIn:        1 * time.Hour,
			Method:           "GET",
			TransformFormat:  "avif",
			TransformQuality: 90,
		}

		assert.Equal(t, "avif", opts.TransformFormat)
		assert.Equal(t, 90, opts.TransformQuality)
	})
}

func TestListOptions_Pagination(t *testing.T) {
	t.Run("first page", func(t *testing.T) {
		opts := ListOptions{
			Prefix:  "documents/",
			MaxKeys: 50,
		}

		assert.Empty(t, opts.StartAfter)
		assert.Equal(t, 50, opts.MaxKeys)
	})

	t.Run("next page", func(t *testing.T) {
		opts := ListOptions{
			Prefix:     "documents/",
			MaxKeys:    50,
			StartAfter: "documents/doc50.pdf",
		}

		assert.Equal(t, "documents/doc50.pdf", opts.StartAfter)
	})
}
