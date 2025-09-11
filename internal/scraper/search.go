package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Search performs search for anime based on keyword
func (s *Scraper) Search(keyword string, page int) (*models.SearchResponse, error) {
	if page < 1 {
		page = 1
	}

	// Rate limiting
	time.Sleep(s.config.RateLimit)

	url := fmt.Sprintf("%s/search?keyword=%s&page=%d", s.config.BaseURL, strings.ReplaceAll(keyword, " ", "+"), page)

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

	response := &models.SearchResponse{
		CurrentPage: page,
	}

	// Extract search results
	response.Results = s.extractGenericAnimeList(doc, ".film_list .film_list-wrap .flw-item")

	// Check if there's a next page
	response.HasNextPage = doc.Find(".pagination .next").Length() > 0

	return response, nil
}

// Suggestions scrapes search suggestions based on keyword
func (s *Scraper) Suggestions(keyword string) (*models.SearchResponse, error) {
	// Rate limiting
	time.Sleep(s.config.RateLimit)

	url := fmt.Sprintf("%s/ajax/search/suggest?keyword=%s", s.config.BaseURL, strings.ReplaceAll(keyword, " ", "+"))

	headers := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
	}

	resp, err := s.client.GetWithHeaders(url, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ajaxResp struct {
		Status any         `json:"status"` // Can be string or bool
		HTML   string      `json:"html"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// Check status - handle both string and boolean values
	var statusOK bool
	switch v := ajaxResp.Status.(type) {
	case string:
		statusOK = v == "success"
	case bool:
		statusOK = v
	default:
		return nil, fmt.Errorf("unexpected status type: %T", ajaxResp.Status)
	}

	if !statusOK {
		return nil, fmt.Errorf("API returned error status")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ajaxResp.HTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	response := &models.SearchResponse{
		CurrentPage: 1,
		HasNextPage: false,
	}

	// Extract suggestions
	doc.Find(".nav-item").Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}

		// Extract ID and title
		linkEl := sel.Find("a")
		href, exists := linkEl.Attr("href")
		if exists {
			parts := strings.Split(href, "/")
			if len(parts) > 0 {
				item.ID = parts[len(parts)-1]
			}
		}

		item.Title = strings.TrimSpace(linkEl.Find(".srp-detail .film-name").Text())
		item.Poster, _ = linkEl.Find(".film-poster img").Attr("data-src")

		// Extract type and year
		item.Type = strings.TrimSpace(linkEl.Find(".srp-detail .film-infor span").First().Text())

		response.Results = append(response.Results, item)
	})

	return response, nil
}
