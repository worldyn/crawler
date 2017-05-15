package main

import (
	"fmt"
	"strconv"
  "strings"
  "sort"
	"net/http"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
	"crawler/scrape"
	"log"
	"encoding/json"
)


// Relationships
const (
  GT  = iota // Greater than
  GTE = iota // Greater than or equal to
  LT  = iota // Less than
  LTE = iota // Less than or equal to
)


// Start listening for incoming connections and handle the requests.
func serveApi(session *mgo.Session) {
	http.HandleFunc("/api", getListings(session))

	// Uncomment to use HTTPS. Make sure you have a cert.pem and key.pem
	// err := http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)

	// Use HTTP
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Coulnd't listenAndServe")
		panic(err)
	}
}

// Takes a relationship (eg GT) and returns a string used when querying mongo
// (ie "$gt"). Assumes valid relationship is passed. Otherwise empty string is
// returned.
func relToJsonRel(rel int) string {
  switch(rel) {
  case GT:
    return "$gt"
  case GTE:
    return "$gte"
  case LT:
    return "$lt"
  case LTE:
    return "$lte"
  default:
    return ""
  }
}

// Takes a string like "gte:2017-05-01" and returns (GTE, "2017-05-01", true) or
// "lt:100" and returns (LT, "100", true). The last return is always true for
// valid inputs. False otherwise.
func parseFilterArg(arg string) (int, string, bool) {
  parts := strings.Split(arg, ":")

  if len(parts) != 2 {
    return 0, "", false
  }

  var relationship int

  switch relationshipStr := strings.ToLower(parts[0]); relationshipStr {
  case "gt":
    relationship = GT
  case "gte":
    relationship = GTE
  case "lt":
    relationship = LT
  case "lte":
    relationship = LTE
  default:
    return 0, "", false
  }

  return relationship, parts[1], true

}

func createDateFilterQuery(c *mgo.Collection, arg string, count int) *mgo.Query {
  rel, dateStr, success := parseFilterArg(arg)
  if !success {
    fmt.Println("Poorly formatted argument for date filter:", arg)
    return c.Find(bson.M{})
  }

	date, timeParseErr := time.Parse("2006-01-02", dateStr)

	// The date passed could not be parsed. Returning empty query.
	if timeParseErr != nil {
		fmt.Println("Bad date string for date filter:", dateStr)
		return c.Find(bson.M{})
	}

  jsonRel := relToJsonRel(rel)

	q := c.Find(bson.M{
		"publishedDate": bson.M{
			jsonRel: date,
		},
	})


  // Smallest valid count is one.
  // If something smaller is passed there is no restriction.
	if count < 1 {
		return q
	}

  if rel == GT || rel == GTE {
    q = q.Sort("publishedDate")
  } else {
    q = q.Sort("-publishedDate")
  }
  return q.Limit(count)

	return q
}

func createSeqFilterQuery(c *mgo.Collection, arg string, count int) *mgo.Query {
  rel, seqStr, success := parseFilterArg(arg)
  if !success {
    fmt.Println("Poorly formatted argument for sequence filter:", arg)
    return c.Find(bson.M{})
  }

  seqNum, errAtoi := strconv.Atoi(seqStr)
  if errAtoi != nil {
    fmt.Println("Passed seqNum not an integer:", arg)
    return c.Find(bson.M{})
  }

  jsonRel := relToJsonRel(rel)

	q := c.Find(bson.M{
		"seqNumber": bson.M{
			jsonRel: seqNum,
		},
	})

  // Smallest valid count is one.
  // If something smaller is passed there is no restriction.
	if count < 1 {
		return q
	}

  if rel == GT || rel == GTE {
    q.Sort("seqNumber")
  } else {
    q.Sort("-seqNumber")
  }
  return q.Limit(count)
}

// Takes the count get argument(s) and returns an int represeting the passed
// count. Returns -1 if no count is (properly passed).
// The count is only found if it is written as an integer and is the first
// count passed in the url.
func extractCount(countStrs []string) int {
  // Default count. Means no restriction.
  count := -1
  if len(countStrs) >= 1 {
    countStr := countStrs[0]
    countRes, err := strconv.Atoi(countStr)

    // count is a valid number.
    if err == nil {
      count = countRes
    }
    // Else we let it stay -1
  }

  return count
}

// Creates a query based off the GET variables in the http request.
func createQuery(c *mgo.Collection, r *http.Request) *mgo.Query {
	values := r.URL.Query()

	var ret *mgo.Query

  countStrs := values["count"]
	seqFilterStrs := values["seqfilter"]
	dateFilterStrs := values["datefilter"]

  count := extractCount(countStrs)

  if len(seqFilterStrs) >= 1 {
    ret = createSeqFilterQuery(c, seqFilterStrs[0], count)
  } else if len(dateFilterStrs) >= 1 {
    ret = createDateFilterQuery(c, dateFilterStrs[0], count)
  } else {
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

    // Sort listings so the ones with the highest seqNumbers come first
    sort.Sort(ListingsBySeq(listings))

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



// Sorting listings by SeqNumber:
type ListingsBySeq []scrape.Listing

func (listings ListingsBySeq) Len() int {
  return len(listings)
}

func (listings ListingsBySeq) Swap(i int, j int) {
  listings[i], listings[j] = listings[j], listings[i]
}

func (listings ListingsBySeq) Less(i int, j int) bool {
  //return listings[i].SeqNumber < listings[j].SeqNumber
  // We want the listings with higher seqNumber (newer) to come first.
  return listings[i].SeqNumber > listings[j].SeqNumber
}
