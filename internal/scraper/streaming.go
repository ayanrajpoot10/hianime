package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// Servers scrapes available servers for a specific episode
func (s *Scraper) Servers(episodeID string) (*models.ServersResponse, error) {
	// Rate limiting
	time.Sleep(s.config.RateLimit)

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
	// Rate limiting
	time.Sleep(s.config.RateLimit)

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

	return s.megacloud(selectedServer, episodeID)
}

// megacloud implements the megacloud streaming logic with decryption and fallback
func (s *Scraper) megacloud(selectedServer *models.Server, id string) (*models.StreamResponse, error) {
	// Extract episode number from ID
	epParts := strings.Split(id, "ep=")
	if len(epParts) != 2 {
		return nil, fmt.Errorf("invalid episode ID format")
	}
	epID := epParts[1]

	fallback1 := "megaplay.buzz"
	fallback2 := "vidwish.live"

	// Fetch sources data and decryption key concurrently
	type sourcesResult struct {
		data map[string]interface{}
		err  error
	}

	type keyResult struct {
		key string
		err error
	}

	sourcesChan := make(chan sourcesResult, 1)
	keyChan := make(chan keyResult, 1)

	// Get sources data
	go func() {
		url := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%s", s.config.BaseURL, selectedServer.ID)
		resp, err := s.client.Get(url)
		if err != nil {
			sourcesChan <- sourcesResult{err: err}
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			sourcesChan <- sourcesResult{err: err}
			return
		}
		sourcesChan <- sourcesResult{data: result}
	}()

	// Get decryption key
	go func() {
		resp, err := s.client.Get("https://raw.githubusercontent.com/itzzzme/megacloud-keys/refs/heads/main/key.txt")
		if err != nil {
			keyChan <- keyResult{err: err}
			return
		}
		defer resp.Body.Close()

		keyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			keyChan <- keyResult{err: err}
			return
		}
		keyChan <- keyResult{key: strings.TrimSpace(string(keyBytes))}
	}()

	// Wait for both results
	sourcesRes := <-sourcesChan
	keyRes := <-keyChan

	if sourcesRes.err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", sourcesRes.err)
	}

	if keyRes.err != nil {
		return nil, fmt.Errorf("failed to get decryption key: %w", keyRes.err)
	}

	sourcesData := sourcesRes.data
	key := keyRes.key

	// Extract ajax link
	ajaxLink, ok := sourcesData["link"].(string)
	if !ok || ajaxLink == "" {
		return nil, fmt.Errorf("missing link in sourcesData")
	}

	// Extract source ID from link
	sourceIDRegex := regexp.MustCompile(`/([^/?]+)\?`)
	sourceIDMatch := sourceIDRegex.FindStringSubmatch(ajaxLink)
	if len(sourceIDMatch) < 2 {
		return nil, fmt.Errorf("unable to extract sourceId from link")
	}
	sourceID := sourceIDMatch[1]

	// Extract base URL
	baseURLRegex := regexp.MustCompile(`^(https?://[^/]+(?:/[^/]+){3})`)
	baseURLMatch := baseURLRegex.FindStringSubmatch(ajaxLink)
	if len(baseURLMatch) < 2 {
		return nil, fmt.Errorf("could not extract base URL from ajaxLink")
	}
	baseURL := baseURLMatch[1]

	var decryptedSources []map[string]interface{}
	var rawSourceData map[string]interface{}

	// Try main decryption method
	tokenURL := fmt.Sprintf("%s/%s?k=1&autoPlay=0&oa=0&asi=1", baseURL, sourceID)
	token, tokenErr := s.extractToken(tokenURL)

	if tokenErr == nil {
		// Get sources with token
		sourcesURL := fmt.Sprintf("%s/getSources?id=%s&_k=%s", baseURL, sourceID, token)
		resp, err := s.client.Get(sourcesURL)
		if err == nil {
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&rawSourceData); err == nil {
				if encrypted, ok := rawSourceData["sources"].(string); ok && encrypted != "" {
					if decrypted, err := simpleAESDecrypt(encrypted, key); err == nil {
						if err := json.Unmarshal([]byte(decrypted), &decryptedSources); err == nil {
							// Success with main method
						}
					}
				}
			}
		}
	}

	// If main method failed, try fallback
	if len(decryptedSources) == 0 {
		fallback := fallback1
		if strings.EqualFold(selectedServer.Name, "hd-1") {
			fallback = fallback1
		} else {
			fallback = fallback2
		}

		fallbackURL := fmt.Sprintf("https://%s/stream/s-2/%s/%s", fallback, epID, selectedServer.Type)
		headers := map[string]string{
			"Referer": fmt.Sprintf("https://%s/", fallback1),
		}

		resp, err := s.client.GetWithHeaders(fallbackURL, headers)
		if err != nil {
			return nil, fmt.Errorf("fallback failed: %w", err)
		}
		defer resp.Body.Close()

		// Read HTML response
		htmlBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read fallback response: %w", err)
		}
		html := string(htmlBytes)

		// Extract data-id
		dataIDRegex := regexp.MustCompile(`data-id=["'](\d+)["']`)
		dataIDMatch := dataIDRegex.FindStringSubmatch(html)
		if len(dataIDMatch) < 2 {
			return nil, fmt.Errorf("could not extract data-id for fallback")
		}
		realID := dataIDMatch[1]

		// Get fallback sources
		fallbackSourcesURL := fmt.Sprintf("https://%s/stream/getSources?id=%s", fallback, realID)
		headers = map[string]string{
			"X-Requested-With": "XMLHttpRequest",
		}

		resp2, err := s.client.GetWithHeaders(fallbackSourcesURL, headers)
		if err != nil {
			return nil, fmt.Errorf("fallback sources request failed: %w", err)
		}
		defer resp2.Body.Close()

		var fallbackData map[string]interface{}
		if err := json.NewDecoder(resp2.Body).Decode(&fallbackData); err != nil {
			return nil, fmt.Errorf("failed to decode fallback data: %w", err)
		}

		// Extract file URL from fallback
		if sources, ok := fallbackData["sources"].(map[string]interface{}); ok {
			if file, ok := sources["file"].(string); ok {
				decryptedSources = []map[string]interface{}{
					{"file": file},
				}
			}
		}

		// Use fallback data for tracks, intro, outro if main data is empty
		if rawSourceData == nil {
			rawSourceData = make(map[string]interface{})
		}
		if rawSourceData["tracks"] == nil {
			if tracks, ok := fallbackData["tracks"]; ok {
				rawSourceData["tracks"] = tracks
			}
		}
		if rawSourceData["intro"] == nil {
			if intro, ok := fallbackData["intro"]; ok {
				rawSourceData["intro"] = intro
			}
		}
		if rawSourceData["outro"] == nil {
			if outro, ok := fallbackData["outro"]; ok {
				rawSourceData["outro"] = outro
			}
		}
	}

	if len(decryptedSources) == 0 {
		return nil, fmt.Errorf("no streaming sources found")
	}

	// Build response
	response := &models.StreamResponse{
		ID:     id,
		Type:   selectedServer.Type,
		Server: selectedServer.Name,
	}

	// Set main stream link
	if file, ok := decryptedSources[0]["file"].(string); ok {
		response.Link = models.StreamLink{
			File: file,
			Type: "hls",
		}
	}

	// Set tracks
	if tracks, ok := rawSourceData["tracks"].([]interface{}); ok {
		for _, track := range tracks {
			if trackMap, ok := track.(map[string]interface{}); ok {
				t := models.Track{}
				if file, ok := trackMap["file"].(string); ok {
					t.File = file
				}
				if kind, ok := trackMap["kind"].(string); ok {
					t.Kind = kind
				}
				response.Tracks = append(response.Tracks, t)
			}
		}
	}

	// Set intro
	if intro, ok := rawSourceData["intro"].(map[string]interface{}); ok {
		tr := &models.TimeRange{}
		if start, ok := intro["start"].(float64); ok {
			tr.Start = int(start)
		}
		if end, ok := intro["end"].(float64); ok {
			tr.End = int(end)
		}
		response.Intro = tr
	}

	// Set outro
	if outro, ok := rawSourceData["outro"].(map[string]interface{}); ok {
		tr := &models.TimeRange{}
		if start, ok := outro["start"].(float64); ok {
			tr.Start = int(start)
		}
		if end, ok := outro["end"].(float64); ok {
			tr.End = int(end)
		}
		response.Outro = tr
	}

	return response, nil
}
