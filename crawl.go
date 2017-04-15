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
	"regexp"
	"gopkg.in/mgo.v2"
  //"gopkg.in/mgo.v2/bson"
)

// Information about each listing on some website
type Listing struct {
	ListingLink string 		`bson:"_id" json:"id"`
	Address string				`bson:"address" json:"addess"`
	Price string					`bson:"price" json:"price"`
	PublishedDate string	`bson:"publishedDate" json:"publishedDate"`
	ImageUrl string				`bson:"imageUrl" json:"imageUrl"`
}

// An interface for scarping different house rental websites
type SiteScraper interface {
	url() string
	scrape(doc *goquery.Document, ch chan<- Listing)
	fillListing(s *goquery.Selection) Listing
}

// akademiskkvart.se
type AkKvartScraper struct {
	urlBase string
	urlParams string
}

func (akt AkKvartScraper) url() string {
	return akt.urlBase + akt.urlParams
}

func (akt AkKvartScraper) fillListing(s *goquery.Selection) Listing {
	listing := Listing{}

	// Only address is extracted...

	// Child divs
	imgDiv := s.Children().First()
	infoDiv := s.Children().First().Next()

	// Listing link
	listingLink, linkHrefExists := imgDiv.Find("a").Attr("href")
	if linkHrefExists {
		listing.ListingLink = akt.urlBase+strings.Trim(listingLink, " ")
	}

	// Image URL
	imageUrl, imgSrcExists := imgDiv.Find("img").Attr("thumb")
	if imgSrcExists {
		listing.ImageUrl = akt.urlBase+strings.Trim(imageUrl, " ")
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
func (akt AkKvartScraper) scrape(doc *goquery.Document, ch chan<- Listing) {
	fmt.Println("Beginning to scrape...")
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
	fmt.Println("Scrape finished...")
}

// Initialize scraping of some site
func parseAndScrape(scraper SiteScraper, listingCh chan Listing, done chan<- bool) {
	fmt.Println("Initializing parseAndScrape...")
	defer func() {
		done <- true
	}()

	url := scraper.url()

	fmt.Println("Beginning html retrivial...")
	jsOut, err := exec.Command("phantomjs", "content.js", url).Output()


	if err != nil {
		fmt.Println("ERROR: error phantomjs, \"" + scraper.url() + "\"")
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

	scraper.scrape(doc, listingCh)
	fmt.Println("parseAndScrape finished...")
}

func main() {
	// MONGO SETUP
	session, conErr := mgo.Dial("127.0.0.1:27017")
	if conErr != nil {
		fmt.Println("ERROR: mongo connection error")
		fmt.Println("MSG:", conErr.Error())
		panic(conErr)
	}
	fmt.Println("Mongodb connected...")

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	// Collection People
	c := session.DB("crawler").C("listings")

	// Index
	index := mgo.Index{
		Key:        []string{"ListingLink"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	indexErr := c.EnsureIndex(index)
	if indexErr != nil {
		fmt.Println("ERROR: mongo index error")
		fmt.Println("MSG:", indexErr.Error())
		panic(indexErr)
	}

	// CRAWLING
	chListings := make(chan Listing)
	chDone := make(chan bool)

	//tmp
	urlCount := 1

	akScraper := AkKvartScraper { "http://akademiskkvart.se", "/?limit=500" }
	go parseAndScrape(akScraper, chListings, chDone)

	for crawlersDone := 0; crawlersDone < urlCount; {
		select {
		case listing := <-chListings:
			insertErr := c.Insert(listing)
			//fmt.Println(json.Marshal(&listing))

			if insertErr != nil &&  !mgo.IsDup(insertErr) {
				fmt.Println("ERROR: mongo insert error")
				fmt.Println("MSG:", insertErr.Error())
				panic(insertErr)
			}

		case <-chDone:
			crawlersDone++
		}
	}
	fmt.Println("Mongodb insertion done...")



}
  /*router := mux.NewRouter().StrictSlash(true)
/*
func Index(w http.ResponseWriter, r *http.Request) {
 chListingsprintf(w, "HelListing", html.EscapeString(r.URL.Path))
}*/
