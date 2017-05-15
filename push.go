package main

import (
	"os/exec"
  "gopkg.in/mgo.v2"
  "fmt"
)

// Information about each listing on some website
type Device struct {
  Id string              `bson:"_id" json:"id"`
	DeviceToken string     `bson:"deviceToken" json:"deviceToken"`
}

func InitPush(s *mgo.Session, link string, area string) {
  fmt.Println("InitPush!")
  session := s.Copy()
  defer session.Close()

  db := session.DB("crawler")
	c := db.C("deviceTokensApple")

	var deviceTokens []Device

	// Get all device tokens and send push notifications to them
  err := c.Find(nil).All(&deviceTokens)
  if err != nil {
		fmt.Println("ERROR: mongo find() error")
		fmt.Println("MSG:", err.Error())
  } else {
		for _,d := range deviceTokens {
      deviceToken := d.DeviceToken
      go SendPushApple(link, area, deviceToken)
    }
  }
}

func SendPushApple(listingLink string, area string, deviceToken string) {
  keyId := "77Q6A8M8LP"
  teamId := "LKE8625923"
  status, err := exec.Command("node", "app.js", keyId, teamId, deviceToken, listingLink, area).Output()

  fmt.Println("Status: ", status, "...")
  if err != nil {
    fmt.Println("Push Script error:", err)
  }
}
