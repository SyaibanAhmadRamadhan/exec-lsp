package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-exec-lsp/go/jsonrpc"
	"log"
)

type Hub struct {
	clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client

	jsonrpchub *jsonrpc.Hub
}

func NewHub(jsonRcpHub *jsonrpc.Hub) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		jsonrpchub: jsonRcpHub,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			h.clients[reg] = true
		case unreg := <-h.unregister:
			delete(h.clients, unreg)
			close(unreg.send)
		case message := <-h.broadcast:
			jsonrpcMsg := jsonrpc.BaseJsonRpc{}
			err := json.Unmarshal(message, &jsonrpcMsg)
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}

			err = h.jsonrpchub.Call(jsonrpcMsg)
			if err != nil {
				log.Printf("ERROR: %v", err)
				continue
			}

			if jsonrpcMsg.ID != "" {
				go func() {
					rawResp, err := h.jsonrpchub.WaitResponseCtx(context.Background(), jsonrpcMsg.ID)
					if err != nil {
						log.Printf("ERROR: %v", err)
						return
					}

					fmt.Printf("Resp (%s):\n%+v\n\n", jsonrpcMsg.Method, string(rawResp))
					for client := range h.clients {
						select {
						case client.send <- rawResp:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}()
			}

		}
	}
}
