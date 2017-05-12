package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"time"
	"crawler/scrape"
	"net/http"
	"gopkg.in/mgo.v2/bson"
)

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

	setupListings(session)

	// Database pruning
	go func() {
		for {
			pruneWaiter()
			prune(session)
		}
	}()

	// Database Updating
	go updateCycle(session)

	// REST API logic
	serveApi(session)
}

// MongoDB setup settings
func setupListings(s *mgo.Session) {
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

// Wait a certain amount of time. Decides how often the db will prune
func pruneWaiter() {
	<-time.After(10 * time.Second)
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
