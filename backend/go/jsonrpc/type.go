package jsonrpc

import (
	"encoding/json"
	"fmt"
)

type LSPMethod string

const (
	Initialize  LSPMethod = "initialize"
	Initialized LSPMethod = "initialized"
)

type BaseJsonRpc struct {
	Method         string `json:"method"`
	ID             string `json:"id,omitempty"`
	JsonrpcVersion string `json:"jsonrpc"`
	Params         any
}

func (base BaseJsonRpc) Marshal() ([]byte, error) {
	b, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(b), b)

	return []byte(msg), nil
}
