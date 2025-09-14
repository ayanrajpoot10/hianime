package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ayanrajpoot10/hianime-api/internal/decrypt"
	"github.com/ayanrajpoot10/hianime-api/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Servers scrapes available servers for a specific episode
func (s *Scraper) Servers(episodeID string) (*models.ServersResponse, error) {
	// Extract episode number from ID
	if !strings.Contains(episodeID, "::ep=") {
		return nil, fmt.Errorf("invalid episode ID format")
	}

	epParts := strings.Split(episodeID, "::ep=")
	if len(epParts) != 2 {
		return nil, fmt.Errorf("invalid episode ID format")
	}

	episodeNum := epParts[1]

	url := fmt.Sprintf("%s/ajax/v2/episode/servers?episodeId=%s", s.config.BaseURL, episodeNum)

	headers := map[string]string{
		"Referer":          fmt.Sprintf("%s/watch/%s", s.config.BaseURL, strings.ReplaceAll(episodeID, "::", "?")),
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
		Status any    `json:"status"` // Can be string or bool
		HTML   string `json:"html"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ajaxResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	// Check status - handle both string and boolean values
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

	response := &models.ServersResponse{}

	// Parse episode number
	if epNum, err := strconv.Atoi(episodeNum); err == nil {
		response.Episode = epNum
	}

	// Extract sub servers
	doc.Find(".ps_-block .ps__-list .server-item[data-type='sub']").Each(func(i int, sel *goquery.Selection) {
		server := models.Server{
			Type:  "sub",
			Index: i,
		}

		server.Name = strings.TrimSpace(sel.Text())
		server.ID, _ = sel.Attr("data-id")

		response.Sub = append(response.Sub, server)
	})

	// Extract dub servers
	doc.Find(".ps_-block .ps__-list .server-item[data-type='dub']").Each(func(i int, sel *goquery.Selection) {
		server := models.Server{
			Type:  "dub",
			Index: i,
		}

		server.Name = strings.TrimSpace(sel.Text())
		server.ID, _ = sel.Attr("data-id")

		response.Dub = append(response.Dub, server)
	})

	return response, nil
}

// StreamLinks scrapes streaming links for a specific episode and server using megacloud decryption
func (s *Scraper) StreamLinks(episodeID, serverType, serverName string) (*models.StreamResponse, error) {
	// First get the servers to find the server ID
	servers, err := s.Servers(episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get servers: %w", err)
	}

	var selectedServer *models.Server
	if strings.ToLower(serverType) == "sub" {
		for _, server := range servers.Sub {
			if strings.EqualFold(server.Name, serverName) {
				selectedServer = &server
				break
			}
		}
	} else {
		for _, server := range servers.Dub {
			if strings.EqualFold(server.Name, serverName) {
				selectedServer = &server
				break
			}
		}
	}

	if selectedServer == nil {
		return nil, fmt.Errorf("server not found: %s (%s)", serverName, serverType)
	}

	// Create megacloud decryptor and use it
	decryptor := decrypt.NewMegacloudDecryptor(s.client, s.config)
	return decryptor.Decrypt(selectedServer, episodeID)
}
