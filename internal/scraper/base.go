package scraper

import (
	"strconv"
	"strings"

	"hianime/config"
	"hianime/pkg/httpclient"
	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Scraper handles all scraping operations
type Scraper struct {
	config *config.Config
	client *httpclient.Client
}

// New creates a new scraper instance
func New(cfg *config.Config) *Scraper {
	clientCfg := httpclient.Config{
		Timeout:   cfg.Timeout,
		UserAgent: cfg.UserAgent,
		BaseURL:   cfg.BaseURL,
		Retries:   cfg.MaxRetries,
	}

	return &Scraper{
		config: cfg,
		client: httpclient.New(clientCfg),
	}
}

// extractGenericAnimeList extracts anime items from a generic list structure
func (s *Scraper) extractGenericAnimeList(doc *goquery.Document, selector string) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}
		item.Rank = i + 1

		// Extract poster and ID
		posterEl := sel.Find(".film-poster")
		item.Poster, _ = posterEl.Find("img").Attr("data-src")

		href, exists := posterEl.Attr("href")
		if exists {
			parts := strings.Split(href, "/")
			if len(parts) > 0 {
				item.ID = parts[len(parts)-1]
			}
		}

		// Extract title and alternative title
		titleEl := sel.Find(".film-detail .film-name a")
		item.Title = strings.TrimSpace(titleEl.Text())
		item.AlternativeTitle, _ = titleEl.Attr("data-jname")

		// Extract type and episodes
		item.Type = strings.TrimSpace(sel.Find(".film-detail .fd-infor .fdi-item:first-child").Text())

		// Extract episode counts
		subText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-sub").Text())
		dubText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-dub").Text())
		epsText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-eps").Text())

		item.Episodes.Sub, _ = strconv.Atoi(subText)
		item.Episodes.Dub, _ = strconv.Atoi(dubText)
		item.Episodes.Eps, _ = strconv.Atoi(epsText)

		items = append(items, item)
	})

	return items
}

// extractRankingList extracts anime items from ranking lists (top 10)
func (s *Scraper) extractRankingList(doc *goquery.Document, selector string) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}
		item.Rank = i + 1

		// Extract poster and ID
		posterEl := sel.Find(".film-poster")
		item.Poster, _ = posterEl.Find("img").Attr("data-src")

		href, exists := posterEl.Attr("href")
		if exists {
			parts := strings.Split(href, "/")
			if len(parts) > 0 {
				item.ID = parts[len(parts)-1]
			}
		}

		// Extract title
		item.Title = strings.TrimSpace(sel.Find(".film-detail .film-name").Text())
		item.AlternativeTitle, _ = sel.Find(".film-detail .film-name").Attr("data-jname")

		// Extract episode counts
		subText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-sub").Text())
		dubText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-dub").Text())
		epsText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-eps").Text())

		item.Episodes.Sub, _ = strconv.Atoi(subText)
		item.Episodes.Dub, _ = strconv.Atoi(dubText)
		item.Episodes.Eps, _ = strconv.Atoi(epsText)

		items = append(items, item)
	})

	return items
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
