package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Homepage scrapes the homepage content including spotlight, trending, etc.
func (s *Scraper) Homepage() (*models.HomepageResponse, error) {
	resp, err := s.client.Get(s.config.BaseURL + "/home")
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

	response := &models.HomepageResponse{}

	// Extract spotlight anime
	response.Spotlight = s.extractSpotlight(doc)

	// Extract trending anime
	response.Trending = s.extractTrending(doc)

	// Extract latest completed
	response.LatestCompleted = s.extractLatestCompleted(doc)

	// Extract top airing
	response.TopAiring = s.extractTopAiring(doc)

	// Extract most popular
	response.MostPopular = s.extractMostPopular(doc)

	// Extract most favorite
	response.MostFavorite = s.extractMostFavorite(doc)

	// Extract recently added
	response.RecentlyAdded = s.extractRecentlyAdded(doc)

	// Extract latest updated
	response.LatestUpdated = s.extractLatestUpdated(doc)

	// Extract top upcoming
	response.TopUpcoming = s.extractTopUpcoming(doc)

	// Extract top 10
	response.Top10 = s.extractTop10(doc)

	// Extract genres
	response.Genres = s.extractGenres(doc)

	return response, nil
}

// extractSpotlight extracts spotlight anime from the homepage
func (s *Scraper) extractSpotlight(doc *goquery.Document) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find(".deslide-wrap .swiper-wrapper .swiper-slide").Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}
		item.Rank = i + 1

		// Extract ID from href
		href, exists := sel.Find(".desi-buttons a").First().Attr("href")
		if exists {
			parts := strings.Split(href, "/")
			if len(parts) > 0 {
				item.ID = parts[len(parts)-1]
			}
		}

		// Extract poster
		item.Poster, _ = sel.Find(".deslide-cover .film-poster-img").Attr("data-src")

		// Extract title and jname
		item.Title = strings.TrimSpace(sel.Find(".desi-head-title").Text())
		item.JName, _ = sel.Find(".desi-head-title").Attr("data-jname")

		// Extract description
		item.Description = strings.TrimSpace(sel.Find(".desi-description").Text())

		// Extract details
		details := sel.Find(".sc-detail")
		item.Type = strings.TrimSpace(details.Find(".scd-item").Eq(0).Text())
		item.Duration = strings.TrimSpace(details.Find(".scd-item").Eq(1).Text())
		item.Aired = strings.TrimSpace(details.Find(".scd-item.m-hide").Text())
		item.Quality = strings.TrimSpace(details.Find(".scd-item .quality").Text())

		// Initialize Episodes to avoid nil pointer dereference
		item.Episodes = &models.Episodes{}

		// Extract episode information
		subText := strings.TrimSpace(details.Find(".tick-sub").Text())
		dubText := strings.TrimSpace(details.Find(".tick-dub").Text())
		epsText := strings.TrimSpace(details.Find(".tick-eps").Text())
		if epsText == "" {
			epsText = subText
		}

		item.Episodes.Sub, _ = strconv.Atoi(subText)
		item.Episodes.Dub, _ = strconv.Atoi(dubText)
		item.Episodes.Eps, _ = strconv.Atoi(epsText)

		items = append(items, item)
	})

	return items
}

// extractTrending extracts trending anime from the homepage
func (s *Scraper) extractTrending(doc *goquery.Document) []models.AnimeItem {
	var items []models.AnimeItem

	doc.Find("#trending-home .swiper-container .swiper-slide").Each(func(i int, sel *goquery.Selection) {
		item := models.AnimeItem{}
		item.Rank = i + 1

		// Extract title and jname
		titleEl := sel.Find(".item .film-title")
		item.Title = strings.TrimSpace(titleEl.Text())
		item.JName, _ = titleEl.Attr("data-jname")

		// Extract poster and ID
		imageEl := sel.Find(".film-poster")
		item.Poster, _ = imageEl.Find("img").Attr("data-src")

		href, exists := imageEl.Attr("href")
		if exists {
			parts := strings.Split(href, "/")
			if len(parts) > 0 {
				item.ID = parts[len(parts)-1]
			}
		}

		items = append(items, item)
	})

	return items
}

// extractLatestCompleted extracts latest completed anime
func (s *Scraper) extractLatestCompleted(doc *goquery.Document) []models.AnimeItem {
	return s.extractMostPopularAnimes(doc, "#anime-featured .row div:nth-of-type(4) .anif-block-ul ul li")
}

// extractTopAiring extracts top airing anime
func (s *Scraper) extractTopAiring(doc *goquery.Document) []models.AnimeItem {
	return s.extractMostPopularAnimes(doc, "#anime-featured .row div:nth-of-type(1) .anif-block-ul ul li")
}

// extractMostPopular extracts most popular anime
func (s *Scraper) extractMostPopular(doc *goquery.Document) []models.AnimeItem {
	return s.extractMostPopularAnimes(doc, "#anime-featured .row div:nth-of-type(2) .anif-block-ul ul li")
}

// extractMostFavorite extracts most favorite anime
func (s *Scraper) extractMostFavorite(doc *goquery.Document) []models.AnimeItem {
	return s.extractMostPopularAnimes(doc, "#anime-featured .row div:nth-of-type(3) .anif-block-ul ul li")
}

// extractRecentlyAdded extracts recently added anime
func (s *Scraper) extractRecentlyAdded(doc *goquery.Document) []models.AnimeItem {
	return s.extractAnimes(doc, "#main-content .block_area_home:contains('Recently Added') .film_list .film_list-wrap .flw-item")
}

// extractLatestUpdated extracts latest updated anime
func (s *Scraper) extractLatestUpdated(doc *goquery.Document) []models.AnimeItem {
	return s.extractAnimes(doc, "#main-content .block_area_home:nth-of-type(1) .tab-content .film_list-wrap .flw-item")
}

// extractTopUpcoming extracts top upcoming anime
func (s *Scraper) extractTopUpcoming(doc *goquery.Document) []models.AnimeItem {
	return s.extractAnimes(doc, "#main-content .block_area_home:nth-of-type(3) .tab-content .film_list-wrap .flw-item")
}

// extractTop10 extracts top 10 rankings
func (s *Scraper) extractTop10(doc *goquery.Document) models.Top10 {
	top10 := models.Top10{}

	// Extract Today's top 10 (day period)
	top10.Today = s.extractTop10Animes(doc, "day")

	// Extract Week's top 10
	top10.Week = s.extractTop10Animes(doc, "week")

	// Extract Month's top 10
	top10.Month = s.extractTop10Animes(doc, "month")

	return top10
}

// extractGenres extracts available genres
func (s *Scraper) extractGenres(doc *goquery.Document) []string {
	var genres []string

	doc.Find(".genre-list a, .footer_menu a[href*='genre']").Each(func(i int, sel *goquery.Selection) {
		genre := strings.TrimSpace(sel.Text())
		if genre != "" && !contains(genres, genre) {
			genres = append(genres, genre)
		}
	})

	return genres
}
