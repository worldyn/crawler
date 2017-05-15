package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"time"
  "crawler/scrape"
	"net/http"
	"gopkg.in/mgo.v2/bson"
)

// Runs indefinitely. Sleeps for some time, then looks for invalid urls and prunes database
func pruneCycle(session *mgo.Session) {
	for {
		pruneWaiter()
		prune(session)
	}
}

// Wait a certain amount of time. Decides how often the db will prune
func pruneWaiter() {
	<-time.After(30 * time.Second)
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
