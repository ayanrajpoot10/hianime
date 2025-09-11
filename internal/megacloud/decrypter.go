package megacloud

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"hianime/pkg/models"

	"github.com/PuerkitoBio/goquery"
)

// HTTPClient defines the interface for making HTTP requests
type HTTPClient interface {
	Get(url string) (*http.Response, error)
	GetWithHeaders(url string, headers map[string]string) (*http.Response, error)
}

// Decrypter handles megacloud decryption operations
type Decrypter struct {
	client  HTTPClient
	baseURL string
}

// Config holds configuration for the megacloud decrypter
type Config struct {
	BaseURL string
}

// VariablePair represents extracted variables from script
type VariablePair struct {
	Key   int
	Value int
}

// ClientKeyResponse represents the response structure for client key extraction
type ClientKeyResponse struct {
	Key string `json:"key"`
}

// MegacloudEndpoints contains various megacloud endpoints
type MegacloudEndpoints struct {
	Script   string
	Sources  string
	BlogV2   string
	BlogV3   string
	MegaPlay string
}

// New creates a new megacloud decrypter instance
func New(client HTTPClient, config Config) *Decrypter {
	return &Decrypter{
		client:  client,
		baseURL: config.BaseURL,
	}
}

// GetEndpoints returns the megacloud endpoints configuration
func (d *Decrypter) GetEndpoints() MegacloudEndpoints {
	return MegacloudEndpoints{
		Script:   "https://megacloud.tv/js/player/a/prod/e1-player.min.js?v=",
		Sources:  "https://megacloud.tv/embed-2/ajax/e-1/getSources?id=",
		BlogV2:   "https://megacloud.blog/embed-2/v2/e-1/getSources?id=",
		BlogV3:   "https://megacloud.blog/embed-2/v3/e-1/getSources?id=",
		MegaPlay: "https://megaplay.buzz/stream/getSources?id=",
	}
}

// Extract3 implements extraction method 3 using megacloud.blog v2 endpoint with external keys
func (d *Decrypter) Extract3(embedURL string) (*models.StreamResponse, error) {
	// Extract source ID from embed URL
	re := regexp.MustCompile(`/([^/?]+)\?`)
	matches := re.FindStringSubmatch(embedURL)
	if len(matches) < 2 {
		return nil, fmt.Errorf("unable to extract source ID from embed URL")
	}
	sourceID := matches[1]

	// Fetch decryption key from external source
	keyURL := "https://raw.githubusercontent.com/itzzzme/megacloud-keys/refs/heads/main/key.txt"
	resp, err := d.client.Get(keyURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decryption key: %w", err)
	}
	defer resp.Body.Close()

	keyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read decryption key: %w", err)
	}
	key := strings.TrimSpace(string(keyBytes))

	// Get sources from megacloud.blog v2 endpoint
	endpoints := d.GetEndpoints()
	sourcesURL := endpoints.BlogV2 + sourceID

	headers := map[string]string{
		"Accept":           "*/*",
		"X-Requested-With": "XMLHttpRequest",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Referer":          embedURL,
	}

	srcResp, err := d.client.GetWithHeaders(sourcesURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	defer srcResp.Body.Close()

	var rawSourceData struct {
		Sources any               `json:"sources"`
		Tracks  []models.Track    `json:"tracks"`
		Intro   *models.TimeRange `json:"intro"`
		Outro   *models.TimeRange `json:"outro"`
	}

	if err := json.NewDecoder(srcResp.Body).Decode(&rawSourceData); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	// Decrypt sources if encrypted
	var decryptedSources []struct {
		File string `json:"file"`
		Type string `json:"type"`
	}

	if sourcesStr, ok := rawSourceData.Sources.(string); ok {
		// Sources are encrypted, decrypt them
		decryptedJSON, err := d.decryptWithCryptoJS(sourcesStr, key)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt sources: %w", err)
		}

		if err := json.Unmarshal([]byte(decryptedJSON), &decryptedSources); err != nil {
			return nil, fmt.Errorf("failed to parse decrypted sources: %w", err)
		}
	} else if sources, ok := rawSourceData.Sources.([]any); ok {
		// Sources are already decrypted
		for _, src := range sources {
			if srcMap, ok := src.(map[string]any); ok {
				file, _ := srcMap["file"].(string)
				srcType, _ := srcMap["type"].(string)
				decryptedSources = append(decryptedSources, struct {
					File string `json:"file"`
					Type string `json:"type"`
				}{File: file, Type: srcType})
			}
		}
	}

	if len(decryptedSources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	// Build response
	response := &models.StreamResponse{
		ID:   sourceID,
		Type: "hls",
		Link: models.StreamLink{
			File: decryptedSources[0].File,
			Type: decryptedSources[0].Type,
		},
		Tracks: rawSourceData.Tracks,
		Intro:  rawSourceData.Intro,
		Outro:  rawSourceData.Outro,
		Server: "megacloud",
		Iframe: embedURL,
	}

	return response, nil
}

