package scraper

import (
	"fmt"
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

// extractAnimes extracts anime items from a generic list structure
func (s *Scraper) extractAnimes(doc *goquery.Document, selector string) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}

		// Extract ID from .dynamic-name href attribute
		dynamicName := sel.Find(".film-detail .film-name .dynamic-name")
		if href, exists := dynamicName.Attr("href"); exists {
			// Remove leading slash and split by "?ref=search"
			cleanHref := strings.TrimPrefix(href, "/")
			parts := strings.Split(cleanHref, "?ref=search")
			if len(parts) > 0 {
				item.ID = parts[0]
			}
		}

		// Extract title and alternative title
		item.Title = strings.TrimSpace(dynamicName.Text())
		if jname, exists := dynamicName.Attr("data-jname"); exists {
			item.AlternativeTitle = strings.TrimSpace(jname)
		}

		// Extract poster
		if poster, exists := sel.Find(".film-poster .film-poster-img").Attr("data-src"); exists {
			item.Poster = strings.TrimSpace(poster)
		}

		// Extract duration
		item.Duration = strings.TrimSpace(sel.Find(".film-detail .fd-infor .fdi-item.fdi-duration").Text())

		// Extract type (first fdi-item)
		item.Type = strings.TrimSpace(sel.Find(".film-detail .fd-infor .fdi-item:nth-of-type(1)").Text())

		// Extract rating
		if rating := strings.TrimSpace(sel.Find(".film-poster .tick-rate").Text()); rating != "" {
			item.Rating = rating
		}

		// Initialize Episodes to avoid nil pointer dereference
		item.Episodes = &models.Episodes{}

		// Extract episode counts from film-poster ticks
		subText := strings.TrimSpace(sel.Find(".film-poster .tick-sub").Text())
		dubText := strings.TrimSpace(sel.Find(".film-poster .tick-dub").Text())

		// Parse sub episodes (get last part after splitting by space)
		if subText != "" {
			parts := strings.Fields(subText)
			if len(parts) > 0 {
				if subCount, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					item.Episodes.Sub = subCount
				}
			}
		}

		// Parse dub episodes (get last part after splitting by space)
		if dubText != "" {
			parts := strings.Fields(dubText)
			if len(parts) > 0 {
				if dubCount, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					item.Episodes.Dub = dubCount
				}
			}
		}

		items = append(items, item)
	})

	return items
}

// extractTop10Animes extracts anime items from top 10 ranking lists
func (s *Scraper) extractTop10Animes(doc *goquery.Document, period string) []models.AnimeItem {
	var items []models.AnimeItem
	selector := fmt.Sprintf("#top-viewed-%s ul li", period)

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}

		// Extract ID from .dynamic-name href attribute
		dynamicName := sel.Find(".film-detail .dynamic-name")
		if href, exists := dynamicName.Attr("href"); exists {
			// Remove leading slash
			item.ID = strings.TrimPrefix(strings.TrimSpace(href), "/")
		}

		// Extract rank from .film-number span
		if rankText := strings.TrimSpace(sel.Find(".film-number span").Text()); rankText != "" {
			if rank, err := strconv.Atoi(rankText); err == nil {
				item.Rank = rank
			}
		}

		// Extract title and alternative title
		item.Title = strings.TrimSpace(dynamicName.Text())
		if jname, exists := dynamicName.Attr("data-jname"); exists {
			item.AlternativeTitle = strings.TrimSpace(jname)
		}

		// Extract poster
		if poster, exists := sel.Find(".film-poster .film-poster-img").Attr("data-src"); exists {
			item.Poster = strings.TrimSpace(poster)
		}

		// Initialize Episodes to avoid nil pointer dereference
		item.Episodes = &models.Episodes{}

		// Extract episode counts from .fd-infor .tick-item
		subText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-item.tick-sub").Text())
		dubText := strings.TrimSpace(sel.Find(".film-detail .fd-infor .tick-item.tick-dub").Text())

		if subText != "" {
			if subCount, err := strconv.Atoi(subText); err == nil {
				item.Episodes.Sub = subCount
			}
		}

		if dubText != "" {
			if dubCount, err := strconv.Atoi(dubText); err == nil {
				item.Episodes.Dub = dubCount
			}
		}

		items = append(items, item)
	})

	return items
}

// extractMostPopularAnimes extracts anime items from most popular sections
func (s *Scraper) extractMostPopularAnimes(doc *goquery.Document, selector string) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}

		// Extract ID from .dynamic-name href attribute
		dynamicName := sel.Find(".film-detail .dynamic-name")
		if href, exists := dynamicName.Attr("href"); exists {
			// Remove leading slash
			item.ID = strings.TrimPrefix(strings.TrimSpace(href), "/")
		}

		// Extract title
		item.Title = strings.TrimSpace(dynamicName.Text())

		// Extract alternative title - check both possible selectors
		if jname, exists := sel.Find(".film-detail .film-name .dynamic-name").Attr("data-jname"); exists {
			item.AlternativeTitle = strings.TrimSpace(jname)
		} else if jname, exists := dynamicName.Attr("data-jname"); exists {
			item.AlternativeTitle = strings.TrimSpace(jname)
		}

		// Extract poster
		if poster, exists := sel.Find(".film-poster .film-poster-img").Attr("data-src"); exists {
			item.Poster = strings.TrimSpace(poster)
		}

		// Initialize Episodes to avoid nil pointer dereference
		item.Episodes = &models.Episodes{}

		// Extract episode counts from .fd-infor .tick
		subText := strings.TrimSpace(sel.Find(".fd-infor .tick .tick-sub").Text())
		dubText := strings.TrimSpace(sel.Find(".fd-infor .tick .tick-dub").Text())

		if subText != "" {
			if subCount, err := strconv.Atoi(subText); err == nil {
				item.Episodes.Sub = subCount
			}
		}

		if dubText != "" {
			if dubCount, err := strconv.Atoi(dubText); err == nil {
				item.Episodes.Dub = dubCount
			}
		}

		// Extract type from .fd-infor .tick text (get last word after cleaning)
		tickText := strings.TrimSpace(sel.Find(".fd-infor .tick").Text())
		if tickText != "" {
			// Replace multiple whitespace/newlines with single space
			cleanText := strings.Join(strings.Fields(tickText), " ")
			parts := strings.Fields(cleanText)
			if len(parts) > 0 {
				item.Type = parts[len(parts)-1]
			}
		}

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
