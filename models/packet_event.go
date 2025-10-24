package models

import (
	"encoding/json"
)

type EventType string

var (
	EventTypeMessage EventType = "message"
	EventTypeEvent   EventType = "event"
)

type PacketEvent struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}
type jsonPacketMessage struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (p *PacketEvent) UnmarshalJSON(data []byte) error {
	var msg jsonPacketMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	p.Type = msg.Type
	switch msg.Type {
	case EventTypeMessage:
		var message Message
		err := json.Unmarshal(msg.Data, &message)
		if err != nil {
			return err
		}
		p.Data = &message
	case EventTypeEvent:
		var event Event
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			return err
		}
		p.Data = &event
	}
	return nil
}
