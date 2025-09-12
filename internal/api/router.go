package api

import (
	"fmt"
	"hianime/config"
	"log"
	"net/http"
	"strings"
)

// Router handles HTTP routing for the API
type Router struct {
	handler *Handler
	config  *config.Config
}

// NewRouter creates a new router instance
func NewRouter(handler *Handler, cfg *config.Config) *Router {
	return &Router{
		handler: handler,
		config:  cfg,
	}
}

// ServeHTTP implements http.Handler interface
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS if configured
	if router.config.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// Log request if verbose
	if router.config.Verbose {
		log.Printf("%s %s", r.Method, r.URL.Path)
	}

	// Route the request
	router.route(w, r)
}

// route handles the actual routing logic
func (router *Router) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	// Root endpoints
	case path == "/":
		router.handler.Root(w, r)
	case path == "/health":
		router.handler.Health(w, r)

	// API endpoints
	case path == "/api" || path == "/api/":
		router.handler.APIRoot(w, r)
	case path == "/api/home":
		router.handler.Homepage(w, r)
	case path == "/api/search":
		router.handler.Search(w, r)
	case path == "/api/suggestion":
		router.handler.Suggestions(w, r)
	case path == "/api/servers":
		router.handler.Servers(w, r)
	case path == "/api/stream":
		router.handler.Stream(w, r)
	case path == "/api/health":
		router.handler.Health(w, r)

	// Dynamic endpoints with path parameters
	case strings.HasPrefix(path, "/api/anime/"):
		router.handler.AnimeDetails(w, r)
	case strings.HasPrefix(path, "/api/qtip/"):
		router.handler.AnimeQtipInfo(w, r)
	case strings.HasPrefix(path, "/api/episodes/"):
		router.handler.Episodes(w, r)
	case strings.HasPrefix(path, "/api/animes/"):
		router.handler.AnimeList(w, r)
	case strings.HasPrefix(path, "/api/genre/"):
		router.handler.GenreList(w, r)

	// Not found
	default:
		router.handleNotFound(w, r)
	}
}

// handleNotFound handles 404 errors
func (router *Router) handleNotFound(w http.ResponseWriter, r *http.Request) {
	router.handler.writeError(w, http.StatusNotFound, fmt.Errorf("endpoint not found: %s", r.URL.Path))
}

// Start starts the HTTP server
func (router *Router) Start() error {
	address := fmt.Sprintf("%s:%s", router.config.Host, router.config.Port)

	server := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  router.config.ReadTimeout,
		WriteTimeout: router.config.WriteTimeout,
	}

	log.Printf("🚀 Starting HiAnime Scraper API server on %s", address)
	log.Printf("📖 API documentation available at http://%s", address)

	return server.ListenAndServe()
}
