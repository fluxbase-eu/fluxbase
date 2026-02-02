package pubsub

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelConstants(t *testing.T) {
	t.Run("BroadcastChannel has expected value", func(t *testing.T) {
		assert.Equal(t, "fluxbase:broadcast", BroadcastChannel)
	})

	t.Run("PresenceChannel has expected value", func(t *testing.T) {
		assert.Equal(t, "fluxbase:presence", PresenceChannel)
	})

	t.Run("SchemaCacheChannel has expected value", func(t *testing.T) {
		assert.Equal(t, "fluxbase:schema_cache", SchemaCacheChannel)
	})

	t.Run("all channels have fluxbase prefix", func(t *testing.T) {
		channels := []string{BroadcastChannel, PresenceChannel, SchemaCacheChannel}
		for _, ch := range channels {
			assert.Contains(t, ch, "fluxbase:", "Channel %s should have fluxbase: prefix", ch)
		}
	})

	t.Run("channels are unique", func(t *testing.T) {
		channels := map[string]bool{
			BroadcastChannel:    true,
			PresenceChannel:     true,
			SchemaCacheChannel:  true,
		}
		assert.Equal(t, 3, len(channels), "All channels should be unique")
	})
}

func TestMessage(t *testing.T) {
	t.Run("Message struct fields", func(t *testing.T) {
		msg := Message{
			Channel: "test-channel",
			Payload: []byte("test payload"),
		}

		assert.Equal(t, "test-channel", msg.Channel)
		assert.Equal(t, []byte("test payload"), msg.Payload)
	})

	t.Run("Message JSON serialization", func(t *testing.T) {
		msg := Message{
			Channel: "test-channel",
			Payload: []byte("test payload"),
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		var decoded Message
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, msg.Channel, decoded.Channel)
		assert.Equal(t, msg.Payload, decoded.Payload)
	})

	t.Run("Message with empty payload", func(t *testing.T) {
		msg := Message{
			Channel: "test-channel",
			Payload: []byte{},
		}

		assert.Equal(t, "test-channel", msg.Channel)
		assert.Empty(t, msg.Payload)
	})

	t.Run("Message with nil payload", func(t *testing.T) {
		msg := Message{
			Channel: "test-channel",
			Payload: nil,
		}

		assert.Equal(t, "test-channel", msg.Channel)
		assert.Nil(t, msg.Payload)
	})

	t.Run("Message with binary payload", func(t *testing.T) {
		binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		msg := Message{
			Channel: "binary-channel",
			Payload: binaryData,
		}

		assert.Equal(t, binaryData, msg.Payload)
	})

	t.Run("Message with JSON payload", func(t *testing.T) {
		jsonPayload := []byte(`{"event":"insert","table":"users","data":{"id":1}}`)
		msg := Message{
			Channel: BroadcastChannel,
			Payload: jsonPayload,
		}

		assert.Equal(t, BroadcastChannel, msg.Channel)

		// Verify payload is valid JSON
		var parsed map[string]interface{}
		err := json.Unmarshal(msg.Payload, &parsed)
		require.NoError(t, err)
		assert.Equal(t, "insert", parsed["event"])
		assert.Equal(t, "users", parsed["table"])
	})

	t.Run("Message with large payload", func(t *testing.T) {
		// 1MB payload
		largePayload := make([]byte, 1024*1024)
		for i := range largePayload {
			largePayload[i] = byte(i % 256)
		}

		msg := Message{
			Channel: "large-channel",
			Payload: largePayload,
		}

		assert.Equal(t, 1024*1024, len(msg.Payload))
	})
}

func TestMessage_JSONTags(t *testing.T) {
	t.Run("JSON tags are correctly applied", func(t *testing.T) {
		msg := Message{
			Channel: "my-channel",
			Payload: []byte("data"),
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		// Verify JSON field names
		var raw map[string]interface{}
		err = json.Unmarshal(data, &raw)
		require.NoError(t, err)

		_, hasChannel := raw["channel"]
		_, hasPayload := raw["payload"]

		assert.True(t, hasChannel, "JSON should have 'channel' field")
		assert.True(t, hasPayload, "JSON should have 'payload' field")
	})
}