// Extract4 implements extraction method 4 using megaplay.buzz endpoint
func (d *Decrypter) Extract4(embedURL, category string) (*models.StreamResponse, error) {
	// Extract episode ID from embed URL
	parts := strings.Split(embedURL, "?ep=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid embed URL format")
	}
	epID := parts[1]

	// Get iframe content
	iframeURL := fmt.Sprintf("https://megaplay.buzz/stream/s-2/%s/%s", epID, category)
	headers := map[string]string{
		"Host":                      "megaplay.buzz",
		"User-Agent":                "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0",
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"DNT":                       "1",
		"Sec-GPC":                   "1",
		"Connection":                "keep-alive",
		"Referer":                   "https://megaplay.buzz/api",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "iframe",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "same-origin",
		"Sec-Fetch-User":            "?1",
		"Priority":                  "u=4",
		"TE":                        "trailers",
	}

	iframeResp, err := d.client.GetWithHeaders(iframeURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get iframe: %w", err)
	}
	defer iframeResp.Body.Close()

	if iframeResp.StatusCode != 200 {
		return nil, fmt.Errorf("episode is not available")
	}

	doc, err := goquery.NewDocumentFromReader(iframeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iframe HTML: %w", err)
	}

	// Extract data-id from megaplay-player element
	dataID, exists := doc.Find("#megaplay-player").Attr("data-id")
	if !exists || dataID == "" {
		return nil, fmt.Errorf("could not find data-id in iframe")
	}

	// Get sources
	sourcesURL := fmt.Sprintf("https://megaplay.buzz/stream/getSources?id=%s&id=%s", dataID, dataID)
	sourcesHeaders := map[string]string{
		"Host":             "megaplay.buzz",
		"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0",
		"Accept":           "application/json, text/javascript, */*; q=0.01",
		"Accept-Language":  "en-US,en;q=0.5",
		"Accept-Encoding":  "gzip, deflate, br, zstd",
		"X-Requested-With": "XMLHttpRequest",
		"DNT":              "1",
		"Sec-GPC":          "1",
		"Connection":       "keep-alive",
		"Referer":          fmt.Sprintf("https://megaplay.buzz/stream/s-2/%s/%s", epID, category),
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-origin",
		"TE":               "trailers",
	}

	sourcesResp, err := d.client.GetWithHeaders(sourcesURL, sourcesHeaders)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	defer sourcesResp.Body.Close()

	var sourcesData struct {
		Sources struct {
			File string `json:"file"`
		} `json:"sources"`
		Tracks []models.Track    `json:"tracks"`
		Intro  *models.TimeRange `json:"intro"`
		Outro  *models.TimeRange `json:"outro"`
	}

	if err := json.NewDecoder(sourcesResp.Body).Decode(&sourcesData); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	// Build response
	response := &models.StreamResponse{
		ID:   epID,
		Type: "hls",
		Link: models.StreamLink{
			File: sourcesData.Sources.File,
			Type: "hls",
		},
		Tracks: sourcesData.Tracks,
		Intro:  sourcesData.Intro,
		Outro:  sourcesData.Outro,
		Server: "megaplay",
		Iframe: embedURL,
	}

	return response, nil
}

