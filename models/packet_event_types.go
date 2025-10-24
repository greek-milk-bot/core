package models

import "time"

type Message struct {
	ID    string       `json:"id"`
	Owner *GuildMember `json:"user"`

	MsgType string `json:"type"`
	Guild   *Guild `json:"guild"`

	Quote   *Message  `json:"quote,omitempty"`
	Content Contents  `json:"content"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Event struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data,omitempty"`
}
