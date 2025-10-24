package models

import "encoding/json"

type PacketType string

var (
	PacketTypeEvent = PacketType("event") // 消息
	PacketTypeCall  = PacketType("call")  // 控制
	PacketTypeMeta  = PacketType("meta")  // 元数据控制
)

type Packet struct {
	Src  string     `json:"src"`
	Dest string     `json:"dest"`
	Type PacketType `json:"type"`
	Data any        `json:"data"`
}

type WithSrcPacket[T any] struct {
	Src  string `json:"src"`  // src id
	Data T      `json:"data"` // data
}

type jsonPacket struct {
	Src  string          `json:"src"`
	Dest string          `json:"dest"`
	Type PacketType      `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (p *Packet) UnmarshalJSON(data []byte) error {
	var jp jsonPacket
	if err := json.Unmarshal(data, &jp); err != nil {
		return err
	}
	p.Src = jp.Src
	p.Dest = jp.Dest
	p.Type = jp.Type
	switch p.Type {
	case PacketTypeEvent:
		var e Event
		if err := json.Unmarshal(jp.Data, &e); err != nil {
			return err
		}
		p.Data = &e
	case PacketTypeCall:
		var c PacketCall
		if err := json.Unmarshal(jp.Data, &c); err != nil {
			return err
		}
		p.Data = &c
	case PacketTypeMeta:
		var m PacketMeta
		if err := json.Unmarshal(jp.Data, &m); err != nil {
			return err
		}
		p.Data = &m
	}
	return nil
}
