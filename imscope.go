package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// Check if the asset type argument is provided
	if len(os.Args) < 2 {
		log.Fatal(`Usage: go run main.go <assetType>
Available options for <assetType>:
  - smart_contract
  - websites_and_applications
  - blockchain_dlt
  - all (fetches all asset types)`)
	}
	assetType := os.Args[1]

	if assetType != "smart_contract" && assetType != "websites_and_applications" && assetType != "blockchain_dlt" && assetType != "all" {
		log.Fatal(`Invalid assetType. Allowed values are:
  - smart_contract
  - websites_and_applications
  - blockchain_dlt
  - all`)
	}

	url := "https://immunefi.com/bug-bounty/"

	// Create a ChromeDP context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var hrefs []string
	var dynamicValue string

	// Run the scraping task
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Scroll to load all data
			for i := 0; i < 10; i++ {
				if err := chromedp.Run(ctx, chromedp.Evaluate(`window.scrollBy(0, document.body.scrollHeight)`, nil)); err != nil {
					return err
				}
				time.Sleep(2 * time.Second)
			}
			return nil
		}),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll("div.hidden.md\\:w-5.lg\\:table-cell a"))
			.map(a => a.href)
		`, &hrefs),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll("script[src*='_ssgManifest.js']"))
			.map(script => script.src)
			[0]
		`, &dynamicValue),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Extract the dynamic part from the script src URL
	dynamicValue = extractDynamicValue(dynamicValue)
	if dynamicValue == "" {
		log.Fatal("Failed to extract dynamic value from script src")
	}

	// Process slugs
	var slugs []string
	for _, href := range hrefs {
		parts := strings.Split(strings.TrimRight(href, "/"), "/")
		if len(parts) > 0 {
			slugs = append(slugs, parts[len(parts)-1])
		}
	}

	// Fetch and print asset URLs
	for _, slug := range slugs {
		apiURL := fmt.Sprintf("https://immunefi.com/_next/data/%s/bug-bounty/%s/information.json", dynamicValue, slug)
		assets, err := fetchAssetURLs(apiURL, assetType)
		if err != nil {
			log.Printf("Error fetching data for slug %s: %v", slug, err)
			continue
		}
		// Print asset URLs in the desired format
		for _, assetURL := range assets {
			fmt.Println(assetURL + ", " + "https://immunefi.com/bug-bounty/" + slug + "/information")
		}
	}
}

// extractDynamicValue extracts the dynamic value from the script src URL.
func extractDynamicValue(scriptSrc string) string {
	parts := strings.Split(scriptSrc, "/static/")
	if len(parts) < 2 {
		return ""
	}
	dynamicValueParts := strings.Split(parts[1], "/")
	if len(dynamicValueParts) < 1 {
		return ""
	}
	return dynamicValueParts[0]
}

// fetchAssetURLs fetches and returns asset URLs of the specified type for a given API URL.
func fetchAssetURLs(url, assetType string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		PageProps struct {
			Bounty struct {
				Assets []struct {
					Type string `json:"type"`
					URL  string `json:"url"`
				} `json:"assets"`
			} `json:"bounty"`
		} `json:"pageProps"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	// Filter assets by type or fetch all if assetType is "all"
	var assetURLs []string
	for _, asset := range response.PageProps.Bounty.Assets {
		if assetType == "all" || asset.Type == assetType {
			assetURLs = append(assetURLs, asset.URL)
		}
	}
	return assetURLs, nil
}
