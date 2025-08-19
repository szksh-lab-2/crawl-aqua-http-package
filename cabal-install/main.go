package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c := colly.NewCollector(
		colly.AllowedDomains("downloads.haskell.org"),
	)

	assets := map[string][]string{} // version => assetNames

	urlPattern := regexp.MustCompile("^https://downloads.haskell.org/~cabal/cabal-install-([0-9.]+)/")

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.HasPrefix(link, "../") {
			return
		}
		url := e.Request.AbsoluteURL(link)
		if !strings.HasPrefix(url, "https://downloads.haskell.org/~cabal/cabal-install-") {
			return
		}
		u := urlPattern.FindStringSubmatch(url)
		if u == nil {
			return
		}
		version := u[1]
		// Print link
		log.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		if strings.HasSuffix(link, "/") {
			if err := c.Visit(url); err != nil {
				log.Println("Error visiting link:", err)
			}
			return
		}
		a, ok := assets[version]
		if !ok {
			a = []string{}
			assets[version] = a
		}
		assets[version] = append(a, path.Base(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.Visit("https://downloads.haskell.org/~cabal/")
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(assets); err != nil {
		return err
	}

	return nil
}
