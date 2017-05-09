package main

import (
	"fmt"
	"net/http"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"crawler/scrape"
	"log"
	"encoding/json"
)

// Start listening for incoming connections and handle the requests.
func serveApi(session *mgo.Session) {
	http.HandleFunc("/api", getListings(session))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Coulnd't listenAndServe")
		panic(err)
	}
}




// Creates a query based off the GET variables in the http request.
func createQuery(r *http.Request) bson.M {
	values := r.URL.Query()
	oldestDates := values["noolderthan"]

	// No oldest date provided: returing empty query which will return
	// all listings from db.
	if(len(oldestDates) < 1) {
		return bson.M{}
	}

	// If multple 'noolderthan'-dates are passed the first one is used,
	// the rest are ignored.
	oldestDateStr := oldestDates[0];

	oldestDate, timeParseErr := time.Parse("2006-01-02", oldestDateStr)

	// The date passed could not be parsed. Returning empty query.
	if timeParseErr != nil {
		fmt.Println("Bad date string for 'noolderthan':", oldestDateStr)
		return bson.M{}
	}

	// The resulting query
	q := bson.M{
		"publishedDate": bson.M{
			"$gte": oldestDate,
		},
	}

	return q
}



// GET request where you get listings from mongo database
func getListings(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		fmt.Printf("Request made with url %s\n", r.URL.String())

		if ! authenticate(s, r) {
			// Request not authenticated
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("No valid API key passed!"))
			return
		}

		fmt.Println("api key ok")

		c := session.DB("crawler").C("listings")

		var listings []scrape.Listing
		q := createQuery(r)
		findErr := c.Find(q).All(&listings)
		if findErr != nil {
		  errorWithJSON(w, "Database error", http.StatusInternalServerError)
		  log.Println("Failed get all books: ", findErr)
		  return
		}

		respBody, respErr := json.MarshalIndent(listings, "", "  ")
		if respErr != nil {
		  log.Fatal(respErr)
		}

		responseWithJSON(w, respBody, http.StatusOK)
	}
}

// Api response logic
func errorWithJSON(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    fmt.Fprintf(w, "{message: %q}", message)
}

func responseWithJSON(w http.ResponseWriter, json []byte, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(json)
}

