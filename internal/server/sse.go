package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// sseEventList is triggered when an item added to or removed from the list
	sseEventList = "list"

	// sseEventCart is triggered when an item added to or removed from the cart
	sseEventCart = "cart"

	// sseEventCategory is triggered when a category is added or removed
	sseEventCategory = "category"
)

type sseServer struct {
	clients map[int]chan<- sseEvent
	m       sync.Mutex
}

func newSSEServer() *sseServer {
	return &sseServer{
		clients: map[int]chan<- sseEvent{},
		m:       sync.Mutex{},
	}
}

func (srv *sseServer) addClient(feed chan<- sseEvent) int {
	srv.m.Lock()
	defer srv.m.Unlock()

	l := len(srv.clients)
	srv.clients[l] = feed
	return l
}

func (srv *sseServer) removeClient(l int) {
	srv.m.Lock()
	defer srv.m.Unlock()

	delete(srv.clients, l)
}

func (srv *sseServer) announce(events ...string) {
	for _, evt := range events {
		srv.broadcast(sseEvent{event: evt})
	}
}

func (srv *sseServer) broadcast(evt sseEvent) {
	slog.Info("broadcasting event", "event", evt.event)
	for _, feed := range srv.clients {
		feed <- evt
	}
}

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")

	feed := make(chan sseEvent, 1)
	clientID := s.sseServer.addClient(feed)
	defer s.sseServer.removeClient(clientID)

	ping := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-s.ctx.Done():
			evt := sseEvent{event: "close"}
			slog.Info("sending sse close directive")
			evt.Write(w)
			return
		case <-r.Context().Done():
			slog.Info("SSE client disconnected")
			return
		case evt := <-feed:
			slog.InfoContext(r.Context(), "sending sse", "event", evt.event)
			evt.Write(w)
		case <-ping.C:
			evt := sseEvent{event: "ping", data: time.Now().String()}
			slog.DebugContext(r.Context(), "sending sse ping")
			evt.Write(w)
		}
	}
}

type sseEvent struct {
	event string
	data  string
}

func (e *sseEvent) Write(w http.ResponseWriter) error {
	if len(e.event) > 0 {
		fmt.Fprint(w, "event: "+e.event+"\n")
	}

	if len(e.data) > 0 {
		// Place each data line with its own prefix
		// This is to avoid newlines in the data from ending the message early
		datas := strings.Split(e.data, "\n")
		fmt.Fprint(w, "data: "+strings.Join(datas, "\ndata: ")+"\n")
	} else {
		// Data MUST always be present to trigger events
		fmt.Fprint(w, "data: \n")
	}

	fmt.Fprint(w, "\n")

	f, ok := w.(http.Flusher)
	if !ok {
		return errors.New("client does not support sse")
	}

	f.Flush()
	return nil
}
