package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/adibaulia/anon-twt/config"
	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/adibaulia/anon-twt/internal/services"
	"github.com/gin-gonic/gin"
)

var svcs services.Serve

func Router(r *gin.Engine, svc services.Serve) {
	svcs = svc
	r.POST("/twitter/webhook", webhookEvent)
	r.GET("/twitter/webhook", CRC)
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "masuk bro")
		return
	})
}

func webhookEvent(c *gin.Context) {

	events := &models.WebhookEvents{}
	err := c.Bind(events)
	if err != nil {
		log.Errorf("ERROR", err)
		return
	}

	bodyByte, err := json.Marshal(events)
	if err != nil {
		log.Errorf("ERROR", err)
	}

	log.Printf("Received Body from Webhooks: %v", string(bodyByte))
	webhookRouter(events)
	c.Status(http.StatusOK)
	return
}

func webhookRouter(events *models.WebhookEvents) {
	switch {
	case events.FollowEvents != nil:
		for _, event := range *events.FollowEvents {
			go svcs.SendWelcomeMessage(event)
		}
	case events.DirectMessageEvents != nil:
		for _, event := range *events.DirectMessageEvents {
			go svcs.DirectMessagesEventProcessor(event)
		}
	}
}

//CRC Check from securing webhooks twitter api
func CRC(c *gin.Context) {
	log.Info("CRC check from twitter API")

	secret := []byte(config.GetConf().APIKeySecret)
	message := []byte(c.Query("crc_token"))

	hash := hmac.New(sha256.New, secret)
	hash.Write(message)

	// to base64
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	resp := map[string]string{
		"response_token": "sha256=" + token,
	}
	log.Info("CRC token successfully checked")
	c.JSON(http.StatusOK, resp)
}
