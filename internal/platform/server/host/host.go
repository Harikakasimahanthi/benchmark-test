package host

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

const hostName = "0.0.0.0"

type WebHost struct {
	server http.Server
}

func New(port uint16, handler http.Handler) *WebHost {
	return &WebHost{
		server: http.Server{
			Addr:    fmt.Sprintf("%s:%d", hostName, port),
			Handler: handler,
		},
	}
}

func (h *WebHost) Run() {
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			const errMsg = "error running web host"
			slog.With("error", err.Error()).Error(errMsg)
			panic(errMsg)
		}
	}()
}

func (h *WebHost) Terminate(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}
