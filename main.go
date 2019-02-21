package main

import (
	"sync"

	"github.com/nerikeshi-k/lemon/hook"
	"github.com/nerikeshi-k/lemon/ws"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		// webhook
		hook.Serve()
		wg.Done()
	}()
	go func() {
		// websocket
		ws.Serve()
		wg.Done()
	}()
	wg.Wait()
}
