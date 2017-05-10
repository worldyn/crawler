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
	* Database Update
	*/
	go func() {
		for {
			updateWaiter()
			update(session)
		}
	}()

	/*
	* Database pruning
	*/
	go func() {
		for {
			pruneWaiter()
			prune(session)
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

}

// Wait a certain amount of time. Decides how often the db will update
func updateWaiter() {
	<-time.After(15 * time.Minute)
}

// Wait a certain amount of time. Decides how often the db will prune
func pruneWaiter() {
	<-time.After(10 * time.Second)
}

// update mongo db by scraping all the house listings
func update(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()
	db := session.DB("crawler")
	c := db.C("listings")

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

// Prune mongo db by checking if listings links still are available
func prune(s *mgo.Session) {
	fmt.Println("Initialized database prune...")
	session := s.Copy()
	defer session.Close()
	db := session.DB("crawler")
	c := db.C("listings")

	var listings []scrape.Listing

	// Get all listings in db and check if what statuscode you get from their ListingLink
	// If you get 404, remove the listing from the database
  err := c.Find(nil).All(&listings)
  if err != nil {
		fmt.Println("ERROR: mongo find() error")
		fmt.Println("MSG:", err.Error())
  } else {
		for _,l := range listings {
			link := l.ListingLink
			fmt.Println(link)

			resp, httpErr := http.Get(link)
			if httpErr != nil {
				fmt.Println("ERROR: http get error")
				fmt.Println("MSG:", httpErr.Error())
				return
			}

			defer resp.Body.Close()

			status := resp.StatusCode
			if status == 404 {
				fmt.Println("Found listing link with 404 status. Removing listing...")

				removeErr := c.Remove(bson.M{"_id": link})
				if removeErr != nil {
					fmt.Println("ERROR: mongo removal error")
					fmt.Println("MSG:", removeErr.Error())
					return
				}
			}
		}
	}

	fmt.Println("Database prune done...")
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
