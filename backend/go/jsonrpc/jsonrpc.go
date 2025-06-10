package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Hub struct {
	conn    net.Conn
	reader  *bufio.Reader
	pending map[string]chan json.RawMessage
	mu      sync.Mutex
}

type WorkspaceFolders struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

func MustNewHub(addrs string, workspaceFolders []WorkspaceFolders) (*Hub, func() error) {
	conn, err := net.Dial("tcp", addrs)
	if err != nil {
		panic(err)
	}

	hub := &Hub{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		mu:      sync.Mutex{},
		pending: make(map[string]chan json.RawMessage),
	}

	go hub.listen()

	return hub, func() error {
		return conn.Close()
	}
}

func (h *Hub) Call(payload BaseJsonRpc) error {
	reqBody, err := payload.Marshal()
	if err != nil {
		return err
	}

	if payload.ID != "" {
		ch := make(chan json.RawMessage, 1)
		h.mu.Lock()
		h.pending[payload.ID] = ch
		h.mu.Unlock()
	}

	_, err = h.conn.Write(reqBody)
	if err != nil {
		return err
	}

	return nil
}

func (h *Hub) WaitResponseCtx(ctx context.Context, id string) (json.RawMessage, error) {
	h.mu.Lock()
	ch, ok := h.pending[id]
	h.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no pending call for id %s", id)
	}

	select {
	case data := <-ch:
		h.mu.Lock()
		delete(h.pending, id)
		h.mu.Unlock()
		return data, nil
	case <-ctx.Done():
		h.mu.Lock()
		delete(h.pending, id)
		h.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (h *Hub) WaitResponseWithMarshalCtx(ctx context.Context, id string, output any) error {
	rawMsg, err := h.WaitResponseCtx(ctx, id)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawMsg, output)
	if err != nil {
		return fmt.Errorf("%w, failed marshal message: %s", err, rawMsg)
	}

	return nil
}

func (h *Hub) listen() {
	for {
		headers := make(map[string]string)
		for {
			line, err := h.reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading header:", err)
				return
			}

			line = strings.TrimRight(line, "\r\n")

			if line == "" {
				break
			}

			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		contentLengthStr := headers["Content-Length"]
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			fmt.Println("Invalid content length:", err)
			continue
		}

		respBody := make([]byte, contentLength)
		_, err = io.ReadFull(h.reader, respBody)
		if err != nil {
			fmt.Println("Failed to read body:", err)
			continue
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(respBody, &raw); err != nil {
			fmt.Println("Failed to unmarshal body:", err)
			continue
		}

		idRaw, ok := raw["id"]
		if !ok {
			fmt.Println("Notification or event:", string(respBody))
			continue
		}

		var id string

		if err := json.Unmarshal(idRaw, &id); err != nil {
			fmt.Println("Failed to parse ID:", err)
			continue
		}

		h.mu.Lock()
		ch, exists := h.pending[id]
		h.mu.Unlock()

		if exists {
			ch <- respBody
		} else {
			fmt.Println("⚠️ Unexpected response ID:", id)
		}
	}
}
