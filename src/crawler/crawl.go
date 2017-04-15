package main

import (
	"fmt"
	//"golang.org/x/net/html"
	"github.com/PuerkitoBio/goquery"
	//"net/http"
	//"encoding/json"
	"sync"
	"strings"
  //"html"
  //"log"
  //"github.com/gorilla/mux"
	//"time"
	"bytes"
	"os/exec"
)

// Information about each listing on some website
type Listing struct {
	address string
	price string
	publishedDate string
	imageUrl string
}

// An interface for scarping different house rental websites
type SiteScraper interface {
	scrape(doc *goquery.Document, ch chan<- Listing)
	fillListing(s *goquery.Selection) Listing
}

// akademiskkvart.se
type AkKvartScraper struct{}

func (akt AkKvartScraper) fillListing(s *goquery.Selection) Listing {
	listing := Listing{}

	// Child divs
	imgDiv := s.Children().First()
	infoDiv := s.Children().First().Next()

	// Image URL
	imageUrl, imgSrcExists := imgDiv.Find("img").Attr("src")
	if imgSrcExists {
		listing.imageUrl = strings.Trim(imageUrl, " ")
	}

	// Address
	address := infoDiv.Find("h3 a")
	if address.Length() > 0 {
		listing.address = strings.Trim(address.Text(), " ")
	}

	// Price
	price := infoDiv.First().Next().First().Find("p.price")
	if price.Length() > 0 {
		listing.price = strings.Trim(price.Text(), " ")
	}

	// Published date
	publishedDate := infoDiv.Find("ul").First().Next().Next().Next().Next().Next().Children().First()
	if publishedDate.Length() > 0 {
		listing.publishedDate = strings.Trim(publishedDate.Text(), " ")
	}

	return listing
}

// Scrape listings on akademiskkvart.se as Listing struct
func (akt AkKvartScraper) scrape(doc *goquery.Document, ch chan<- Listing) {
	var wg sync.WaitGroup

	findings := doc.Find("#listings li.template")

	findings.Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			ch <- akt.fillListing(s)
			wg.Done()
		}(&wg)
	})

	wg.Wait()
}

// Initialize scraping of some site
func crawl(url string, scraper SiteScraper, listingCh chan Listing, done chan<- bool) {
	defer func() {
		done <- true
	}()

	jsOut, err := exec.Command("phantomjs", "content.js", url).Output()

	if err != nil {
		fmt.Println("ERROR: error phantomjs, \"" + url + "\"")
		fmt.Println("MSG:", err.Error())
		return
	}

	docReader := bytes.NewReader(jsOut)

	doc, parseErr := goquery.NewDocumentFromReader(docReader)

	if parseErr != nil {
		fmt.Println("ERROR: goquery parsing error, \"" + url + "\"")
		fmt.Println("MSG:", err.Error())
		return
	}

	scraper.scrape(doc, listingCh)
}

func main() {
	chListings := make(chan Listing)
	chDone := make(chan bool)

	//tmp
	urlCount := 1

	akScraper := AkKvartScraper{}
	go crawl("http://akademiskkvart.se/?limit=500", akScraper, chListings, chDone)

	for crawlersDone := 0; crawlersDone < urlCount; {
		select {
		case listing := <-chListings:
				fmt.Printf("Found listing at %s.\n", listing.address)
		case <-chDone:
			crawlersDone++
		}
	}

}
  /*router := mux.NewRouter().StrictSlash(true)
/*
func Index(w http.ResponseWriter, r *http.Request) {
 chListingsprintf(w, "HelListing", html.EscapeString(r.URL.Path))
}*/
