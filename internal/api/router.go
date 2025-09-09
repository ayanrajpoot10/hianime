package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"hianime/config"
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

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HiAnime Scraper API</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .endpoints { margin-top: 30px; }
        .endpoint { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .method { color: #007bff; font-weight: bold; }
        .path { font-family: monospace; background: #e9ecef; padding: 2px 6px; border-radius: 3px; }
        .description { margin-top: 5px; color: #666; }
        .footer { text-align: center; margin-top: 40px; color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎌 HiAnime Scraper API</h1>
        <p>A RESTful API for scraping anime content from hianime.to</p>
        
        <div class="endpoints">
            <h2>Available Endpoints</h2>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/home</span></div>
                <div class="description">Get homepage content including spotlight, trending, and latest anime</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/search?keyword={query}&page={page}</span></div>
                <div class="description">Search for anime by keyword</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/suggestion?keyword={query}</span></div>
                <div class="description">Get search suggestions for a keyword</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/anime/{id}</span></div>
                <div class="description">Get detailed information about a specific anime</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/episodes/{id}</span></div>
                <div class="description">Get episode list for a specific anime</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/animes/{category}?page={page}</span></div>
                <div class="description">Get anime list by category (most-popular, top-airing, etc.)</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/genre/{genre}?page={page}</span></div>
                <div class="description">Get anime list by genre</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/servers?id={episodeId}</span></div>
                <div class="description">Get available servers for an episode</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/api/stream?id={episodeId}&type={sub|dub}&server={serverName}</span></div>
                <div class="description">Get streaming links for an episode</div>
            </div>
            
            <div class="endpoint">
                <div><span class="method">GET</span> <span class="path">/health</span></div>
                <div class="description">Health check endpoint</div>
            </div>
        </div>
        
        <div class="footer">
            <p>HiAnime Scraper API - Built with Go</p>
            <p><strong>Note:</strong> This is an unofficial API for educational purposes only.</p>
        </div>
    </div>
</body>
</html>`

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
