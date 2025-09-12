package scraper

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"hianime/pkg/models"
)

// GetProducerAnimes scrapes anime list from a producer page
func (s *Scraper) GetProducerAnimes(producerName string, page int) (*models.ProducerResponse, error) {
	if producerName == "" {
		return nil, fmt.Errorf("producer name is required")
	}

	if page < 1 {
		page = 1
	}

	// URL encode the producer name
	encodedName := url.QueryEscape(producerName)
	requestURL := fmt.Sprintf("%s/producer/%s?page=%d", s.config.BaseURL, encodedName, page)

	resp, err := s.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch producer page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract producer name from page title
	realProducerName := strings.TrimSpace(doc.Find("h1.title").Text())
	if realProducerName == "" {
		realProducerName = producerName
	}

	// Extract main anime list
	animes := s.extractProducerAnimes(doc)

	// Extract top 10 animes using existing method with "today" period
	top10AnimesToday := s.extractTop10Animes(doc, "today")
	top10AnimesWeek := s.extractTop10Animes(doc, "week")
	top10AnimesMonth := s.extractTop10Animes(doc, "month")

	top10Animes := models.Top10{
		Today: top10AnimesToday,
		Week:  top10AnimesWeek,
		Month: top10AnimesMonth,
	}

	// Extract top airing animes
	topAiringAnimes := s.extractTopAiringAnimes(doc)

	// Extract pagination information
	totalPages := s.extractTotalPages(doc)
	currentPage := page // Use the page parameter passed to the function
	hasNextPage := s.extractHasNextPage(doc)

	response := &models.ProducerResponse{
		ProducerName:          realProducerName,
		Animes:                animes,
		Top10Animes:           top10Animes,
		TopAiringAnimes:       topAiringAnimes,
		TotalPages:            totalPages,
		CurrentPage:           currentPage,
		HasNextPage:           hasNextPage,
	}

	return response, nil
}

// extractProducerAnimes extracts the main anime list from the producer page
func (s *Scraper) extractProducerAnimes(doc *goquery.Document) []models.ProducerAnime {
	var animes []models.ProducerAnime

	doc.Find(".film_list-wrap .flw-item").Each(func(i int, selection *goquery.Selection) {
		// Extract anime ID from href
		href, exists := selection.Find(".film-poster a").Attr("href")
		if !exists {
			return
		}
		id := strings.TrimPrefix(href, "/")

		// Extract name
		name := strings.TrimSpace(selection.Find(".film-detail .film-name a").Text())

		// Extract poster
		poster, _ := selection.Find(".film-poster img").Attr("data-src")
		if poster == "" {
			poster, _ = selection.Find(".film-poster img").Attr("src")
		}

		// Extract duration
		duration := strings.TrimSpace(selection.Find(".film-detail .fd-infor .fdi-item.fdi-duration").Text())

		// Extract type
		animeType := strings.TrimSpace(selection.Find(".film-detail .fd-infor .fdi-item:first-child").Text())

		// Extract rating
		rating := strings.TrimSpace(selection.Find(".film-detail .fd-infor .fdi-item .imdb").Text())

		// Extract episodes info
		var episodes *models.Episodes
		subEpisodes := strings.TrimSpace(selection.Find(".film-poster .tick-sub").Text())
		dubEpisodes := strings.TrimSpace(selection.Find(".film-poster .tick-dub").Text())

		if subEpisodes != "" || dubEpisodes != "" {
			episodes = &models.Episodes{}
			if subEpisodes != "" {
				if subCount, err := strconv.Atoi(subEpisodes); err == nil {
					episodes.Sub = subCount
				}
			}
			if dubEpisodes != "" {
				if dubCount, err := strconv.Atoi(dubEpisodes); err == nil {
					episodes.Dub = dubCount
				}
			}
		}

		anime := models.ProducerAnime{
			ID:       id,
			Name:     name,
			Poster:   poster,
			Duration: duration,
			Type:     animeType,
			Rating:   rating,
			Episodes: episodes,
		}

		animes = append(animes, anime)
	})

	return animes
}

// extractTopAiringAnimes extracts top airing animes from the sidebar
func (s *Scraper) extractTopAiringAnimes(doc *goquery.Document) []models.TopAiringAnime {
	var animes []models.TopAiringAnime

	doc.Find("#top-viewed-month .anif-block-ul li").Each(func(i int, selection *goquery.Selection) {
		// Extract anime ID from href
		href, exists := selection.Find("a").Attr("href")
		if !exists {
			return
		}
		id := strings.TrimPrefix(href, "/")

		// Extract name
		name := strings.TrimSpace(selection.Find(".film-name").Text())

		// Extract poster
		poster, _ := selection.Find("img").Attr("data-src")
		if poster == "" {
			poster, _ = selection.Find("img").Attr("src")
		}

		// Extract episodes info
		var episodes *models.Episodes
		subEpisodes := strings.TrimSpace(selection.Find(".tick-sub").Text())
		dubEpisodes := strings.TrimSpace(selection.Find(".tick-dub").Text())

		if subEpisodes != "" || dubEpisodes != "" {
			episodes = &models.Episodes{}
			if subEpisodes != "" {
				if subCount, err := strconv.Atoi(subEpisodes); err == nil {
					episodes.Sub = subCount
				}
			}
			if dubEpisodes != "" {
				if dubCount, err := strconv.Atoi(dubEpisodes); err == nil {
					episodes.Dub = dubCount
				}
			}
		}

		anime := models.TopAiringAnime{
			ID:       id,
			Name:     name,
			Poster:   poster,
			Episodes: episodes,
		}

		animes = append(animes, anime)
	})

	return animes
}
