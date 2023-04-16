package main

import (
	"testing"
	"time"

	"github.com/mimrock/rocketchat_openai_bot/openai"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	// Create a new history with expiration set to 1 second.
	history := NewHistory()
	history.Expiration = 2 * time.Second
	history.Size = 4

	// Test adding messages.
	message1 := openai.Message{Role: "user", Content: "Hello"}
	message2 := openai.Message{Role: "assistant", Content: "Hi"}

	history.Add("chat1", message1)
	history.Add("chat1", message2)

	// Test GetAsString method.
	assert.Equal(t, "Hello\nHi", history.GetAsString("chat1"))

	// Test AsOpenAIMessages method.
	messages := history.AsOpenAIMessages("chat1")
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, message1, messages[0])
	assert.Equal(t, message2, messages[1])

	// Test Clear method.
	history.Clear("chat1")
	assert.Equal(t, 0, len(history.AsOpenAIMessages("chat1")))

	// Test message expiration.
	history.Add("chat1", message1)
	time.Sleep(2 * time.Second)
	history.Add("chat1", message2)
	time.Sleep(1 * time.Second)

	// Only message2 should be left in the history after expiration.
	assert.Equal(t, "Hi", history.GetAsString("chat1"))

	// Test history size.
	m3 := openai.Message{Role: "user", Content: "m3"}
	history.Add("chat1", m3)
	m4 := openai.Message{Role: "assistant", Content: "m4"}
	history.Add("chat1", m4)
	m5 := openai.Message{Role: "user", Content: "m5"}
	history.Add("chat1", m5)
	m6 := openai.Message{Role: "assistant", Content: "m6"}
	history.Add("chat1", m6)

	// message2 should be removed from the history because of size limit.
	assert.Equal(t, "m3\nm4\nm5\nm6", history.GetAsString("chat1"))
}
