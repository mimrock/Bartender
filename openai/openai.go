package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mimrock/rocketchat_openai_bot/config"
	"net/http"
	"net/url"
	"strings"
)

//var ErrorContextLengthExceeded = errors.New("context length exceeded")

type ErrorContextLengthExceeded struct {
	message string
}

func (e *ErrorContextLengthExceeded) Error() string {
	return e.message
}

func NewErrorContextLengthExceeded(msg string) error {
	return &ErrorContextLengthExceeded{msg}
}

func (e *ErrorContextLengthExceeded) Is(tgt error) bool {
	_, ok := tgt.(*ErrorContextLengthExceeded)
	if !ok {
		return false
	}
	return true
}

type OpenAI struct {
	HostName           string
	CompletionEndpoint string
	ModerationEndpoint string
	ApiToken           string
	PrePrompt          string
	Model              string
	InputModeration    bool
	OutputModeration   bool
	ModelParams        config.ModelParams
}

type HTTPError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

func NewFromConfig(config *config.Config) *OpenAI {
	oa := OpenAI{
		HostName:           config.OpenAI.HostName,
		ApiToken:           config.OpenAI.ApiToken,
		PrePrompt:          strings.TrimSpace(config.OpenAI.PrePrompt),
		Model:              config.OpenAI.Model,
		ModerationEndpoint: config.OpenAI.ModerationEndpoint,
		CompletionEndpoint: config.OpenAI.CompletionEndpoint,
		InputModeration:    config.OpenAI.InputModeration,
		OutputModeration:   config.OpenAI.OutputModeration,
		ModelParams:        config.OpenAI.ModelParams,
	}
	return &oa
}

func (o *OpenAI) CompletionURL() (string, error) {
	url, err := url.JoinPath("https://", o.HostName, o.CompletionEndpoint)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (o *OpenAI) ModerationURL() (string, error) {
	url, err := url.JoinPath("https://", o.HostName, o.ModerationEndpoint)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (o *OpenAI) Completion(cReq *CompletionRequest) (*CompletionResponse, error) {
	var cResp CompletionResponse
	url, err := o.CompletionURL()
	if err != nil {
		return nil, fmt.Errorf("cannot assemble endpoint url: %w", err)
	}
	err = o.request(url, cReq, &cResp)
	if cResp.Error.Code == "context_length_exceeded" {
		return nil, NewErrorContextLengthExceeded(cResp.Error.Message)
	} else if cResp.Error.Message != "" {
		return nil, fmt.Errorf("%w: %s ", err, cResp.Error.Message)
	}
	// All other errors
	if err != nil {
		return &cResp, fmt.Errorf("an error occured while performing the request: %w", err)
	}

	return &cResp, nil
}

func (o *OpenAI) Moderation(mReq *ModerationRequest) (*ModerationResponse, error) {
	var mResp ModerationResponse
	url, err := o.ModerationURL()
	if err != nil {
		return nil, fmt.Errorf("cannot assemble endpoint url: %w", err)
	}
	err = o.request(url, mReq, &mResp)
	if mResp.Error.Message != "" {
		return nil, fmt.Errorf("%w: %s ", err, mResp.Error.Message)
	}
	if err != nil {
		return &mResp, fmt.Errorf("an error occured during performing the request: %w", err)
	}

	if len(mResp.ID) == 0 {
		// It is dangerous to proceed if the moderation response does not look like to be legit.
		return nil, fmt.Errorf("empty moderation response")
	}

	return &mResp, nil
}

func (o *OpenAI) request(url string, request interface{}, oaResponse interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("cannot marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("cannot create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.ApiToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return parseError(resp, oaResponse)
	}

	err = json.NewDecoder(resp.Body).Decode(oaResponse)
	if err != nil {
		return fmt.Errorf("cannot parse response body: %w", err)
	}

	return nil
}

func (o *OpenAI) NewCompletionRequest(messages []Message, user string) *CompletionRequest {
	r := &CompletionRequest{
		Model:            o.Model,
		Messages:         messages,
		Temperature:      o.ModelParams.Temperature,
		TopP:             o.ModelParams.TopP,
		MaxTokens:        o.ModelParams.MaxTokens,
		PresencePenalty:  o.ModelParams.PresencePenalty,
		FrequencyPenalty: o.ModelParams.FrequencyPenalty,
	}

	if len(user) > 0 {
		r.User = &user
	}
	
	return r
}

func parseError(resp *http.Response, oaResponse interface{}) error {
	if resp.Body == nil {
		return fmt.Errorf("HTTP Error: %d", resp.StatusCode)
	}

	err := json.NewDecoder(resp.Body).Decode(oaResponse)
	if err != nil {
		return fmt.Errorf("HTTP Error: %d but response body is not available: %w", resp.StatusCode, err)
	}

	return fmt.Errorf("Non-OK status code: %d", resp.StatusCode)
}
