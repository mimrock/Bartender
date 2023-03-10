package main

import (
	"fmt"
	"strings"

	"github.com/mimrock/rocketchat_openai_bot/openai"
)

type History struct {
	Messages  map[string][]openai.Message
	Size      int
	MaxLength int
}

func NewHistory() *History {
	h := new(History)
	h.Messages = make(map[string][]openai.Message)
	return h
}

/*type Message struct {
	//Author  string
	Content string
	Role    string // user, assistant or system.
}*/

/*func (m *Message) AsString() string {
	return fmt.Sprintf("%s", m.Content)
}*/

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
	if messages, ok := h.Messages[place]; ok {
		return messages
	}
	// @todo cut so it won't be longer than maxLength
	return []openai.Message{}
}

func (h *History) Add(place string, message openai.Message) {
	if messages, ok := h.Messages[place]; ok {
		messages = append(messages, message)
		if len(messages) > h.Size {
			messages = messages[len(messages)-h.Size:]
		}
		h.Messages[place] = messages
		return
	}
	h.Messages[place] = []openai.Message{message}
}

func (h *History) Clear(place string) {
	h.Messages[place] = []openai.Message{}
}
