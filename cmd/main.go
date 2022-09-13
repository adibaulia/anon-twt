package main

import (
	"github.com/adibaulia/anon-twt/api"
	"github.com/adibaulia/anon-twt/config"
	"github.com/adibaulia/anon-twt/internal/services"
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
	r.Use(config.Logger())
	svc := services.NewService(conn.TwtCliV1.DirectMessages, conn.TwtCliV2)
	api.Router(r, svc)

	r.Run(":3000")

}
