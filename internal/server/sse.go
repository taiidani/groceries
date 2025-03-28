package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/taiidani/groceries/internal/events"
)

const (
	// sseEventList is triggered when an item added to or removed from the list
	sseEventList = "list"

	// sseEventCart is triggered when an item added to or removed from the cart
	sseEventCart = "cart"

	// sseEventCategory is triggered when a category is added or removed
	sseEventCategory = "category"
)

func (s *Server) sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")

	sub := s.sseServer.Subscribe(r.Context(),
		sseEventCart,
		sseEventList,
		sseEventCategory,
	)

	ping := time.NewTicker(time.Second * 2)
	for {
		select {
		case <-s.ctx.Done():
			evt := events.Event{Event: "close"}
			slog.Info("sending sse close directive")
			evt.Write(w)
			return
		case <-r.Context().Done():
			slog.Info("SSE client disconnected")
			return
		case evt := <-sub:
			slog.InfoContext(r.Context(), "sending sse", "event", evt.Event)
			evt.Write(w)
		case <-ping.C:
			evt := events.Event{Event: "ping", Data: time.Now()}
			slog.DebugContext(r.Context(), "sending sse ping")
			evt.Write(w)
		}
	}
}
