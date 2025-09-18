package decrypt

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ayanrajpoot10/hianime-api/config"
	"github.com/ayanrajpoot10/hianime-api/pkg/httpclient"
	"github.com/ayanrajpoot10/hianime-api/pkg/models"
)

// MegacloudDecryptor handles megacloud streaming decryption
type MegacloudDecryptor struct {
	client         *httpclient.Client
	config         *config.Config
	tokenExtractor *TokenExtractor
}

// NewMegacloudDecryptor creates a new megacloud decryptor
func NewMegacloudDecryptor(client *httpclient.Client, config *config.Config) *MegacloudDecryptor {
	return &MegacloudDecryptor{
		client:         client,
		config:         config,
		tokenExtractor: NewTokenExtractor(client, config),
	}
}

// Decrypt implements the megacloud streaming logic with decryption and fallback
func (md *MegacloudDecryptor) Decrypt(selectedServer *models.Server, id string) (*models.StreamResponse, error) {
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
		data map[string]any
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
		url := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%s", md.config.BaseURL, selectedServer.ID)
		resp, err := md.client.Get(url)
		if err != nil {
			sourcesChan <- sourcesResult{err: err}
			return
		}
		defer resp.Body.Close()

		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			sourcesChan <- sourcesResult{err: err}
			return
		}
		sourcesChan <- sourcesResult{data: result}
	}()

	// Get decryption key
	go func() {
		resp, err := md.client.Get("https://raw.githubusercontent.com/itzzzme/megacloud-keys/refs/heads/main/key.txt")
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

	var decryptedSources []map[string]any
	var rawSourceData map[string]any

	// Try main decryption method
	tokenURL := fmt.Sprintf("%s/%s?k=1&autoPlay=0&oa=0&asi=1", baseURL, sourceID)
	token, tokenErr := md.tokenExtractor.ExtractToken(tokenURL)

	if tokenErr == nil {
		// Get sources with token
		sourcesURL := fmt.Sprintf("%s/getSources?id=%s&_k=%s", baseURL, sourceID, token)
		resp, err := md.client.Get(sourcesURL)
		if err == nil {
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&rawSourceData); err == nil {
				if encrypted, ok := rawSourceData["sources"].(string); ok && encrypted != "" {
					if decrypted, err := SimpleAESDecrypt(encrypted, key); err == nil {
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

		resp, err := md.client.GetWithHeaders(fallbackURL, headers)
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

		resp2, err := md.client.GetWithHeaders(fallbackSourcesURL, headers)
		if err != nil {
			return nil, fmt.Errorf("fallback sources request failed: %w", err)
		}
		defer resp2.Body.Close()

		var fallbackData map[string]any
		if err := json.NewDecoder(resp2.Body).Decode(&fallbackData); err != nil {
			return nil, fmt.Errorf("failed to decode fallback data: %w", err)
		}

		// Extract file URL from fallback
		if sources, ok := fallbackData["sources"].(map[string]any); ok {
			if file, ok := sources["file"].(string); ok {
				decryptedSources = []map[string]any{
					{"file": file},
				}
			}
		}

		// Use fallback data for tracks, intro, outro if main data is empty
		if rawSourceData == nil {
			rawSourceData = make(map[string]any)
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
	if tracks, ok := rawSourceData["tracks"].([]any); ok {
		for _, track := range tracks {
			if trackMap, ok := track.(map[string]any); ok {
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
	if intro, ok := rawSourceData["intro"].(map[string]any); ok {
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
	if outro, ok := rawSourceData["outro"].(map[string]any); ok {
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
