package main

import (
	"fmt"
	"log"
	"time"

	"github.com/adibaulia/anon-twt/config"
	"github.com/dghubble/go-twitter/twitter"
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

		//	log.Printf("params '%+v'", params)
		twitt, _, err := conn.TwtCliV1.Statuses.Update(body.Tweet, params)
		if err != nil {
			log.Print(err)
			c.String(500, "error bro")
			return
		}
		log.Printf("tweet created at %v, with id https://twitter.com/anontwtdm/status/%v", time.Now().Format("15:04:05 Monday, January 2006"), twitt.ID)
		c.String(500, fmt.Sprintf("tweet created at %v, with id https://twitter.com/anontwtdm/status/%v", time.Now().Format("15:04:05 Monday, January 2006"), twitt.ID))
	})

	r.Run(":3000")

}
