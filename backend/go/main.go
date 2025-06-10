package main

import (
	"context"
	"flag"
	"go-exec-lsp/go/jsonrpc"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	hub, closeFunc := jsonrpc.MustNewHub("127.0.0.1:3737", []jsonrpc.WorkspaceFolders{
		{
			Uri:  "file:///Users/ibanrama-master/Documents/Developments/OpenSourceProject/ExecLSP/backend/go/sample",
			Name: "sample",
		},
	})
	wsHub := NewHub(hub)
	go wsHub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	go func() {
		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	<-ctx.Done()

	if err := closeFunc(); err != nil {
		panic(err)
	}
}
