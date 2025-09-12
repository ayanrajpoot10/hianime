package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// GetAZList scrapes anime list organized alphabetically by sort option
func (s *Scraper) GetAZList(sortOption string, page int) (*models.AZListResponse, error) {
	if s.config.Verbose {
		fmt.Printf("Fetching A-Z list for sort option: %s (page %d)\n", sortOption, page)
	}

	// Initialize response with defaults
	response := &models.AZListResponse{
		SortOption:  strings.TrimSpace(sortOption),
		Animes:      []models.AnimeItem{},
		TotalPages:  0,
		HasNextPage: false,
		CurrentPage: page,
	}

	// Normalize current page
	if page < 1 {
		response.CurrentPage = 1
	}
	page = response.CurrentPage

	// Validate sort option
	sortOption = response.SortOption
	if sortOption == "" || !models.ValidAZListSortOptions[sortOption] {
		return nil, fmt.Errorf("invalid az-list sort option: %s", sortOption)
	}

	// Transform sort option for URL
	urlSortOption := sortOption
	switch sortOption {
	case "all":
		urlSortOption = ""
	case "other":
		urlSortOption = "other"
	default:
		// Convert to uppercase for letters A-Z
		urlSortOption = strings.ToUpper(sortOption)
	}

	// Construct the A-Z list URL
	var azURL string
	if urlSortOption == "" {
		azURL = fmt.Sprintf("/az-list?page=%d", page)
	} else {
		azURL = fmt.Sprintf("/az-list/%s?page=%d", urlSortOption, page)
	}
	url := s.config.BaseURL + azURL

	if s.config.Verbose {
		fmt.Printf("Making request to: %s\n", url)
	}

	// Make the HTTP request
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch A-Z list: %w", err)
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract animes using the main content selector
	selector := "#main-wrapper .tab-content .film_list-wrap .flw-item"
	response.Animes = s.extractAnimes(doc, selector)

	// Extract pagination information
	response.HasNextPage = s.extractHasNextPage(doc)
	response.TotalPages = s.extractTotalPages(doc)

	// If no animes found and no next page, set total pages to 0
	if len(response.Animes) == 0 && !response.HasNextPage {
		response.TotalPages = 0
	}

	return response, nil
}

// extractHasNextPage determines if there's a next page based on pagination
func (s *Scraper) extractHasNextPage(doc *goquery.Document) bool {
	paginationItems := doc.Find(".pagination > li")

	// If no pagination items, no next page
	if paginationItems.Length() == 0 {
		return false
	}

	activeItem := doc.Find(".pagination li.active")

	// If no active item, no next page
	if activeItem.Length() == 0 {
		return false
	}

	// Check if the last pagination item is active (if so, no next page)
	lastItem := paginationItems.Last()
	return !lastItem.HasClass("active")
}

// extractTotalPages extracts the total number of pages from pagination
func (s *Scraper) extractTotalPages(doc *goquery.Document) int {
	// Try to get total pages from "Last" link
	lastPageLink := doc.Find(`.pagination > .page-item a[title="Last"]`)
	if lastPageLink.Length() > 0 {
		if href, exists := lastPageLink.Attr("href"); exists {
			parts := strings.Split(href, "=")
			if len(parts) > 1 {
				if pages, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					return pages
				}
			}
		}
	}

	// Try to get total pages from "Next" link
	nextPageLink := doc.Find(`.pagination > .page-item a[title="Next"]`)
	if nextPageLink.Length() > 0 {
		if href, exists := nextPageLink.Attr("href"); exists {
			parts := strings.Split(href, "=")
			if len(parts) > 1 {
				if pages, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					return pages
				}
			}
		}
	}

	// Try to get current page from active pagination item
	activePageLink := doc.Find(".pagination > .page-item.active a")
	if activePageLink.Length() > 0 {
		if pageText := strings.TrimSpace(activePageLink.Text()); pageText != "" {
			if pages, err := strconv.Atoi(pageText); err == nil {
				return pages
			}
		}
	}

	// Default to 1 if we can't determine total pages
	return 1
}
