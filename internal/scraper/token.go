package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// extractToken extracts various tokens and parameters from HTML pages
func (s *Scraper) extractToken(url string) (string, error) {
	// Rate limiting
	time.Sleep(s.config.RateLimit)

	headers := map[string]string{
		"Referer": s.config.BaseURL + "/",
	}

	resp, err := s.client.GetWithHeaders(url, headers)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	html, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML content: %w", err)
	}

	results := make(map[string]string)

	// 1. Meta tag
	if meta, exists := doc.Find(`meta[name="_gg_fb"]`).Attr("content"); exists && meta != "" {
		results["meta"] = meta
	}

	// 2. Data attribute
	if dpi, exists := doc.Find(`[data-dpi]`).Attr("data-dpi"); exists && dpi != "" {
		results["dataDpi"] = dpi
	}

	// 3. Nonce from empty script
	doc.Find("script[nonce]").Each(func(i int, sel *goquery.Selection) {
		text := sel.Text()
		if strings.Contains(text, "empty nonce script") {
			if nonce, exists := sel.Attr("nonce"); exists && nonce != "" {
				results["nonce"] = nonce
			}
		}
	})

	// 4. JS string assignment: window.<key> = "value";
	stringAssignRegex := regexp.MustCompile(`window\.(\w+)\s*=\s*["']([\w-]+)["']`)
	stringMatches := stringAssignRegex.FindAllStringSubmatch(html, -1)
	for _, match := range stringMatches {
		if len(match) >= 3 {
			key := fmt.Sprintf("window.%s", match[1])
			value := match[2]
			results[key] = value
		}
	}

	// 5. JS object assignment: window.<key> = { ... };
	objectAssignRegex := regexp.MustCompile(`window\.(\w+)\s*=\s*(\{[\s\S]*?\});`)
	objectMatches := objectAssignRegex.FindAllStringSubmatch(html, -1)
	for _, match := range objectMatches {
		if len(match) >= 3 {
			varName := match[1]
			rawObj := match[2]

			// Try to parse as JSON
			var parsedObj map[string]interface{}
			if err := json.Unmarshal([]byte(rawObj), &parsedObj); err == nil {
				var stringValues []string
				for _, val := range parsedObj {
					if strVal, ok := val.(string); ok {
						stringValues = append(stringValues, strVal)
					}
				}
				concatenated := strings.Join(stringValues, "")
				if len(concatenated) >= 20 {
					key := fmt.Sprintf("window.%s", varName)
					results[key] = concatenated
				}
			}
		}
	}

	// 6. HTML comment: <!-- _is_th:... -->
	doc.Contents().Each(func(i int, sel *goquery.Selection) {
		for _, node := range sel.Nodes {
			if node.Type == 8 { // Comment node
				comment := strings.TrimSpace(node.Data)
				commentRegex := regexp.MustCompile(`^_is_th:([\w-]+)$`)
				if match := commentRegex.FindStringSubmatch(comment); match != nil {
					results["commentToken"] = strings.TrimSpace(match[1])
				}
			}
		}
	})

	// Return the first non-empty token found
	for _, token := range results {
		if token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no token found")
}
