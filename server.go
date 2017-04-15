package main

import (
	"crawler/scrape"
	"crawler/scrapeImplementations"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"html"
	"log"
	//"encoding/json"
	"gopkg.in/mgo.v2"
	"time"
  //"gopkg.in/mgo.v2/bson"
	//"sync"
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

	// Collection People
	c := session.DB("crawler").C("listings")

	// Index
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

	// CRAWLING DATABASE Updating
	go func() {
		for {
			<-time.After(20 * time.Second)
			chListings := make(chan scrape.Listing)
			chDone := make(chan bool)

			urlCount := 1

			akScraper := scrapeImplementations.AkKvartScraper { "http://akademiskkvart.se", "/?limit=500" }
			go scrape.ParseAndScrape(akScraper, chListings, chDone)

			for crawlersDone := 0; crawlersDone < urlCount; {
				select {
				case listing := <-chListings:
					insertErr := c.Insert(listing)
					//fmt.Println(json.Marshal(&listing))

					if insertErr != nil &&  !mgo.IsDup(insertErr) {
						fmt.Println("ERROR: mongo insert error")
						fmt.Println("MSG:", insertErr.Error())
						panic(insertErr)
					}

				case <-chDone:
					crawlersDone++
				}
			}
			fmt.Println("Mongodb insertion done...")
		}
	}()

	// REST API path
	router := mux.NewRouter().StrictSlash(true)
  router.HandleFunc("/", GetListings)
  log.Fatal(http.ListenAndServe(":8080", router))

}

func GetListings(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

/*func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    w.Write(json)
}*/
