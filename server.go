package main

import (
	"fmt"
	"gopkg.in/mgo.v2"
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

	// Concurrent Database Updating
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

