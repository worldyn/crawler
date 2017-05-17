package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
)

type counterEntry struct {
	CounterName string	`bson:"counterName" json:"counterName"`
	SeqNumber int				`bson:"seqNumber" json:"seqNumber"`
}


func nextSeqNumber(s *mgo.Session) int {
	session := s.Copy()
	defer session.Close()

	setupCounterCollection(session)

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


// Makes sure the collection with counters exists.
// Also sets listing counter to 1 if it doesn't exist.
func setupCounterCollection(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("crawler").C("counters")

	findQuery := bson.M{
		"counterName": "listingCounter",
	}

	var counterRes counterEntry
	err := c.Find(findQuery).One(&counterRes)

	if err == mgo.ErrNotFound {
		// No counter for listings. Creating one.
		fmt.Println("Creating new counter for listings.")
		newCounterEntry := bson.M {
			"counterName": "listingCounter",
			"seqNumber": 1,
		}
		errIns := c.Insert(newCounterEntry)
		if errIns != nil {
			fmt.Println("Couldn't init counter for listings.")
			panic(errIns)
		}
	}

	// Else: do nothing as the collection and document is already there.

}
