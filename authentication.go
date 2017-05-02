package main

import (
	"fmt"
	"net/http"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)


type apiEntry struct {
	ID bson.ObjectId	`bson:"_id json:"_id"`
	KeyString string	`bson:"keyString" json:"keyString"`
	Enabled bool		`bson:"enabled" json:"enabled"`
	Label string			`bson:"label" json:"label"`
}

// Make sure the request is authenticated with a valid (enabled) api key.
// Returns true if proper key was passed, false otherwise.
func Authenticate(s *mgo.Session, r *http.Request) bool {
	values := r.URL.Query()
	apiKeys := values["apikey"]

	if len(apiKeys) != 1 {
		fmt.Println("No api key passed")
		return false
	}

	apiKey := apiKeys[0]

	fmt.Println("DBG: apikey=", apiKey)

	var resSlice []apiEntry
	err := s.DB("crawler").C("apiKeys").Find(bson.M{"keyString": apiKey}).All(&resSlice)

	if err != nil {
		fmt.Println("Error querying db for api key: ", err)
		return false
	}

	if len(resSlice) > 1 {
		fmt.Println("Multiple matches for api key. Should never happen!")
		return false
	} else if len(resSlice) < 1 {
		fmt.Printf("No such api key %s. Not auth!\n", apiKey)
		return false
	}

	res := resSlice[0]

	if ! res.Enabled {
		fmt.Printf("This api key is disabled: %s. Not auth!\n", apiKey)
		return false
	}

	return true
}
