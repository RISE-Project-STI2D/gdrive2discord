package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type StatusCode int

const (
	Ok StatusCode = iota
	CannotConnect
	CannotDeserialize
	InvalidWebhookToken
	UnknownWebhook
	MsgTooLong
	NoText
	RateLimited
	NotAuthed
	InvalidAuth
	TokenRevoked
	AccountInactive
	UserIsBot
	UnknownError
)

func (e StatusCode) String() string {
	return statusCodes[e]
}
func (e StatusCode) Error() string {
	return e.String()
}

func NewStatusCodeFromError(ec string) StatusCode {
	v, ok := errorLabelToStatusCode[ec]
	if ok {
		return v
	}
	return UnknownError
}

var errorLabelToStatusCode = map[string]StatusCode{
	"Invalid Webhook Token": InvalidWebhookToken,
	"Unknown Webhook":       UnknownWebhook,
	"msg_too_long":          MsgTooLong,
	"no_text":               NoText,
	"rate_limited":          RateLimited,
	"not_authed":            NotAuthed,
	"invalid_auth":          InvalidAuth,
	"token_revoked":         TokenRevoked,
	"account_inactive":      AccountInactive,
	"user_is_bot":           UserIsBot,
}

var statusCodes = []string{
	Ok:                  "ok",
	CannotConnect:       "cannot_connect",
	CannotDeserialize:   "cannot_deserialize",
	InvalidWebhookToken: "Invalid Webhook Token",
	UnknownWebhook:      "Unknown Webhook",
	MsgTooLong:          "msg_too_long",
	NoText:              "no_text",
	RateLimited:         "rate_limited",
	NotAuthed:           "not_authed",
	InvalidAuth:         "invalid_auth",
	TokenRevoked:        "token_revoked",
	AccountInactive:     "account_inactive",
	UserIsBot:           "user_is_bot",
	UnknownError:        "unknown_error",
}

// Attachment Discord Description and fields
type Attachment struct {
	Fallback string  `json:"description"`
	Fields   []Field `json:"fields"`
}

// Field Discord Field
type Field struct {
	Title string `json:"name"`
	Value string `json:"value"`
	Short bool   `json:"inline"`
}

// Message Old Slack message format
type Message struct {
	Username    string       `json:"username"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	IconUrl     string       `json:"icon_url"`
}

// discordMessage Represent the structure of a discord message webhook, used to convert from the slack format to the discord format
type discordMessage struct {
	Username    string       `json:"username"`
	Text        string       `json:"content"`
	Attachments []Attachment `json:"embeds"`
	IconUrl     string       `json:"avatar_url"`
}

// WebhookInfo Structure of a webhook, informations given by the discord api
type WebhookInfo struct {
	Name      string `json:"name"`
	ChannelID string `json:"channel_id"`
	Avatar    string `json:"avatar"`
	GuildID   string `json:"guild_id"`
	ID        string `json:"id"`
}

type webhookInfoResponse struct {
	Code  *int   `json:"code,omitempty"`
	Error string `json:"message,omitempty"`
	*WebhookInfo
}

// GetWebhookInfo Check if webhook is valid and get some infos about it
func GetWebhookInfo(client *http.Client, webhookURL string) (*WebhookInfo, StatusCode, error) {
	response, err := client.Get(webhookURL)
	if err != nil {
		return nil, CannotConnect, err
	}
	defer response.Body.Close()
	var self = new(webhookInfoResponse)
	err = json.NewDecoder(response.Body).Decode(self)
	if err != nil {
		return nil, CannotDeserialize, err
	}
	if self.Code != nil {
		return nil, NewStatusCodeFromError(self.Error), errors.New(self.Error)
	}
	return self.WebhookInfo, Ok, nil
}

func PostMessage(client *http.Client, webhookURL string, message *Message) (StatusCode, error) {

	payload := discordMessage{
		Username:    message.Username,
		Text:        message.Text,
		IconUrl:     message.IconUrl,
		Attachments: message.Attachments}
	jsonValue, _ := json.Marshal(payload)
	fmt.Printf(string(jsonValue))
	//"attachments": {string(payload)},
	response, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return CannotConnect, err
	}
	defer response.Body.Close()
	return Ok, nil
}
