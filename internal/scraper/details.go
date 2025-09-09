package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// AnimeDetails scrapes detailed information about a specific anime
func (s *Scraper) AnimeDetails(animeID string) (*models.AnimeDetailResponse, error) {
	// Rate limiting
	time.Sleep(s.config.RateLimit)

	url := fmt.Sprintf("%s/%s", s.config.BaseURL, animeID)

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

	detail := &models.AnimeDetailResponse{}

	// Basic information
	detail.ID = animeID
	detail.Title = strings.TrimSpace(doc.Find(".anisc-info .film-name").Text())
	detail.AlternativeTitle, _ = doc.Find(".anisc-info .film-name").Attr("data-jname")
	detail.Poster, _ = doc.Find(".anisc-poster .film-poster-img").Attr("src")

	// Description
	detail.Description = strings.TrimSpace(doc.Find(".anisc-info .film-description .text").Text())
	detail.Synopsis = detail.Description

	// Extract details from info items
	doc.Find(".anisc-info .item").Each(func(i int, sel *goquery.Selection) {
		label := strings.TrimSpace(sel.Find(".item-head").Text())
		value := strings.TrimSpace(sel.Find(".name").Text())

		switch strings.ToLower(label) {
		case "type:":
			detail.Type = value
		case "status:":
			detail.Status = value
		case "aired:":
			detail.Aired = value
		case "duration:":
			detail.Duration = value
		case "quality:":
			detail.Quality = value
		case "rating:":
			detail.Rating = value
		case "scored:":
			detail.Scored = value
		case "source:":
			detail.Source = value
		case "premiered:":
			detail.PremiereDate = value
		case "studios:":
			studios := []string{}
			sel.Find(".name a").Each(func(i int, studio *goquery.Selection) {
				if studio := strings.TrimSpace(studio.Text()); studio != "" {
					studios = append(studios, studio)
				}
			})
			detail.Studios = studios
		case "producers:":
			producers := []string{}
			sel.Find(".name a").Each(func(i int, producer *goquery.Selection) {
				if producer := strings.TrimSpace(producer.Text()); producer != "" {
					producers = append(producers, producer)
				}
			})
			detail.Producers = producers
		case "genres:":
			genres := []string{}
			sel.Find(".name a").Each(func(i int, genre *goquery.Selection) {
				if genre := strings.TrimSpace(genre.Text()); genre != "" {
					genres = append(genres, genre)
				}
			})
			detail.Genres = genres
		}
	})

	// Extract episode information
	subText := strings.TrimSpace(doc.Find(".anisc-info .tick-sub").Text())
	dubText := strings.TrimSpace(doc.Find(".anisc-info .tick-dub").Text())
	epsText := strings.TrimSpace(doc.Find(".anisc-info .tick-eps").Text())

	detail.Episodes.Sub, _ = strconv.Atoi(subText)
	detail.Episodes.Dub, _ = strconv.Atoi(dubText)
	detail.Episodes.Eps, _ = strconv.Atoi(epsText)

	// Extract related animes
	detail.RelatedAnimes = s.extractGenericAnimeList(doc, ".block_area .tab-content .flw-item")

	// Extract recommended animes
	detail.RecommendedAnimes = s.extractGenericAnimeList(doc, ".block_area:contains('You might also like') .flw-item")

	return detail, nil
}
