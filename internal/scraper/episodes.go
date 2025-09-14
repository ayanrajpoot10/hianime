package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ayanrajpoot10/hianime-api/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Episodes scrapes episode list for a specific anime
func (s *Scraper) Episodes(animeID string) (*models.EpisodesResponse, error) {
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
		Status     any    `json:"status"`
		HTML       string `json:"html"`
		TotalItems int    `json:"totalItems"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// Handle status
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

	response := &models.EpisodesResponse{
		TotalItems: ajaxResp.TotalItems,
}

	doc.Find(".ss-list a.ssl-item.ep-item").Each(func(i int, sel *goquery.Selection) {
		episode := models.EpisodeInfo{}

		// Extract episode number
		if epNumStr, exists := sel.Attr("data-number"); exists {
			episode.Episode, _ = strconv.Atoi(epNumStr)
		}

		// Extract episode ID
		if epID, exists := sel.Attr("data-id"); exists {
			episode.ID = fmt.Sprintf("%s::ep=%s", animeID, epID)
		}

		// Extract episode URL
		if href, exists := sel.Attr("href"); exists {
			episode.URL = fmt.Sprintf("%s%s", s.config.BaseURL, href)
		}

		// Title and JName
		titleSel := sel.Find(".ep-name.e-dynamic-name")
		episode.Title = strings.TrimSpace(titleSel.AttrOr("title", ""))
		episode.JName = strings.TrimSpace(titleSel.AttrOr("data-jname", ""))

		if episode.Title == "" {
			episode.Title = fmt.Sprintf("Episode %d", episode.Episode)
		}

		// Is it a filler episode?
		episode.IsFiller = sel.HasClass("ssl-item-filler")

		response.Episodes = append(response.Episodes, episode)
	})

	return response, nil
}
