package main

import (
	"fmt"
	"crawler/scrape"
	"crawler/scrapeImplementations"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type counterEntry struct {
	CounterName string	`bson:"counterName" json:"counterName"`
	SeqNumber int				`bson:"seqNumber" json:"seqNumber"`
}


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
		listing.SeqNumber = nextSeqNumber(s)
		insertErr := c.Insert(listing)
		//fmt.Println(json.Marshal(&listing))

		if insertErr != nil &&  !mgo.IsDup(insertErr) {
			fmt.Println("ERROR: mongo insert error")
			fmt.Println("MSG:", insertErr.Error())
			panic(insertErr)
		}
	});
}

func nextSeqNumber(s *mgo.Session) int {
	session := s.Copy()
	defer session.Close()
	c := session.DB("crawler").C("counters")

	findQuery := bson.M{
		"counterName": "listingCounter",
	}

	var counterRes counterEntry
	errFind := c.Find(findQuery).One(&counterRes)

	if errFind != nil {
		fmt.Println("Coulnd't retrieve listings counter:", errFind)
		panic(errFind)
	}

	ret := counterRes.SeqNumber

	counterRes.SeqNumber++

	errUpdate := c.Update(findQuery, counterRes)

	if errUpdate != nil {
		fmt.Println("Error updating listings counter:", errUpdate)
		panic(errUpdate)
	}

	return ret
}
