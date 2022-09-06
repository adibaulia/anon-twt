package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adibaulia/anon-twt/config"
	"github.com/dghubble/go-twitter/twitter"
	twitterV2 "github.com/g8rswimmer/go-twitter/v2"
	"github.com/gin-gonic/gin"
)

type (
	tweetReq struct {
		Tweet string `json:"tweet"`
	}
)

func main() {
	conn := config.GetInstance()

	r := gin.Default()
	r.POST("/tweet", func(c *gin.Context) {
		body := &tweetReq{}
		err := c.Bind(body)
		if err != nil {
			log.Printf("error :%v", err)
		}

		params := &twitter.StatusUpdateParams{Status: body.Tweet}
		twitt, resp, err := conn.TwtCliV1.Statuses.Update(body.Tweet, params)
		if err != nil {
			log.Print(err)
			c.String(500, "error bro")
			return
		}
		log.Printf("%+v\n", resp)
		log.Printf("tweet created at %v, with id https://twitter.com/anontwtdm/status/%v", time.Now().Format("15:04:05 Monday, January 2006"), twitt.ID)

		req := twitterV2.CreateTweetRequest{
			Text: body.Tweet + " V2",
		}
		fmt.Println("Callout to create tweet callout")

		tweetResponse, err := conn.TwtCliV2.CreateTweet(context.Background(), req)
		if err != nil {
			log.Panicf("create tweet error: %v", err)
		}

		enc, err := json.MarshalIndent(tweetResponse, "", "    ")
		if err != nil {
			log.Panic(err)
		}
		fmt.Println(string(enc))
		c.String(200, fmt.Sprintf("tweet created at %v, with id https://twitter.com/anontwtdm/status/%v", time.Now().Format("15:04:05 Monday, January 2006"), twitt.ID))

	})

	r.Run(":3000")

}
