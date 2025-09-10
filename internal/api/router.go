package api

import (
	_ "embed"
	"fmt"
	"hianime/config"
	"log"
	"net/http"
	"strings"
)

//go:embed templates/index.html
var htmlTemplate string

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
		router.handleRoot(w, r)
	case path == "/health":
		router.handler.Health(w, r)

	// API endpoints
	case path == "/api" || path == "/api/":
		router.handleAPIRoot(w, r)
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

// handleRoot handles requests to the root path
func (router *Router) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := htmlTemplate

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// handleAPIRoot handles requests to the API root path
func (router *Router) handleAPIRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"name":        "HiAnime Scraper API",
		"description": "A RESTful API for scraping anime content from hianime.to",
		"endpoints": map[string]string{
			"homepage":    "/api/home",
			"search":      "/api/search?keyword={query}&page={page}",
			"suggestions": "/api/suggestion?keyword={query}",
			"anime":       "/api/anime/{id}",
			"episodes":    "/api/episodes/{id}",
			"anime_list":  "/api/animes/{category}?page={page}",
			"genre_list":  "/api/genre/{genre}?page={page}",
			"servers":     "/api/servers?id={episodeId}",
			"stream":      "/api/stream?id={episodeId}&type={sub|dub}&server={serverName}",
			"health":      "/api/health",
		},
		"categories": []string{
			"most-popular", "top-airing", "most-favorite", "completed",
			"recently-added", "recently-updated", "top-upcoming",
			"subbed-anime", "dubbed-anime", "movie", "tv", "ova", "ona", "special", "events",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	router.handler.writeJSON(w, http.StatusOK, response)
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