// Extract5 implements extraction method 5 using megacloud.blog v3 endpoint with client key
func (d *Decrypter) Extract5(embedURL string) (*models.StreamResponse, error) {
	// Extract source ID from embed URL
	re := regexp.MustCompile(`/([^/?]+)\?`)
	matches := re.FindStringSubmatch(embedURL)
	if len(matches) < 2 {
		return nil, fmt.Errorf("unable to extract source ID from embed URL")
	}
	sourceID := matches[1]

	// Fetch decryption keys from external source
	keysURL := "https://raw.githubusercontent.com/yogesh-hacker/MegacloudKeys/refs/heads/main/keys.json"
	resp, err := d.client.Get(keysURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decryption keys: %w", err)
	}
	defer resp.Body.Close()

	var keysData map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&keysData); err != nil {
		return nil, fmt.Errorf("failed to decode keys response: %w", err)
	}

	megacloudKey, exists := keysData["mega"]
	if !exists {
		return nil, fmt.Errorf("megacloud key not found in keys data")
	}

	// Extract client key from iframe
	clientKey, err := d.extractClientKey(embedURL, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract client key: %w", err)
	}

	// Get sources from megacloud.blog v3 endpoint
	endpoints := d.GetEndpoints()
	sourcesURL := fmt.Sprintf("%s%s&_k=%s", endpoints.BlogV3, sourceID, clientKey)

	headers := map[string]string{
		"Accept":           "*/*",
		"X-Requested-With": "XMLHttpRequest",
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Referer":          embedURL,
	}

	srcResp, err := d.client.GetWithHeaders(sourcesURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	defer srcResp.Body.Close()

	var rawSourceData struct {
		Sources   any               `json:"sources"`
		Tracks    []models.Track    `json:"tracks"`
		Intro     *models.TimeRange `json:"intro"`
		Outro     *models.TimeRange `json:"outro"`
		Encrypted bool              `json:"encrypted"`
	}

	if err := json.NewDecoder(srcResp.Body).Decode(&rawSourceData); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	// Decrypt sources if encrypted
	var decryptedSources []struct {
		File string `json:"file"`
		Type string `json:"type"`
	}

	if !rawSourceData.Encrypted {
		// Sources are already decrypted
		if sources, ok := rawSourceData.Sources.([]any); ok {
			for _, src := range sources {
				if srcMap, ok := src.(map[string]any); ok {
					file, _ := srcMap["file"].(string)
					srcType, _ := srcMap["type"].(string)
					decryptedSources = append(decryptedSources, struct {
						File string `json:"file"`
						Type string `json:"type"`
					}{File: file, Type: srcType})
				}
			}
		}
	} else {
		// Sources are encrypted, decrypt them
		if sourcesStr, ok := rawSourceData.Sources.(string); ok {
			decryptedJSON, err := d.decryptSrc2(sourcesStr, clientKey, megacloudKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt sources: %w", err)
			}

			if err := json.Unmarshal([]byte(decryptedJSON), &decryptedSources); err != nil {
				return nil, fmt.Errorf("failed to parse decrypted sources: %w", err)
			}
		}
	}

	if len(decryptedSources) == 0 {
		return nil, fmt.Errorf("no sources found")
	}

	// Build response
	response := &models.StreamResponse{
		ID:   sourceID,
		Type: decryptedSources[0].Type,
		Link: models.StreamLink{
			File: decryptedSources[0].File,
			Type: decryptedSources[0].Type,
		},
		Tracks: rawSourceData.Tracks,
		Intro:  rawSourceData.Intro,
		Outro:  rawSourceData.Outro,
		Server: "megacloud",
		Iframe: embedURL,
	}

	return response, nil
}

