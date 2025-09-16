package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ayanrajpoot10/hianime-api/config"
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
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Enable CORS if configured
	if r.config.EnableCORS {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// Log request if verbose
	if r.config.Verbose {
		log.Printf("%s %s", req.Method, req.URL.Path)
	}

	// Route the request
	r.route(w, req)
}

// route handles the actual routing logic
func (r *Router) route(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	switch {
	// Root endpoints
	case path == "/":
		r.handler.Root(w, req)
	case path == "/health":
		r.handler.Health(w, req)

	// API endpoints
	case path == "/api" || path == "/api/":
		r.handler.APIRoot(w, req)
	case path == "/api/home":
		r.handler.Homepage(w, req)
	case path == "/api/search":
		r.handler.Search(w, req)
	case path == "/api/suggestion":
		r.handler.Suggestions(w, req)
	case path == "/api/servers":
		r.handler.Servers(w, req)
	case path == "/api/stream":
		r.handler.Stream(w, req)
	case path == "/api/schedule":
		r.handler.EstimatedSchedule(w, req)
	case path == "/api/health":
		r.handler.Health(w, req)

	// Dynamic endpoints with path parameters
	case strings.HasPrefix(path, "/api/anime/"):
		r.handler.AnimeDetails(w, req)
	case strings.HasPrefix(path, "/api/qtip/"):
		r.handler.AnimeQtipInfo(w, req)
	case strings.HasPrefix(path, "/api/next-episode/"):
		r.handler.NextEpisodeSchedule(w, req)
	case strings.HasPrefix(path, "/api/episodes/"):
		r.handler.Episodes(w, req)
	case strings.HasPrefix(path, "/api/animes/"):
		r.handler.AnimeList(w, req)
	case strings.HasPrefix(path, "/api/genre/"):
		r.handler.GenreList(w, req)
	case strings.HasPrefix(path, "/api/azlist/"):
		r.handler.AZList(w, req)
	case strings.HasPrefix(path, "/api/producer/"):
		r.handler.Producer(w, req)

	// Not found
	default:
		r.handleNotFound(w, req)
	}
}

// handleNotFound handles 404 errors
func (r *Router) handleNotFound(w http.ResponseWriter, req *http.Request) {
	r.handler.writeError(w, http.StatusNotFound, fmt.Errorf("endpoint not found: %s", req.URL.Path))
}

// Start starts the HTTP server
func (r *Router) Start() error {
	address := fmt.Sprintf("%s:%s", r.config.Host, r.config.Port)

	server := &http.Server{
		Addr:         address,
		Handler:      r,
		ReadTimeout:  r.config.ReadTimeout,
		WriteTimeout: r.config.WriteTimeout,
	}

	log.Printf("🚀 Starting HiAnime Scraper API server on %s", address)
	log.Printf("📖 API documentation available at http://%s", address)

	return server.ListenAndServe()
}
