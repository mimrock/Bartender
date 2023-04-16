package main

import (
	"fmt"
	"github.com/mimrock/rocketchat_openai_bot/config"
	"strings"
	"time"

	"github.com/mimrock/rocketchat_openai_bot/openai"
)

type TimedMessage struct {
	openai.Message
	Timestamp time.Time
}

type History struct {
	Messages   map[string][]TimedMessage
	Size       int
	MaxLength  int
	Expiration time.Duration
}

func NewHistory() *History {
	h := new(History)
	h.Messages = make(map[string][]TimedMessage)
	return h
}

func NewHistoryFromConfig(cfg *config.Config) *History {
	h := new(History)
	h.Messages = make(map[string][]TimedMessage)
	h.Size = cfg.OpenAI.HistorySize
	h.MaxLength = cfg.OpenAI.HistoryMaxLength
	if cfg.OpenAI.MessageRetention != nil {
		h.Expiration = *cfg.OpenAI.MessageRetention
	} else {
		h.Expiration = 100 * 8765 * time.Hour // 100 years
	}
	return h
}

func (h *History) GetAsString(place string) string {
	var ret string
	if messages, ok := h.Messages[place]; ok {
		for _, m := range messages {
			ret += fmt.Sprintf("\n%s", m.Content)
		}
	}
	// @todo cut so it won't be longer than maxLength
	return strings.TrimSpace(ret)
}

func (h *History) AsOpenAIMessages(place string) []openai.Message {
	now := time.Now()
	h.Messages[place] = h.clearExpired(place, now)

	if messages, ok := h.Messages[place]; ok {
		openaiMessages := make([]openai.Message, len(messages))
		for i, m := range messages {
			openaiMessages[i] = m.Message
		}
		return openaiMessages
	}
	// @todo cut so it won't be longer than maxLength
	return []openai.Message{}
}

func (h *History) Add(place string, message openai.Message) {
	// Remove any expired messages
	now := time.Now()
	h.Messages[place] = h.clearExpired(place, now)

	timedMessage := TimedMessage{
		Message:   message,
		Timestamp: now,
	}

	if messages, ok := h.Messages[place]; ok {
		messages = append(messages, timedMessage)
		if len(messages) > h.Size {
			messages = messages[len(messages)-h.Size:]
		}
		h.Messages[place] = messages
		return
	}
	h.Messages[place] = []TimedMessage{timedMessage}
}

func (h *History) Clear(place string) {
	h.Messages[place] = []TimedMessage{}
}

// clearExpired removes any expired messages from the history.
func (h *History) clearExpired(place string, now time.Time) []TimedMessage {
	if messages, ok := h.Messages[place]; ok {
		var validMessages []TimedMessage
		for _, message := range messages {
			if now.Sub(message.Timestamp) <= h.Expiration {
				validMessages = append(validMessages, message)
			}
		}
		return validMessages
	}
	return []TimedMessage{}
}