// extractClientKey extracts client key from iframe for newer endpoints
func (d *Decrypter) extractClientKey(embedURL, sourceID string) (string, error) {
	// This is a simplified implementation
	// In a full implementation, you would parse the iframe HTML and extract the client key
	// For now, we'll try to extract it from meta tags or script variables

	resp, err := d.client.Get(embedURL)
	if err != nil {
		return "", fmt.Errorf("failed to get iframe: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse iframe HTML: %w", err)
	}

	// Try to extract from meta tags
	if content, exists := doc.Find("meta[name='_gg_fb']").Attr("content"); exists && content != "" {
		return content, nil
	}

	// Try to extract from data attributes
	if dpi, exists := doc.Find("[data-dpi]").Attr("data-dpi"); exists && dpi != "" {
		return dpi, nil
	}

	// Try to extract from script variables
	html, _ := doc.Html()
	windowRegex := regexp.MustCompile(`window\.(\w+)\s*=\s*["']([\w-]+)["']`)
	matches := windowRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 2 && len(match[2]) >= 20 {
			return match[2], nil
		}
	}

	// Default fallback (this should be improved)
	return "default_client_key", nil
}

// decryptWithCryptoJS decrypts using CryptoJS AES compatible method
func (d *Decrypter) decryptWithCryptoJS(encryptedData, key string) (string, error) {
	// Decode base64 encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// For CryptoJS compatibility, we need to handle the "Salted__" prefix
	if len(ciphertext) < 16 {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Check for "Salted__" prefix (CryptoJS format)
	if string(ciphertext[:8]) == "Salted__" {
		salt := ciphertext[8:16]
		encryptedData := ciphertext[16:]

		// Derive key and IV using MD5 (CryptoJS method)
		keyIV := d.deriveKeyIV([]byte(key), salt, 32, 16)
		derivedKey := keyIV[:32]
		iv := keyIV[32:48]

		// Decrypt using AES-256-CBC
		return d.decryptAESCBC(encryptedData, derivedKey, iv)
	}

	// Fallback to standard AES decryption
	return d.decryptAES(encryptedData, key)
}

// decryptSrc2 implements the src2 decryption method
func (d *Decrypter) decryptSrc2(encryptedData, clientKey, megacloudKey string) (string, error) {
	// This is a simplified implementation
	// The actual implementation would need to match the exact algorithm used in the reference

	// Try CryptoJS decryption first
	result, err := d.decryptWithCryptoJS(encryptedData, clientKey)
	if err == nil {
		return result, nil
	}

	// Fallback to megacloud key
	result, err = d.decryptWithCryptoJS(encryptedData, megacloudKey)
	if err == nil {
		return result, nil
	}

	// Final fallback to standard AES
	return d.decryptAES(encryptedData, clientKey)
}

// deriveKeyIV derives key and IV using MD5 (CryptoJS method)
func (d *Decrypter) deriveKeyIV(password, salt []byte, keyLen, ivLen int) []byte {
	var result []byte
	var data []byte
	data = append(data, password...)
	data = append(data, salt...)

	for len(result) < keyLen+ivLen {
		hash := md5.Sum(data)
		result = append(result, hash[:]...)
		data = append(hash[:], password...)
		data = append(data, salt...)
	}

	return result[:keyLen+ivLen]
}

// decryptAESCBC decrypts data using AES-256-CBC
func (d *Decrypter) decryptAESCBC(ciphertext, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS7 padding
	if len(ciphertext) == 0 {
		return "", fmt.Errorf("decrypted data is empty")
	}

	padding := int(ciphertext[len(ciphertext)-1])
	if padding > aes.BlockSize || padding > len(ciphertext) {
		return "", fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := len(ciphertext) - padding; i < len(ciphertext); i++ {
		if ciphertext[i] != byte(padding) {
			return "", fmt.Errorf("invalid padding")
		}
	}

	return string(ciphertext[:len(ciphertext)-padding]), nil
}

// extractVariablesAdvanced extracts variables using improved regex patterns
func (d *Decrypter) extractVariablesAdvanced(text string) ([]VariablePair, error) {
	// Simplified regex pattern that works with Go's RE2 engine
	regex := regexp.MustCompile(`case\s*0x[0-9a-f]+:\s*\w+\s*=\s*(\w+)\s*,\s*\w+\s*=\s*(\w+);`)
	matches := regex.FindAllStringSubmatch(text, -1)

	var vars []VariablePair
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		key1, err := d.matchingKey(match[1], text)
		if err != nil {
			continue
		}

		key2, err := d.matchingKey(match[2], text)
		if err != nil {
			continue
		}

		intKey1, err1 := strconv.ParseInt(key1, 16, 64)
		intKey2, err2 := strconv.ParseInt(key2, 16, 64)

		if err1 == nil && err2 == nil {
			vars = append(vars, VariablePair{
				Key:   int(intKey1),
				Value: int(intKey2),
			})
		}
	}

	if len(vars) == 0 {
		return nil, fmt.Errorf("no variables found")
	}

	return vars, nil
}

// matchingKey finds the matching key value in the script
func (d *Decrypter) matchingKey(value, script string) (string, error) {
	pattern := fmt.Sprintf(`,%s=((?:0x)?([0-9a-fA-F]+))`, regexp.QuoteMeta(value))
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(script)

	if len(match) > 1 {
		return strings.TrimPrefix(match[1], "0x"), nil
	}

	return "", fmt.Errorf("failed to match key for value: %s", value)
}

// getSecretAdvanced extracts secret using improved algorithm
func (d *Decrypter) getSecretAdvanced(encryptedString string, variables []VariablePair) (string, string, error) {
	var secret strings.Builder
	encryptedSource := make([]rune, len(encryptedString))
	copy(encryptedSource, []rune(encryptedString))

	currentIndex := 0

	for _, variable := range variables {
		start := variable.Key + currentIndex
		end := start + variable.Value

		if start >= 0 && end <= len(encryptedString) {
			for i := start; i < end; i++ {
				if i < len(encryptedString) {
					secret.WriteRune(rune(encryptedString[i]))
					if i < len(encryptedSource) {
						encryptedSource[i] = 0 // Mark as removed
					}
				}
			}
			currentIndex += variable.Value
		}
	}

	// Rebuild encrypted source without removed characters
	var finalEncryptedSource strings.Builder
	for _, r := range encryptedSource {
		if r != 0 {
			finalEncryptedSource.WriteRune(r)
		}
	}

	return secret.String(), finalEncryptedSource.String(), nil
}

func (d *Decrypter) Decrypt(selectedServer *models.Server, episodeID string) (*models.StreamResponse, error) {
	// Try multiple extraction methods in order of preference
	methods := []func(*models.Server, string) (*models.StreamResponse, error){
		d.DecryptWithExtract5,
		d.DecryptWithExtract3,
		d.DecryptWithExtract4,
		d.DecryptOriginal,
	}

	var lastErr error
	for i, method := range methods {
		result, err := method(selectedServer, episodeID)
		if err == nil && result != nil {
			return result, nil
		}
		lastErr = err

		// Log the attempt for debugging
		fmt.Printf("Extraction method %d failed: %v\n", i+1, err)
	}

	return nil, fmt.Errorf("all extraction methods failed, last error: %w", lastErr)
}

// DecryptWithExtract5 uses extraction method 5
func (d *Decrypter) DecryptWithExtract5(selectedServer *models.Server, episodeID string) (*models.StreamResponse, error) {
	// Extract the embed URL from the server data
	embedURL := selectedServer.Name // Assuming this contains the embed URL
	if embedURL == "" {
		return nil, fmt.Errorf("no embed URL found in server data")
	}

	return d.Extract5(embedURL)
}

// DecryptWithExtract3 uses extraction method 3
func (d *Decrypter) DecryptWithExtract3(selectedServer *models.Server, episodeID string) (*models.StreamResponse, error) {
	// Extract the embed URL from the server data
	embedURL := selectedServer.Name // Assuming this contains the embed URL
	if embedURL == "" {
		return nil, fmt.Errorf("no embed URL found in server data")
	}

	return d.Extract3(embedURL)
}

// DecryptWithExtract4 uses extraction method 4
func (d *Decrypter) DecryptWithExtract4(selectedServer *models.Server, episodeID string) (*models.StreamResponse, error) {
	// Extract the embed URL and category from the server data
	embedURL := selectedServer.Name // Assuming this contains the embed URL
	if embedURL == "" {
		return nil, fmt.Errorf("no embed URL found in server data")
	}

	// Default to "sub" category, this could be configurable
	category := "sub"
	if strings.Contains(strings.ToLower(selectedServer.Type), "dub") {
		category = "dub"
	}

	return d.Extract4(embedURL, category)
}

// DecryptOriginal implements the original megacloud decryption logic
func (d *Decrypter) DecryptOriginal(selectedServer *models.Server, episodeID string) (*models.StreamResponse, error) {
	epID := strings.Split(episodeID, "::ep=")
	if len(epID) != 2 {
		return nil, fmt.Errorf("invalid episode ID format")
	}

	// Step 1: Get sources data
	sourcesURL := fmt.Sprintf("%s/ajax/v2/episode/sources?id=%s", d.baseURL, selectedServer.ID)

	headers := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
	}

	resp, err := d.client.GetWithHeaders(sourcesURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}
	defer resp.Body.Close()

	var sourcesData struct {
		Link string `json:"link"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&sourcesData); err != nil {
		return nil, fmt.Errorf("failed to decode sources response: %w", err)
	}

	if sourcesData.Link == "" {
		return nil, fmt.Errorf("no link found in sources data")
	}

	// Step 2: Extract source ID from link
	sourceIDRegex := regexp.MustCompile(`/([^/?]+)\?`)
	matches := sourceIDRegex.FindStringSubmatch(sourcesData.Link)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract source ID from link")
	}
	sourceID := matches[1]

	// Step 3: Extract base URL
	baseURLRegex := regexp.MustCompile(`^(https?://[^/]+(?:/[^/]+){3})`)
	baseMatches := baseURLRegex.FindStringSubmatch(sourcesData.Link)
	if len(baseMatches) < 2 {
		return nil, fmt.Errorf("could not extract base URL from link")
	}
	baseURL := baseMatches[1]

	// Step 4: Extract token
	token, err := d.extractToken(fmt.Sprintf("%s/%s?k=1&autoPlay=0&oa=0&asi=1", baseURL, sourceID))
	if err != nil {
		return nil, fmt.Errorf("failed to extract token: %w", err)
	}

	// Step 5: Get encrypted sources
	getSrcURL := fmt.Sprintf("%s/getSources?id=%s&_k=%s", baseURL, sourceID, token)

	headers2 := map[string]string{
		"Referer": fmt.Sprintf("%s/%s?k=1", baseURL, sourceID),
	}

	resp2, err := d.client.GetWithHeaders(getSrcURL, headers2)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted sources: %w", err)
	}
	defer resp2.Body.Close()

	var rawSourceData struct {
		Sources any               `json:"sources"`
		Tracks  []models.Track    `json:"tracks"`
		Intro   *models.TimeRange `json:"intro"`
		Outro   *models.TimeRange `json:"outro"`
	}

	if err := json.NewDecoder(resp2.Body).Decode(&rawSourceData); err != nil {
		return nil, fmt.Errorf("failed to decode encrypted sources: %w", err)
	}

	// Step 6: Decrypt sources
	var decryptedSources []struct {
		File string `json:"file"`
	}

	// Try to decrypt if sources is a string (encrypted)
	if sourcesStr, ok := rawSourceData.Sources.(string); ok {
		// Fetch the decryption key from GitHub
		key, err := d.fetchDecryptionKey()
		if err != nil {
			// Fallback to alternative providers if key fetch fails
			fallbackSources, fallbackErr := d.fallbackDecrypt(epID[1], selectedServer.Type)
			if fallbackErr != nil {
				return nil, fmt.Errorf("failed to decrypt sources and fallback failed: %w", fallbackErr)
			}
			// Convert fallback sources to our format
			for _, src := range fallbackSources {
				decryptedSources = append(decryptedSources, struct {
					File string `json:"file"`
				}{File: src.File})
			}
		} else {
			// Try AES decryption with the fetched key
			decryptedJSON, err := d.decryptAES(sourcesStr, key)
			if err != nil {
				// Fallback if decryption fails
				fallbackSources, fallbackErr := d.fallbackDecrypt(epID[1], selectedServer.Type)
				if fallbackErr != nil {
					return nil, fmt.Errorf("failed to decrypt sources and fallback failed: %w", fallbackErr)
				}
				// Convert fallback sources to our format
				for _, src := range fallbackSources {
					decryptedSources = append(decryptedSources, struct {
						File string `json:"file"`
					}{File: src.File})
				}
			} else {
				// Parse decrypted JSON
				if err := json.Unmarshal([]byte(decryptedJSON), &decryptedSources); err != nil {
					return nil, fmt.Errorf("failed to parse decrypted sources: %w", err)
				}
			}
		}
	} else {
		// Sources are already decrypted
		if sources, ok := rawSourceData.Sources.([]any); ok && len(sources) > 0 {
			if sourceMap, ok := sources[0].(map[string]any); ok {
				if file, ok := sourceMap["file"].(string); ok {
					decryptedSources = []struct {
						File string `json:"file"`
					}{{File: file}}
				}
			}
		}
	}

	if len(decryptedSources) == 0 {
		return nil, fmt.Errorf("no decrypted sources found")
	}

	// Step 7: Build response
	response := &models.StreamResponse{
		ID:   episodeID,
		Type: selectedServer.Type,
		Link: models.StreamLink{
			File: decryptedSources[0].File,
			Type: "hls",
		},
		Tracks: rawSourceData.Tracks,
		Intro:  rawSourceData.Intro,
		Outro:  rawSourceData.Outro,
		Server: selectedServer.Name,
		Iframe: sourcesData.Link,
	}

	return response, nil
}

// TryAdvancedExtraction attempts to use advanced extraction with variable parsing
func (d *Decrypter) TryAdvancedExtraction(encryptedString, scriptText string) ([]struct {
	File string `json:"file"`
}, error) {
	// Extract variables using improved method
	vars, err := d.extractVariablesAdvanced(scriptText)
	if err != nil {
		return nil, fmt.Errorf("failed to extract variables: %w", err)
	}

	// Get secret and encrypted source
	secret, encryptedSource, err := d.getSecretAdvanced(encryptedString, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	// Decrypt the source
	decryptedJSON, err := d.decryptAES(encryptedSource, secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt source: %w", err)
	}

	var sources []struct {
		File string `json:"file"`
	}

	if err := json.Unmarshal([]byte(decryptedJSON), &sources); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted sources: %w", err)
	}

	return sources, nil
}

// extractToken extracts token from megacloud page
func (d *Decrypter) extractToken(url string) (string, error) {
	resp, err := d.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Try to extract token from various sources

	// 1. Meta tag
	if meta, exists := doc.Find("meta[name='_gg_fb']").Attr("content"); exists && meta != "" {
		return meta, nil
	}

	// 2. Data attribute
	if dpi, exists := doc.Find("[data-dpi]").Attr("data-dpi"); exists && dpi != "" {
		return dpi, nil
	}

	// 3. Script nonce
	doc.Find("script[nonce]").Each(func(i int, sel *goquery.Selection) {
		if strings.Contains(sel.Text(), "empty nonce script") {
			if nonce, exists := sel.Attr("nonce"); exists && nonce != "" {
				return
			}
		}
	})

	// 4. Extract from JavaScript variables
	html, _ := doc.Html()

	// Look for window assignments
	windowRegex := regexp.MustCompile(`window\.(\w+)\s*=\s*["']([\w-]+)["']`)
	matches := windowRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 2 && len(match[2]) >= 20 {
			return match[2], nil
		}
	}

	// 5. HTML comments
	commentRegex := regexp.MustCompile(`<!--\s*_is_th:([\w-]+)\s*-->`)
	if commentMatch := commentRegex.FindStringSubmatch(html); len(commentMatch) > 1 {
		return commentMatch[1], nil
	}

	// Default fallback
	return "default_token", nil
}

