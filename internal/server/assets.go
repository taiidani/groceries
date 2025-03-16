package server

import (
	"embed"
	"log/slog"
	"net/http"
	"path/filepath"
)

//go:embed assets
var assets embed.FS

func (s *Server) assetsHandler(resp http.ResponseWriter, req *http.Request) {
	// Special case for apple-touch-icon, which must live in the root
	switch req.URL.Path {
	case "/apple-touch-icon.png":
		slog.Debug("Serving apple touch icon", "path", req.URL.Path)
		http.ServeFileFS(resp, req, assets, filepath.Join("assets", req.URL.Path))
	default:
		slog.Debug("Serving file", "path", req.URL.Path)
		if DevMode {
			http.ServeFile(resp, req, filepath.Join("internal", "server", req.URL.Path))
		} else {
			http.ServeFileFS(resp, req, assets, req.URL.Path)
		}
	}
}
