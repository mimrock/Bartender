package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/mimrock/rocketchat_openai_bot/openai"
	"github.com/mimrock/rocketchat_openai_bot/rocket"

	log "github.com/sirupsen/logrus"
)

func DemoResponse(msg rocket.Message) {
	reply, err := msg.Reply(fmt.Sprintf("@%s %s %d", msg.UserName, msg.GetNotAddressedText(), 0))

	// If no error replying, take the reply and edit it to count to 10 asynchronously
	if err == nil {
		go func() {
			msg.SetIsTyping(true)
			for i := 1; i <= 10; i++ {
				time.Sleep(time.Second)
				reply.EditText(fmt.Sprintf("@%s %s %d", msg.UserName, msg.GetNotAddressedText(), i))
			}
			msg.SetIsTyping(false)
		}()
	}

	// React to the initial message
	msg.React(":grinning:")
}

func OpenAIResponse(rocketmsg rocket.Message, oa *openai.OpenAI, hist *History) error {
	msg := openai.Message{
		Role:    "user",
		Content: rocketmsg.GetNotAddressedText(),
	}
	rocketmsg.SetIsTyping(true)
	defer func() {
		rocketmsg.SetIsTyping(false)
	}()

	place := rocketmsg.RoomName

	if oa.InputModeration {
		// Send the input to the OpenAI moderation endpoint, and if it is flagged, return an error instead of sending anything to the completion endpoint.
		mresp, err := oa.Moderation(&openai.ModerationRequest{
			Input: rocketmsg.GetNotAddressedText(),
		})
		if err != nil {
			return fmt.Errorf("cannot perform perliminary request to the moderation endpoint: %w", err)
		}

		log.WithField("moderationResponse", mresp).Debug("Preliminary (input) moderation response.")

		if mresp.IsFlagged() {
			// @todo configurable message?
			_, err = rocketmsg.Reply(fmt.Sprintf("@%s :triangular_flag_on_post: Our bot uses OpenAI's moderation system, which flagged your message as inappropriate. Please try rephrasing your message to avoid any offensive or inappropriate content. REASON: %s :triangular_flag_on_post:",
				rocketmsg.UserName, mresp.FlaggedReason()))
			if err != nil {
				return fmt.Errorf("cannot send reply to rocketchat: %w", err)
			}
			return nil
		}
	}

	var systemMessage = openai.Message{
		Role:    "system",
		Content: oa.PrePrompt,
	}

	// Prepend the preprompt
	var messages []openai.Message
	if len(oa.PrePrompt) > 0 {
		messages = append(messages, systemMessage)
	}
	messages = append(messages, hist.AsOpenAIMessages(place)...)

	messages = append(messages, msg)
	
	cresp, err := oa.Completion(oa.NewCompletionRequest(messages, ""))
	if err != nil {
		if errors.Is(err, &openai.ErrorContextLengthExceeded{}) {
			// If the reason for the error is context_length_exceeded, we clear history, so it does not happen on the next comment.
			hist.Clear(place)
			log.Debug("To prevent context_length_exceeded error to happen repeatedly, the history has been cleared.")
		}
		return fmt.Errorf("cannot perform completion request: %w", err)
	}

	if len(cresp.Choices) == 0 {
		return fmt.Errorf("no choices returned")
	}

	log.WithField("completionResponse", cresp).Trace("Completion response.")

	var response string
	var mresp *openai.ModerationResponse
	if oa.OutputModeration {
		mresp, err = oa.Moderation(&openai.ModerationRequest{
			Input: cresp.Choices[0].Message.Content,
		})
		if err != nil {
			return fmt.Errorf("cannot perform follow-up request to the moderation endpoint (output check): %w", err)
		}

		if mresp.IsFlagged() {
			// @todo better explanation that it is the output that got flagged.
			response = fmt.Sprintf(":triangular_flag_on_post: (output flagged: %s) :triangular_flag_on_post:", mresp.FlaggedReason())
		}

		log.WithField("moderationResponse", mresp).Trace("Follow-up (output) moderation response.")

	}

	response += cresp.Choices[0].Message.Content

	// @todo further calls if finishReason indicates that the response is not completed.
	_, err = rocketmsg.Reply(fmt.Sprintf("@%s %s", rocketmsg.UserName, response))
	if err != nil {
		return fmt.Errorf("cannot send reply to rocketchat: %w", err)
	}

	if mresp == nil || !mresp.IsFlagged() {
		hist.Add(place, msg)
		hist.Add(place, openai.Message{
			Role:    "assistant",
			Content: cresp.Choices[0].Message.Content,
		})
	}

	return nil
}
