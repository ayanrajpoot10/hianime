package api

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"

	"hianime/internal/scraper"
	"hianime/pkg/models"
)

//go:embed templates/index.html
var htmlTemplate string

// Handler holds the scraper instance and handles HTTP requests
type Handler struct {
	scraper *scraper.Scraper
}

// NewHandler creates a new API handler
func NewHandler(s *scraper.Scraper) *Handler {
	return &Handler{
		scraper: s,
	}
}

// writeJSON writes a JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.APIResponse{
		Success: statusCode < 400,
		Data:    data,
	}

	if statusCode >= 400 {
		if err, ok := data.(error); ok {
			response.Error = err.Error()
			response.Data = nil
		} else if msg, ok := data.(string); ok {
			response.Error = msg
			response.Data = nil
		}
	}

	json.NewEncoder(w).Encode(response)
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, statusCode int, err error) {
	h.writeJSON(w, statusCode, err.Error())
}

// Homepage handles GET /api/home
func (h *Handler) Homepage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	data, err := h.scraper.Homepage()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// AnimeDetails handles GET /api/anime/{id}
func (h *Handler) AnimeDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	// Extract anime ID from URL path
	path := r.URL.Path
	animeID := path[len("/api/anime/"):]

	if animeID == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	data, err := h.scraper.AnimeDetails(animeID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Search handles GET /api/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	query := r.URL.Query()
	keyword := query.Get("keyword")
	if keyword == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	data, err := h.scraper.Search(keyword, page)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Suggestions handles GET /api/suggestion
func (h *Handler) Suggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	query := r.URL.Query()
	keyword := query.Get("keyword")
	if keyword == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	data, err := h.scraper.Suggestions(keyword)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Episodes handles GET /api/episodes/{id}
func (h *Handler) Episodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	// Extract anime ID from URL path
	path := r.URL.Path
	animeID := path[len("/api/episodes/"):]

	if animeID == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	data, err := h.scraper.Episodes(animeID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Servers handles GET /api/servers
func (h *Handler) Servers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	query := r.URL.Query()
	episodeID := query.Get("id")
	if episodeID == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	data, err := h.scraper.Servers(episodeID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Stream handles GET /api/stream
func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	query := r.URL.Query()
	episodeID := query.Get("id")
	if episodeID == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	serverType := query.Get("type")
	if serverType == "" {
		serverType = "sub"
	}

	serverName := query.Get("server")
	if serverName == "" {
		serverName = "HD-1"
	}

	data, err := h.scraper.StreamLinks(episodeID, serverType, serverName)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// AnimeList handles GET /api/animes/{category}
func (h *Handler) AnimeList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	// Extract category from URL path
	path := r.URL.Path
	category := path[len("/api/animes/"):]

	if category == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	query := r.URL.Query()
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	data, err := h.scraper.AnimeList(category, page)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// GenreList handles GET /api/genre/{genre}
func (h *Handler) GenreList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	// Extract genre from URL path
	path := r.URL.Path
	genre := path[len("/api/genre/"):]

	if genre == "" {
		h.writeError(w, http.StatusBadRequest, http.ErrMissingFile)
		return
	}

	query := r.URL.Query()
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	data, err := h.scraper.GenreList(genre, page)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSON(w, http.StatusOK, data)
}

// Health handles GET /api/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"message": "hianime API is running",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Root handles requests to the root path
func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	html := htmlTemplate

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// APIRoot handles requests to the API root path
func (h *Handler) APIRoot(w http.ResponseWriter, r *http.Request) {
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
	h.writeJSON(w, http.StatusOK, response)
}
