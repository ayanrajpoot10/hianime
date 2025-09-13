package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ayanrajpoot10/hianime-api/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// AnimeList scrapes anime list by category (most-popular, top-airing, etc.)
func (s *Scraper) AnimeList(category string, page int) (*models.ListPageResponse, error) {
	if page < 1 {
		page = 1
	}

	// Rate limiting
	time.Sleep(s.config.RateLimit)

	var url string
	switch category {
	case "most-popular", "top-airing", "most-favorite", "completed", "recently-added", "recently-updated", "top-upcoming":
		url = fmt.Sprintf("%s/%s?page=%d", s.config.BaseURL, category, page)
	case "subbed-anime", "dubbed-anime", "movie", "tv", "ova", "ona", "special", "events":
		url = fmt.Sprintf("%s/%s?page=%d", s.config.BaseURL, category, page)
	default:
		return nil, fmt.Errorf("unsupported category: %s", category)
	}

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	response := &models.ListPageResponse{
		CurrentPage: page,
		Category:    category,
	}

	// Extract anime list
	response.Results = s.extractAnimes(doc, ".film_list .film_list-wrap .flw-item")

	// Check if there's a next page
	response.HasNextPage = doc.Find(".pagination .next").Length() > 0

	return response, nil
}

// GenreList scrapes anime list by genre
func (s *Scraper) GenreList(genre string, page int) (*models.ListPageResponse, error) {
	if page < 1 {
		page = 1
	}

	// Rate limiting
	time.Sleep(s.config.RateLimit)

	url := fmt.Sprintf("%s/genre/%s?page=%d", s.config.BaseURL, strings.ToLower(genre), page)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	response := &models.ListPageResponse{
		CurrentPage: page,
		Category:    fmt.Sprintf("genre:%s", genre),
	}

	// Extract anime list
	response.Results = s.extractAnimes(doc, ".film_list .film_list-wrap .flw-item")

	// Check if there's a next page
	response.HasNextPage = doc.Find(".pagination .next").Length() > 0

	return response, nil
}
