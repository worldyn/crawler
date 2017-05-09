package main

import (
	"fmt"
	"crawler/scrape"
	"crawler/scrapeImplementations"
	"gopkg.in/mgo.v2"
	"time"
)

// Runs indefinitely. Sleeps for some time, then scrapes and inserts new
// listings into db. Then repeats.
func updateCycle(session *mgo.Session) {
	for {
		updateWaiter()
		update(session)
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


	akScraper := scrapeImplementations.AkKvartScraper {
		"http://akademiskkvart.se", "/?limit=500",
	}
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

