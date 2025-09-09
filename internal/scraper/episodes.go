package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Episodes scrapes episode list for a specific anime
func (s *Scraper) Episodes(animeID string) (*models.EpisodesResponse, error) {
	// Rate limiting
	time.Sleep(s.config.RateLimit)

	// Extract numeric ID from anime ID
	parts := strings.Split(animeID, "-")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid anime ID format")
	}
	numericID := parts[len(parts)-1]

	url := fmt.Sprintf("%s/ajax/v2/episode/list/%s", s.config.BaseURL, numericID)

	headers := map[string]string{
		"Referer":          fmt.Sprintf("%s/watch/%s", s.config.BaseURL, animeID),
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
		Status interface{} `json:"status"` // Can be string or bool
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

	response := &models.EpisodesResponse{}

	doc.Find(".detail-infor-content .ss-list a").Each(func(i int, sel *goquery.Selection) {
		episode := models.EpisodeInfo{}

		// Extract episode number and ID
		href, exists := sel.Attr("href")
		if exists {
			// href format: /watch/anime-name?ep=123
			if strings.Contains(href, "?ep=") {
				epParts := strings.Split(href, "?ep=")
				if len(epParts) == 2 {
					episode.Episode, _ = strconv.Atoi(epParts[1])
					episode.ID = fmt.Sprintf("%s::ep=%s", animeID, epParts[1])
				}
			}
		}

		// Extract title
		episode.Title = strings.TrimSpace(sel.Find(".ssli-detail .ep-name").Text())
		if episode.Title == "" {
			episode.Title = fmt.Sprintf("Episode %d", episode.Episode)
		}

		// Check if it's a filler episode
		episode.IsFiller = sel.HasClass("ssl-item-filler")

		response.Episodes = append(response.Episodes, episode)
	})

	return response, nil
}
