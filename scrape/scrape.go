package scrape

import (
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"bytes"
	"time"
)

//// Information about each listing on some website
type Listing struct {
	ListingLink string			`bson:"_id" json:"id"`
	Address string					`bson:"address" json:"addess"`
	Price string						`bson:"price" json:"price"`
	PublishedDate time.Time	`bson:"publishedDate" json:"publishedDate"`
	ImageUrl string					`bson:"imageUrl" json:"imageUrl"`
	Contract string					`bson:"contract" json:"contract"`
	Area string							`bson:"area" json:"area"`
	Size string							`bson:"size" json:"size"`
}

// An interface for scarping different house rental websites
type SiteScraper interface {
	Url() string
	Scrape(doc *goquery.Document, ch chan<- Listing)
	FillListing(s *goquery.Selection) Listing
}

// Initialize scraping of some site
func ParseAndScrape(scraper SiteScraper, listingCh chan Listing, done chan<- bool) {
	fmt.Println("Initializing parseAndScrape...")
	defer func() {
		done <- true
	}()

	url := scraper.Url()

	fmt.Println("Beginning html retrivial...")
	jsOut, err := ReadSite(url)


	if err != nil {
		fmt.Println("ERROR: error phantomjs, \"" + url + "\"")
		fmt.Println("MSG:", err.Error())
		return
	}
	fmt.Println("Html retrivial okay...")

	docReader := bytes.NewReader(jsOut)

	doc, parseErr := goquery.NewDocumentFromReader(docReader)

	if parseErr != nil {
		fmt.Println("ERROR: goquery parsing error, \"" + url + "\"")
		fmt.Println("MSG:", err.Error())
		panic(parseErr)
	}

	scraper.Scrape(doc, listingCh)
	fmt.Println("parseAndScrape finished...")
}

// Concurrently scrapes using all scrapers in the passed slice. Every found
// Listing is passed to the handler.
func ParseAndScrapeMultiple(scrapers []SiteScraper, handler func(Listing)) {
	chListings := make(chan Listing)
	chDone := make(chan bool)

	scrapersCount := len(scrapers)

	for _, s := range scrapers {
		go ParseAndScrape(s, chListings, chDone)
	}

	for scrapersDone := 0; scrapersDone < scrapersCount; {
		select {
		case listing := <-chListings:
			handler(listing)
		case <-chDone:
			scrapersDone++
		}
	}

	fmt.Println("Mongodb insertion done...")

}
