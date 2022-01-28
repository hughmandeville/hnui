package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	hn "github.com/hughmandeville/hnui/pkg/hackernews"
	"github.com/otiai10/opengraph/v2"
)

var (
	numStories int
	outFile    string
	verbose    bool
)

// Get top 70 Hacker News stories. If there are no errors, writes to tn_topstories.json.
// Calls the Hacker News API.
//   https://github.com/HackerNews/API
// Uses a Go library to get additional Open Graph data for the article (image, icon, and publisher).
//   https://github.com/otiai10/opengraph
// To Do:
//   - Support using the previous file as a cache for the OG values.
//   - Set timeout on Open Graph fetch.
//   - Setup cron to update data every 10 minutes.
//   - Set user agent when calling URLs.
//   - Add sanitfy check of data.
func main() {
	start := time.Now()

	// Parse command line flags.
	flag.IntVar(&numStories, "num", 70, "number of top stories to get")
	flag.StringVar(&outFile, "out", "hn_topstories.json", "output file JSON")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.Parse()

	if verbose {
		fmt.Printf("Get Hacker News Top Stories\n")
		fmt.Printf("---------------------------\n")
		fmt.Printf("Out File:    %s\n", outFile)
		fmt.Printf("Num Stories: %d\n\n", numStories)
	}

	// Get top stories from Hacker News.
	items, err := hn.GetTopStories(numStories)
	if err != nil {
		log.Fatalf("Problem getting top stories: %s", err)
		return
	}

	// Add Open Graph data.
	for i := 0; i < len(items); i++ {
		time.Sleep(100 * time.Millisecond)
		addOGData(&items[i])
		if verbose {
			fmt.Printf(" %9d  %-20s  %s\n", items[i].ID, items[i].Publisher, items[i].Title)
		}
	}

	if len(items) < 10 {
		fmt.Printf("Hacker News API returned less than 10 stories, so not writing to %s.\n", outFile)
		return
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		log.Fatalf("Problem marshalling items: %s", err)
		return
	}

	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		log.Fatalf("Problem saving to file: %s", err)
		return
	}
	if verbose {
		fmt.Println()
		fmt.Printf("Wrote:       %s (%d items, %d bytes).\n", outFile, len(items), len(data))
		fmt.Printf("Took:        %s\n", time.Since(start))
		fmt.Println()
	}
}

// Add Open Graph data to the item (image, icon, and publisher).
// https://pkg.go.dev/github.com/otiai10/opengraph
func addOGData(item *hn.Item) (err error) {

	// Get URL's domain name and remove www.
	domain := ""
	pu, err := url.Parse(item.URL)
	if err == nil {
		domain = strings.TrimPrefix(pu.Hostname(), "www.")
	}

	// set publisher to the URL's domain name by default
	if item.Publisher == "" {
		item.Publisher = domain
	}

	// TBD: set timeout.
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	ogp, err := opengraph.Fetch(item.URL)
	if err != nil {
		return
	}

	// Set icon.
	item.Icon = sanitizeURL(item.URL, ogp.Favicon.URL)

	// Set image.
	if len(ogp.Image) > 0 {
		item.Image = sanitizeURL(item.URL, ogp.Image[0].URL)
	}

	// Set publisher.
	publisher := strings.TrimSpace(ogp.SiteName)
	if publisher != "" {
		item.Publisher = publisher
	}

	// Fix bad data.
	correctData(item, domain)

	item.OGDescription = ogp.Description
	item.OGTitle = ogp.Title
	return
}

// Turn relative URLs into absolute URLs (/foo/bar.jpg -> https://example.com/foo/bar.jpg).
func sanitizeURL(parentURL string, childURL string) (sanitizedURL string) {
	sanitizedURL = strings.TrimSpace(childURL)
	if sanitizedURL == "" || strings.HasPrefix(sanitizedURL, "http:") || strings.HasPrefix(sanitizedURL, "https:") {
		return
	}
	if strings.HasPrefix(childURL, "//") {
		sanitizedURL = fmt.Sprintf("https:%s", childURL)
		return
	}
	pu, err := url.Parse(parentURL)
	if err != nil {
		return
	}
	if strings.HasPrefix(childURL, "/") {
		sanitizedURL = fmt.Sprintf("%s://%s%s", pu.Scheme, pu.Hostname(), childURL)
		return
	}
	path := pu.Path
	pi := strings.LastIndex(path, "/")
	if pi > 0 {
		path = path[:pi]
	}
	sanitizedURL = fmt.Sprintf("%s://%s%s/%s", pu.Scheme, pu.Hostname(), path, childURL)
	return
}

// Fix known images with icon, image, and publisher data.
func correctData(item *hn.Item, domain string) {
	// set icon if missing for some well known publishers
	if item.Icon == "" {
		switch strings.ToLower(domain) {
		case "npr.org":
			item.Icon = "https://www.npr.org/favicon.ico"
		case "ourworldindata.org":
			item.Icon = "https://ourworldindata.org/favicon.ico"
		case "wpr.org":
			item.Icon = "https://www.wpr.org/sites/default/files/favicon_0_0.ico"
		}
	}

	// fix broken icons of some well known publishers
	switch item.Icon {
	case "https://www.bloomberg.com/favicon.ico":
		item.Icon = "https://assets.bwbx.io/s3/javelin/public/hub/images/favicon-black-63fe5249d3.png"
	case "https://news.ycombinator.com/item/favicon.ico":
		item.Icon = "https://news.ycombinator.com/favicon.ico"
	}

	// fix proublisher name for some well known publishers
	switch strings.ToLower(item.Publisher) {
	case "bbc.com":
		item.Publisher = "BBC"
	case "bloomberg.com":
		item.Publisher = "Bloomberg"
	case "business-standard.com":
		item.Publisher = "Business Standard"
	case "hudsonreview.com":
		item.Publisher = "The Hudson Review"
	case "kaggle.com":
		item.Publisher = "Kaggle"
	case "nasdaq.com":
		item.Publisher = "Nasdaq"
	case "nature.com":
		item.Publisher = "Nature"
	case "nytimes.com":
		item.Publisher = "The New York Times"
	case "thelocal.com":
		item.Publisher = "The Local"
	case "vice.com":
		item.Publisher = "Vice"
	}

	// Shorten long publisher names with pipe symbol (|) by removing text after pipe symbol.
	i := strings.Index(item.Publisher, "|")
	if len(item.Publisher) > 20 && i > 3 {
		item.Publisher = strings.TrimSpace(item.Publisher[:i-1])
	}
}
