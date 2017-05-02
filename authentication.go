package main

import (
	"fmt"
	"net/http"
	"encoding/json"
)


type apiEntry struct {
	ID bson.ObjectId	`bson:"_id json:"_id"`
	KeyString string	`bson:"keyString" json:"keyString"`
	Enabled string		`bson:"enabled" json:"enabled"`
	Label string			`bson:"label" json:"label"`
}

// Make sure the request is authenticated with a valid (enabled) api key
func Authenticate(s *mgo.Session, r *http.Request) bool {
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

	if err != nil || count != 1 {
		fmt.Println("Error get count")
		return false
	}

	var res apiEntry
	err2 := query.One(&res)

	if err2 != nil {
		fmt.Println("couldn't get struct from query", err2)
		return false
	}

	if ! res.Enabled {
		return false
	}

	return true
}