// fallbackDecrypt implements fallback decryption using alternative providers
func (d *Decrypter) fallbackDecrypt(epID, serverType string) ([]struct{ File string }, error) {
	// Try megaplay.buzz fallback
	fallbackURL := fmt.Sprintf("https://megaplay.buzz/stream/s-2/%s/%s", epID, serverType)

	headers := map[string]string{
		"Referer": "https://megaplay.buzz/",
	}

	resp, err := d.client.GetWithHeaders(fallbackURL, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract data-id from HTML
	html, _ := doc.Html()
	dataIDRegex := regexp.MustCompile(`data-id=["'](\d+)["']`)
	matches := dataIDRegex.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract data-id from fallback")
	}

	realID := matches[1]

	// Get sources from fallback API
	fallbackSrcURL := fmt.Sprintf("https://megaplay.buzz/stream/getSources?id=%s", realID)

	headers2 := map[string]string{
		"X-Requested-With": "XMLHttpRequest",
	}

	resp2, err := d.client.GetWithHeaders(fallbackSrcURL, headers2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var fallbackData struct {
		Sources struct {
			File string `json:"file"`
		} `json:"sources"`
	}

	if err := json.NewDecoder(resp2.Body).Decode(&fallbackData); err != nil {
		return nil, err
	}

	if fallbackData.Sources.File == "" {
		return nil, fmt.Errorf("no file found in fallback data")
	}

	return []struct{ File string }{{File: fallbackData.Sources.File}}, nil
}

// fetchDecryptionKey fetches the AES decryption key from GitHub
func (d *Decrypter) fetchDecryptionKey() (string, error) {
	keyURL := "https://raw.githubusercontent.com/itzzzme/megacloud-keys/refs/heads/main/key.txt"

	resp, err := d.client.Get(keyURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	keyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(keyBytes)), nil
}

// decryptAES decrypts AES encrypted data using the provided key
func (d *Decrypter) decryptAES(encryptedData, key string) (string, error) {
	// Decode base64 encrypted data
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Trim key to 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	} else if len(keyBytes) < 32 {
		// Pad key if it's too short
		paddedKey := make([]byte, 32)
		copy(paddedKey, keyBytes)
		keyBytes = paddedKey
	}

	// Create AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt using CBC mode
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Ensure ciphertext length is multiple of block size
	if len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS7 padding
	if len(ciphertext) == 0 {
		return "", fmt.Errorf("decrypted data is empty")
	}

	padding := int(ciphertext[len(ciphertext)-1])
	if padding > aes.BlockSize || padding > len(ciphertext) {
		return "", fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := len(ciphertext) - padding; i < len(ciphertext); i++ {
		if ciphertext[i] != byte(padding) {
			return "", fmt.Errorf("invalid padding")
		}
	}

	return string(ciphertext[:len(ciphertext)-padding]), nil
}

// GetHD4FallbackURL generates a fallback URL for HD-4 server
func (d *Decrypter) GetHD4FallbackURL(episodeID, serverType string) (string, error) {
	epID := strings.Split(episodeID, "::ep=")
	if len(epID) != 2 {
		return "", fmt.Errorf("invalid episode ID format")
	}

	return fmt.Sprintf("https://megaplay.buzz/stream/s-2/%s/%s", epID[1], serverType), nil
}
