package route

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	router *http.ServeMux
}

func NewRouter() *Router {
	router := http.NewServeMux()
	return &Router{
		router: router,
	}
}

func (r *Router) WithMetrics() *Router {
	r.router.Handle("/metrics", promhttp.Handler())
	return r
}

func (r *Router) Router() *http.ServeMux {
	return r.router
}
