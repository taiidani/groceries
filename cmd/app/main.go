//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"syscall/js"
)

var (
	document js.Value
	list     js.Value
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer cancel()

	document = js.Global().Get("document")
	list = document.Call("getElementById", "list")
	events()

	fmt.Println("wasm loaded")
	<-ctx.Done()
}

func events() {
	if !list.Truthy() {
		slog.Error("Missing DOM element list")
		return
	}
}
