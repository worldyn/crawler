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
	* REST API logic
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
	<-time.After(20 * time.Second)
}

// update mongo db by scraping all the house listings
func update(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()
	c := session.DB("crawler").C("listings")


	akScraper := scrapeImplementations.AkKvartScraper { "http://akademiskkvart.se", "/?limit=500" }
	scrapers := []scrape.SiteScraper{akScraper}
	scrape.ParseAndScrapeMultiple(scrapers, func(listing scrape.Listing) {
		insertErr := c.Insert(listing)
		//fmt.Println(json.Marshal(&listing))

		if insertErr != nil &&  !mgo.IsDup(insertErr) {
			fmt.Println("ERROR: mongo insert error")
			fmt.Println("MSG:", insertErr.Error())
			panic(insertErr)
		}
	});
}

/*
type apiEntry struct {
	keyString string
	enabled bool
	label string
}
*/

// Make sure the request is authenticated with a valid (enabled) api key
func handleApiKey(s *mgo.Session, w http.ResponseWriter, r *http.Request) bool {
	values := r.URL.Query()
	apiKeys := values["apikey"]

	if len(apiKeys) != 1 {
		fmt.Println("No api key passed")
		return false
	}

	apiKey := apiKeys[0]

	fmt.Println("DBG: apikey=", apiKey)

	query := s.DB("crawler").C("apiKeys").Find(bson.M{"keyString": apiKey})

	count, err := query.Count()

	if err != nil {
		fmt.Println("Error get count")
	}

	fmt.Println("Found", count, "matching keys")
	return true
}

// GET request where you get listings from mongo database
func getListings(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		handleApiKey(s, w, r)

		c := session.DB("crawler").C("listings")

		var listings []scrape.Listing
		findErr := c.Find(bson.M{}).All(&listings)
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
