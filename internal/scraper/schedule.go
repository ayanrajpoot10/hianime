package scraper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// GetEstimatedSchedule scrapes estimated schedule for a specific date
func (s *Scraper) GetEstimatedSchedule(date string, tzOffset int) (*models.EstimatedScheduleResponse, error) {
	if s.config.Verbose {
		fmt.Printf("Fetching estimated schedule for date: %s (timezone offset: %d)\n", date, tzOffset)
	}

	// Validate date format (YYYY-MM-DD)
	date = strings.TrimSpace(date)
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if date == "" || !datePattern.MatchString(date) {
		return nil, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", date)
	}

	// Validate timezone offset
	if tzOffset == 0 {
		tzOffset = -330 // Default timezone offset (India Standard Time)
	}

	// Construct the schedule URL
	scheduleURL := fmt.Sprintf("/ajax/schedule/list?tzOffset=%d&date=%s", tzOffset, date)
	url := s.config.BaseURL + scheduleURL

	if s.config.Verbose {
		fmt.Printf("Making request to: %s\n", url)
	}

	// Make the HTTP request with proper headers
	resp, err := s.client.GetWithHeaders(url, map[string]string{
		"Accept":           "*/*",
		"Referer":          s.config.BaseURL,
		"X-Requested-With": "XMLHttpRequest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response to extract HTML
	var jsonResp struct {
		HTML string `json:"html"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Parse the HTML content
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(jsonResp.HTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Initialize the response
	response := &models.EstimatedScheduleResponse{
		ScheduledAnimes: []models.ScheduledAnime{},
	}

	// Check if there's no data
	if strings.Contains(doc.Text(), "No data to display") {
		return response, nil
	}

	// Extract scheduled animes from li elements
	doc.Find("li").Each(func(i int, sel *goquery.Selection) {
		anime := models.ScheduledAnime{}

		// Extract anime ID from href attribute
		link := sel.Find("a")
		if href, exists := link.Attr("href"); exists {
			// Remove leading slash and trim
			anime.ID = strings.TrimPrefix(strings.TrimSpace(href), "/")
		}

		// Extract time
		timeText := strings.TrimSpace(sel.Find("a .time").Text())
		anime.Time = timeText

		// Extract name
		nameText := strings.TrimSpace(sel.Find("a .film-name.dynamic-name").Text())
		anime.Name = nameText

		// Extract Japanese name
		nameElement := sel.Find("a .film-name.dynamic-name")
		if jname, exists := nameElement.Attr("data-jname"); exists {
			anime.JName = strings.TrimSpace(jname)
		}

		// Calculate airing timestamp
		if timeText != "" {
			// Create timestamp from date and time (assuming format HH:MM)
			timestampStr := fmt.Sprintf("%sT%s:00", date, timeText)
			if airingTime, err := time.Parse("2006-01-02T15:04:05", timestampStr); err == nil {
				anime.AiringTimestamp = airingTime.UnixMilli()
				// Calculate seconds until airing
				now := time.Now().UnixMilli()
				anime.SecondsUntilAiring = (anime.AiringTimestamp - now) / 1000
			}
		}

		// Extract episode number
		episodeButton := sel.Find("a .fd-play button")
		episodeText := strings.TrimSpace(episodeButton.Text())
		if episodeText != "" {
			// Parse episode number from text like "EP 1" or "Episode 1"
			parts := strings.Fields(episodeText)
			if len(parts) >= 2 {
				if episodeNum, err := strconv.Atoi(parts[1]); err == nil {
					anime.Episode = episodeNum
				}
			}
		}

		// Only add if we have at least an ID
		if anime.ID != "" {
			response.ScheduledAnimes = append(response.ScheduledAnimes, anime)
		}
	})

	return response, nil
}

// GetNextEpisodeSchedule scrapes the next episode schedule for a specific anime
func (s *Scraper) GetNextEpisodeSchedule(animeID string) (*models.NextEpisodeScheduleResponse, error) {
	if s.config.Verbose {
		fmt.Printf("Fetching next episode schedule for anime: %s\n", animeID)
	}

	// Validate anime ID format
	animeID = strings.TrimSpace(animeID)
	if animeID == "" || !strings.Contains(animeID, "-") {
		return nil, fmt.Errorf("invalid anime id: %s", animeID)
	}

	// Construct the anime watch URL
	animeURL := fmt.Sprintf("/watch/%s", animeID)
	url := s.config.BaseURL + animeURL

	if s.config.Verbose {
		fmt.Printf("Making request to: %s\n", url)
	}

	// Make the HTTP request with proper headers
	resp, err := s.client.GetWithHeaders(url, map[string]string{
		"Accept":  "*/*",
		"Referer": s.config.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch anime page: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Initialize the response
	response := &models.NextEpisodeScheduleResponse{}

	// Extract timestamp from the schedule alert
	selector := ".schedule-alert > .alert.small > span:last-child"
	scheduleSpan := doc.Find(selector)

	if scheduleSpan.Length() > 0 {
		if timestamp, exists := scheduleSpan.Attr("data-value"); exists {
			timestamp = strings.TrimSpace(timestamp)
			if timestamp != "" {
				// Parse the timestamp
				if schedule, err := time.Parse(time.RFC3339, timestamp); err == nil {
					response.AiringISOTimestamp = schedule.Format(time.RFC3339)
					airingTimestamp := schedule.UnixMilli()
					response.AiringTimestamp = &airingTimestamp

					// Calculate seconds until airing
					now := time.Now().UnixMilli()
					secondsUntilAiring := (airingTimestamp - now) / 1000
					response.SecondsUntilAiring = &secondsUntilAiring
				}
			}
		}
	}

	return response, nil
}
