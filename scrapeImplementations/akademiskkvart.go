package scrapeImplementations

import (
	"crawler/scrape"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"github.com/PuerkitoBio/goquery"
)

// akademiskkvart.se
type AkKvartScraper struct {
	UrlBase string
	UrlParams string
}

func (akt AkKvartScraper) Url() string {
	return akt.UrlBase + akt.UrlParams
}

func (akt AkKvartScraper) FillListing(s *goquery.Selection) scrape.Listing {
	listing := scrape.Listing{}

	// Only address is extracted...

	// Child divs
	imgDiv := s.Children().First()
	infoDiv := s.Children().First().Next()

	// Listing link
	listingLink, linkHrefExists := imgDiv.Find("a").Attr("href")
	if linkHrefExists {
		listing.ListingLink = akt.UrlBase+strings.Trim(listingLink, " ")
	}

	// Image URL
	imageUrl, imgSrcExists := imgDiv.Find("img").Attr("thumb")
	if imgSrcExists {
		listing.ImageUrl = akt.UrlBase+strings.Trim(imageUrl, " ")
	}

	// Address
	address := infoDiv.Find("h3 a")
	if address.Length() > 0 {
		listing.Address = strings.Trim(address.Text(), " ")
	}

	// Price
	price := infoDiv.Find("p.price")
	if price.Length() > 0 {
		// Use regex to extract only the price (numbers) from the text
		re := regexp.MustCompile("[0-9]+")
		match := re.FindString(price.Text())
		if len(match) > 0 {
			listing.Price = match
		}
	}

	// Published date
	publishedDate := infoDiv.Find("ul").Children().Last().Children().Last()
	if publishedDate.Length() > 0 {
		// Use regex to extract onlt date in format YYYY-MM-DD from the text
		re := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}")
		match := re.FindString(publishedDate.Text())
		if len(match) > 0 {
			listing.PublishedDate = match
		}
	}

	return listing
}

// Scrape listings on akademiskkvart.se as Listing struct
func (akt AkKvartScraper) Scrape(doc *goquery.Document, ch chan<- scrape.Listing) {
	fmt.Println("Beginning to scrape...")
	var wg sync.WaitGroup

	findings := doc.Find("#listings li.template")

	findings.Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			ch <- akt.FillListing(s)
			wg.Done()
		}(&wg)
	})

	wg.Wait()
	fmt.Println("Scrape finished...")
}
