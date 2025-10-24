package models

import "encoding/json"

type CallType string

var (
	CallTypeRequest  = CallType("request")
	CallTypeResponse = CallType("response")
)

type PacketCall struct {
	Type CallType `json:"type"`
	Data any      `json:"data"`
}

type jsonPacketCall struct {
	Type CallType        `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (p *PacketCall) UnmarshalJSON(data []byte) error {
	var jp jsonPacketCall
	if err := json.Unmarshal(data, &jp); err != nil {
		return err
	}
	p.Type = jp.Type
	switch jp.Type {
	case CallTypeRequest:
		var req CallRequest
		if err := json.Unmarshal(jp.Data, &req); err != nil {
			return err
		}
		p.Data = &req
	case CallTypeResponse:
		var resp CallResponse
		if err := json.Unmarshal(jp.Data, &resp); err != nil {
			return err
		}
		p.Data = &resp
	}
	return nil
}

type CallResponse struct {
	ID       string `json:"id"`
	OK       bool   `json:"ok"`
	ErrorMsg string `json:"error,omitempty"`

	Data []string `json:"data,omitempty"`
}

type CallRequest struct {
	ID     string   `json:"id"`
	Action string   `json:"action"`
	Params []string `json:"params,omitempty"`
}
