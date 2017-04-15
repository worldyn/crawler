package main

import (
	"crawler/scrape"
	"crawler/scrapeImplementations"
	"fmt"
	"github.com/gorilla/mux"
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
	router := mux.NewRouter().StrictSlash(true)
  router.HandleFunc("/", getListings(session))
  log.Fatal(http.ListenAndServe(":8080", router))
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

	chListings := make(chan scrape.Listing)
	chDone := make(chan bool)

	urlCount := 1

	akScraper := scrapeImplementations.AkKvartScraper { "http://akademiskkvart.se", "/?limit=500" }
	go scrape.ParseAndScrape(akScraper, chListings, chDone)

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

// GET request where you get listings from mongo database
func getListings(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()
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
