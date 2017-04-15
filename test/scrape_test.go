package test

import (
	"testing"
	"fmt"

	"crawler/scrapeImplementations"
	"crawler/scrape"
)


func TestAkKvartScraper(t *testing.T) {

	correctListings := correctMap()

	akScraper := scrapeImplementations.AkKvartScraper { "testdata/akKvart.html", "" }

	scrapers := []scrape.SiteScraper{akScraper}

	// Make sure every listing found is supposed to be there, fail otherwise.
	scrape.ParseAndScrapeMultiple(scrapers, func(l scrape.Listing) {
		isCorrect := false

		for cl, _ := range correctListings {
			if compareListings(cl, l) { // if they have the same info
				isCorrect = true
				delete(correctListings, cl)
				break
			}
		}

		if !isCorrect {
			fmt.Println("Found a listing that shouldn't have been there")
			t.Fail()
		}
	})

	if length := len(correctListings); length > 0 {
		fmt.Printf("%d listings were missed.\n", length)
		t.Fail()
	}

	// Nothing was found that wasn't supposed to be found and everyting that was
	// supposed to be found was found: success.
}

func compareListings(a scrape.Listing, b scrape.Listing) bool {
	linkOk := a.ListingLink == b.ListingLink
	addressOk := a.Address == b.Address
	priceOk := a.Price == b.Price
	pubDateOk := a.PublishedDate == b.PublishedDate
	imgUrlOk := a.ImageUrl == b.ImageUrl

	return linkOk && addressOk && priceOk && pubDateOk && imgUrlOk
}


func correctMap() map[scrape.Listing]bool {
	return map[scrape.Listing]bool {
		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104982",
			Address: "Solhemsbackarna 24",
			Price: "4000",
			PublishedDate: "2017-04-04",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/66a3efa02890769510753febdb9ff7668e380ac56db6256dd0bcd2b00097cd7c.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104925",
			Address: "Hagvägen 20",
			Price: "5000",
			PublishedDate: "2017-04-04",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/default_main_photo_100000.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/105042",
			Address: "Spånga vägen",
			Price: "10000",
			PublishedDate: "2017-04-11",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/63c85a458db67cc5fa6d9c60edb9d4ee7ed6a040f1af1c7efed743d5a4215e6b.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/105001",
			Address: "Forsvägen",
			Price: "4200",
			PublishedDate: "2017-04-10",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/c8fc0eb6a21a055dcb5d2f74db1cc79c330f8b1373a055a332cf55e24fc36f51.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/105039",
			Address: "Sätra Torg 16, LGH 1201",
			Price: "4000",
			PublishedDate: "2017-04-10",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/5fbc50998e88b35c257817e217edab1e5ce5483e1aeaae03727a197c48ae81c8.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/105045",
			Address: "Magnus Ladulåsgatan 16",
			Price: "8000",
			PublishedDate: "2017-04-11",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/c65fa3b689cef8e96b5558b4a2decb9026a43c74dff12ae2cd7958be4aa780c8.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104978",
			Address: "regeringsgatan 93",
			Price: "4000",
			PublishedDate: "2017-04-04",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/db78e67cf0cfcd36b69bd2b2843ebf4c77bbaaf32044407747df49368ed9a0fd.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104981",
			Address: "Kristinehamnsgatan 81",
			Price: "3800",
			PublishedDate: "2017-04-04",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/45e128257394387f6ce7a031e4cabc7fef33993c28db669068a2b2df9cd74ea7.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104683",
			Address: "Sockenvägen 366",
			Price: "3000",
			PublishedDate: "2017-04-10",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/default_main_photo_100000.jpg",
		}: true,

		scrape.Listing {
			ListingLink: "testdata/akKvart.html/classifieds/view/104192",
			Address: "Henriksdalsringen 29",
			Price: "4000",
			PublishedDate: "2017-04-10",
			ImageUrl: "testdata/akKvart.html/uploads_dump/thumbnails/default_main_photo_100000.jpg",
		}: true,
	}
}

