package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// GetAnimeQtipInfo scrapes anime qtip information by ID
func (s *Scraper) GetAnimeQtipInfo(animeID string) (*models.QtipResponse, error) {
	// Validate anime ID format
	animeID = strings.TrimSpace(animeID)
	if animeID == "" || !strings.Contains(animeID, "-") {
		return nil, fmt.Errorf("invalid anime id: %s", animeID)
	}

	// Extract the numeric ID from the anime ID (last part after splitting by "-")
	parts := strings.Split(animeID, "-")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid anime id format: %s", animeID)
	}
	id := parts[len(parts)-1]

	// Construct the qtip URL
	qtipURL := fmt.Sprintf("/ajax/movie/qtip/%s", id)
	url := s.config.BaseURL + qtipURL

	// Make the HTTP request with proper headers
	resp, err := s.client.GetWithHeaders(url, map[string]string{
		"Referer":          s.config.BaseURL,
		"X-Requested-With": "XMLHttpRequest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch qtip info: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Initialize the response
	response := &models.QtipResponse{
		Anime: models.QtipAnime{
			ID:       animeID,
			Episodes: &models.Episodes{},
			Genres:   []string{},
		},
	}

	// Main selector for qtip content
	selector := ".pre-qtip-content"
	qtipContent := doc.Find(selector)

	if qtipContent.Length() == 0 {
		return nil, fmt.Errorf("qtip content not found")
	}

	// Extract ID from the play button href
	playButton := qtipContent.Find(".pre-qtip-button a.btn-play")
	if href, exists := playButton.Attr("href"); exists {
		parts := strings.Split(strings.TrimSpace(href), "/")
		if len(parts) > 0 {
			response.Anime.ID = parts[len(parts)-1]
		}
	}

	// Extract title
	if title := strings.TrimSpace(qtipContent.Find(".pre-qtip-title").Text()); title != "" {
		response.Anime.Name = title
	}

	// Extract MAL score (first child of pre-qtip-detail)
	detailFirst := qtipContent.Find(".pre-qtip-detail").Children().First()
	if malScore := strings.TrimSpace(detailFirst.Text()); malScore != "" {
		response.Anime.MalScore = malScore
	}

	// Extract quality
	if quality := strings.TrimSpace(qtipContent.Find(".tick .tick-quality").Text()); quality != "" {
		response.Anime.Quality = quality
	}

	// Extract type
	if animeType := strings.TrimSpace(qtipContent.Find(".badge.badge-quality").Text()); animeType != "" {
		response.Anime.Type = animeType
	}

	// Extract episode counts
	if subText := strings.TrimSpace(qtipContent.Find(".tick .tick-sub").Text()); subText != "" {
		if subCount, err := strconv.Atoi(subText); err == nil {
			response.Anime.Episodes.Sub = subCount
		}
	}

	if dubText := strings.TrimSpace(qtipContent.Find(".tick .tick-dub").Text()); dubText != "" {
		if dubCount, err := strconv.Atoi(dubText); err == nil {
			response.Anime.Episodes.Dub = dubCount
		}
	}

	// Extract description
	if description := strings.TrimSpace(qtipContent.Find(".pre-qtip-description").Text()); description != "" {
		response.Anime.Description = description
	}

	// Extract additional details from .pre-qtip-line elements
	qtipContent.Find(".pre-qtip-line").Each(func(i int, sel *goquery.Selection) {
		// Get the key from .stick element (remove trailing colon and convert to lowercase)
		keyText := strings.TrimSpace(sel.Find(".stick").Text())
		if keyText == "" {
			return
		}

		// Remove trailing colon
		if strings.HasSuffix(keyText, ":") {
			keyText = keyText[:len(keyText)-1]
		}
		key := strings.ToLower(keyText)

		var value string
		if key != "genres" {
			// For non-genres, get value from .stick-text
			value = strings.TrimSpace(sel.Find(".stick-text").Text())
		} else {
			// For genres, get all text after the key
			fullText := strings.TrimSpace(sel.Text())
			if len(fullText) > len(keyText)+1 {
				value = strings.TrimSpace(fullText[len(keyText)+1:])
			}
		}

		if value == "" {
			return
		}

		// Set values based on key
		switch key {
		case "japanese":
			response.Anime.JName = value
		case "synonyms":
			response.Anime.Synonyms = value
		case "aired":
			response.Anime.Aired = value
		case "status":
			response.Anime.Status = value
		case "genres":
			// Split genres by comma and trim each
			genreList := strings.Split(value, ",")
			for _, genre := range genreList {
				if trimmedGenre := strings.TrimSpace(genre); trimmedGenre != "" {
					response.Anime.Genres = append(response.Anime.Genres, trimmedGenre)
				}
			}
		}
	})

	return response, nil
}
