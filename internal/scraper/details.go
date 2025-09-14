package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ayanrajpoot10/hianime-api/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// AnimeDetails scrapes detailed information about a specific anime
func (s *Scraper) AnimeDetails(animeID string) (*models.AnimeDetailResponse, error) {
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

	getText := func(selector string) string {
		return strings.TrimSpace(doc.Find(selector).Text())
	}

	extractList := func(sel *goquery.Selection) []string {
		list := []string{}
		sel.Find("a").Each(func(i int, a *goquery.Selection) {
			if text := strings.TrimSpace(a.Text()); text != "" {
				list = append(list, text)
			}
		})
		return list
	}

	detail := &models.AnimeDetailResponse{}

	detail.ID = animeID
	detail.Title = strings.TrimSpace(doc.Find(".anisc-detail h2.film-name.dynamic-name").Text())
	detail.JName = doc.Find(".anisc-info .film-name").AttrOr("data-jname", "")
	detail.Poster = doc.Find(".anisc-poster .film-poster-img").AttrOr("src", "")
	detail.Description = strings.TrimSpace(doc.Find(".film-description.m-hide .text").Text())
	detail.Episodes = &models.Episodes{}
	detail.RelatedAnimes = s.extractAnimes(doc, ".block_area .tab-content .flw-item")
	detail.RecommendedAnimes = s.extractAnimes(doc, ".block_area:contains('You might also like') .flw-item")

	// Extract Japanese title and synonyms
	doc.Find(".anisc-info .item").Each(func(i int, sel *goquery.Selection) {
		label := strings.ToLower(strings.TrimSpace(sel.Find(".item-head").Text()))
		value := strings.TrimSpace(sel.Find(".name").Text())

		switch label {
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
			detail.Studios = extractList(sel)
		case "producers:":
			detail.Producers = extractList(sel)
		case "genres:":
			detail.Genres = extractList(sel)
		}
	})

	// Extract episode counts safely
	if subCount, err := strconv.Atoi(getText(".anisc-info .tick-sub")); err == nil {
		detail.Episodes.Sub = subCount
	}
	if dubCount, err := strconv.Atoi(getText(".anisc-info .tick-dub")); err == nil {
		detail.Episodes.Dub = dubCount
	}
	if epsCount, err := strconv.Atoi(getText(".anisc-info .tick-eps")); err == nil {
		detail.Episodes.Eps = epsCount
	}

	// Extract other seasons
	detail.OtherSeasons = []models.Season{}
	doc.Find(".block_area-seasons .os-list .os-item").Each(func(i int, sel *goquery.Selection) {
		href := sel.AttrOr("href", "")
		id := strings.TrimPrefix(href, "/")
		season := models.Season{
			ID:     id,
			Title:  sel.Find(".title").Text(),
			URL:    fmt.Sprintf("%s%s", s.config.BaseURL, href),
			Poster: strings.TrimPrefix(sel.Find(".season-poster").AttrOr("style", ""), "background-image: url("),
		}
		season.Poster = strings.TrimRight(season.Poster, ");")
		detail.OtherSeasons = append(detail.OtherSeasons, season)
	})

	return detail, nil
}
