package main

import (
	"crawler/scrape"
	"crawler/scrapeImplementations"
	"fmt"
	"net/http"
	"log"
	"encoding/json"
	"gopkg.in/mgo.v2"
	"time"
  "gopkg.in/mgo.v2/bson"
)

func main() {
	/*
	* MONGO SETUP
	*/
	session, conErr := mgo.Dial("127.0.0.1:27017")
	if conErr != nil {
		fmt.Println("ERROR: mongo connection error")
		fmt.Println("MSG:", conErr.Error())
		panic(conErr)
	}
	fmt.Println("Mongodb connected...")

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	ensureIndex(session)

	/*
	* Concurrent Database Updating
	*/

	go func() {
		for {
			updateWaiter()
			update(session)
		}
	}()

	/*
	* REST API
	*/
	 http.HandleFunc("/api", getListings(session))
	 err := http.ListenAndServe(":8080", nil)
	 if err != nil {
		fmt.Println("Coulnd't listenAndServe")
		panic(err)
	 }
}

// MongoDB setup settings
func ensureIndex(s *mgo.Session) {
    session := s.Copy()
    defer session.Close()

    c := session.DB("crawler").C("listings")
		cTemp := session.DB("crawler").C("listingsTemp")

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

		indexErr2 := cTemp.EnsureIndex(index)
		if indexErr2 != nil {
			fmt.Println("ERROR: mongo index error")
			fmt.Println("MSG:", indexErr2.Error())
			panic(indexErr2)
		}
}

// Wait a certain amount of time. Decides how often the db will update
func updateWaiter() {
	<-time.After(20 * time.Second)
}

// update mongo db by scraping all the house listings
func update(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()
	db := session.DB("crawler")
	c := db.C("listings")
	//cTemp := db.C("listingsTemp")

	akScraper := scrapeImplementations.AkKvartScraper { "http://akademiskkvart.se", "/?limit=500" }
	scrapers := []scrape.SiteScraper{akScraper}
	scrape.ParseAndScrapeMultiple(scrapers, func(listing scrape.Listing) {
		insertErr := c.Insert(listing)

		if insertErr != nil && !mgo.IsDup(insertErr) {
			fmt.Println("ERROR: mongo insert error")
			fmt.Println("MSG:", insertErr.Error())
			panic(insertErr)
		}
	});
}

// Creates a query based off the GET variables in the http request.
func createQuery(r *http.Request) bson.M {
	values := r.URL.Query()
	oldestDates := values["noolderthan"]

	// No oldest date provided: returing empty query which will return
	// all listings from db.
	if(len(oldestDates) < 1) {
		return bson.M{}
	}

	// If multple 'noolderthan'-dates are passed the first one is used,
	// the rest are ignored.
	oldestDateStr := oldestDates[0];

	oldestDate, timeParseErr := time.Parse("2006-01-02", oldestDateStr)

	// The date passed could not be parsed. Returning empty query.
	if timeParseErr != nil {
		fmt.Println("Bad date string for 'noolderthan':", oldestDateStr)
		return bson.M{}
	}

	// The resulting query
	q := bson.M{
		"publishedDate": bson.M{
			"$gte": oldestDate,
		},
	}

	return q
}

// GET request where you get listings from mongo database
func getListings(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		fmt.Printf("Request made with url %s\n", r.URL.String())

		if ! Authenticate(s, r) {
			// Request not authenticated
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("No valid API key passed!"))
			return
		}

		fmt.Println("api key ok")

		c := session.DB("crawler").C("listings")

		var listings []scrape.Listing
		q := createQuery(r)
		findErr := c.Find(q).All(&listings)
		if findErr != nil {
		  ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
		  log.Println("Failed get all books: ", findErr)
		  return
		}

		respBody, respErr := json.MarshalIndent(listings, "", "  ")
		if respErr != nil {
		  log.Fatal(respErr)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

// Api response logic
func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(json)
}
