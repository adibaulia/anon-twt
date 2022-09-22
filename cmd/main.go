package main

import (
	"fmt"

	"github.com/adibaulia/anon-twt/api"
	"github.com/adibaulia/anon-twt/config"
	"github.com/adibaulia/anon-twt/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	conn := config.GetInstance()

	r := gin.New()
	r.Use(config.Logger())
	svc := services.NewService(conn.TwtCliV1.DirectMessages, conn.TwtCliV2)
	api.Router(r, svc)

	r.Run(fmt.Sprintf(":%v", conn.Port))
}
