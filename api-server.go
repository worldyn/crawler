package main

import (
	"fmt"
	"strconv"
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


func createDateFilterQuery(c *mgo.Collection, oldestDates []string) *mgo.Query {
	// If multple 'noolderthan'-dates are passed the first one is used,
	// the rest are ignored.
	oldestDateStr := oldestDates[0];

	oldestDate, timeParseErr := time.Parse("2006-01-02", oldestDateStr)

	// The date passed could not be parsed. Returning empty query.
	if timeParseErr != nil {
		fmt.Println("Bad date string for 'noolderthan':", oldestDateStr)
		return c.Find(bson.M{})
	}

	// The resulting query
	q := c.Find(bson.M{
		"publishedDate": bson.M{
			"$gte": oldestDate,
		},
	})

	return q
}

func createSeqFilterQuery(c *mgo.Collection, seqNumbers []string, counts []string) *mgo.Query {
	seqNumberStr := seqNumbers[0]
	seqNumber, errSeq := strconv.Atoi(seqNumberStr)

	// seqNumber not a valid number.
	if errSeq != nil {
		fmt.Println("not a valid seq number ", seqNumberStr)
		return c.Find(bson.M{})
	}

	q := c.Find(bson.M{
		"seqNumber": bson.M{
			"$gt": seqNumber,
		},
	})

	// no count was passed => returing query for all listings with ok seqNumber.
	if len(counts) < 1  {
		return q
	}

	countStr := counts[0]
	count, errCount := strconv.Atoi(countStr)

	// The first count passed was not an ok number. Logging and ignoring.
	if errCount != nil {
		fmt.Println("Not a valid count:", countStr)
		return q
	}

	return q.Sort("seqNumber").Limit(count)
}

// Creates a query based off the GET variables in the http request.
func createQuery(c *mgo.Collection, r *http.Request) *mgo.Query {
	values := r.URL.Query()

	var ret *mgo.Query

	oldestDates := values["noolderthan"]
	seqNumbers := values["afterseqnumber"]

	if len(oldestDates) > 0  {
		ret = createDateFilterQuery(c, oldestDates)
	} else if len(seqNumbers) > 0 {
		ret = createSeqFilterQuery(c, seqNumbers, values["count"])
	}	else {
		ret = c.Find(bson.M{})
	}

	return ret
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
		query := createQuery(c, r)
		findErr := query.All(&listings)
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

