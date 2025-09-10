package lsp

import (
	"encoding/json"
)

// Message represents a JSON-RPC 2.0 message
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int32           `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

// ResponseError represents a JSON-RPC 2.0 error
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewRequest(id int32, method string, params any) (*Message, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return &Message{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

func NewNotification(method string, params any) (*Message, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return &Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
	}, nil
}
