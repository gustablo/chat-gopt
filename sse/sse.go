package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gustablo/chat-gopt/consts"
	"github.com/tidwall/gjson"
)

const (
	EOF_TEXT = "[DONE]"
)

type MixMap = map[string]interface{}

type ChatText struct {
	data           string
	ConversationID string
	MessageID      string
	Content        string
}

type ChatStream struct {
	Stream chan *ChatText
	Err    error
}

func (c *ChatText) String() string {
	return c.data
}

func GetChatText(message string) (*ChatText, error) {
	resp, err := sendMessage(message)
	if err != nil {
		return nil, fmt.Errorf("send message failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}

	arr := strings.Split(string(body), "\n\n")

	const TEXT_ARR_MIN_LEN = 3
	const TEXT_STR_MIN_LEN = 6

	l := len(arr)

	str := arr[l-TEXT_ARR_MIN_LEN]

	text := str[TEXT_STR_MIN_LEN:]

	return parseChatText(text)
}

func parseChatText(text string) (*ChatText, error) {
	if text == "" || text == EOF_TEXT {
		return nil, fmt.Errorf("invalid chat text: %s", text)
	}

	res := gjson.Parse(text)

	conversationID := res.Get("conversation_id").String()
	messageID := res.Get("message.id").String()
	content := res.Get("message.content.parts.0").String()

	if conversationID == "" || messageID == "" {
		return nil, fmt.Errorf("invalid chat text")
	}

	return &ChatText{
		data:           text,
		ConversationID: conversationID,
		MessageID:      messageID,
		Content:        content,
	}, nil
}

func sendMessage(message string) (*http.Response, error) {
	accessToken := consts.ACCESS_TOKEN

	params := MixMap{
		"action":            "next",
		"model":             "text-davinci-002-render-sha",
		"parent_message_id": "2cb35d44-d71f-4a01-973d-7f2c711b840e",
		"messages": []MixMap{
			{
				"role": "user",
				"id":   "0d3907e7-a66b-4ee5-976e-2f8133b73696",
				"content": MixMap{
					"content_type": "text",
					"parts":        []string{message},
				},
			},
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://bypass.churchless.tech/api/conversation", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("new request failed: %v", err)
	}

	bearerToken := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Set("Authorization", bearerToken)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)

	return resp, err
}
