package services

import (
	"fmt"

	"github.com/adibaulia/anon-twt/config"
	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	log "github.com/sirupsen/logrus"
)

type (
	svc struct {
		client *config.Connection
	}

	Serve interface {
		SendWelcomeMessage(models.FollowEvent) error
	}
)

func NewService(client *config.Connection) *svc {
	return &svc{client}
}

func (s *svc) SendWelcomeMessage(event models.FollowEvent) error {
	if event.Type == FOLLOW {
		follower := event.Target
		log.Infof("Welcome Message triggered")
		s.client.TwtCliV1.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
			Event: &twitter.DirectMessageEvent{
				Type: "message_create",
				Message: &twitter.DirectMessageEventMessage{
					Target: &twitter.DirectMessageTarget{
						RecipientID: follower.ID,
					},
					Data: &twitter.DirectMessageData{
						Text:       wellcomeMessage(follower.Name),
						QuickReply: defaultQuickReply(),
					},
				},
			},
		})

	}

	return nil
}
func defaultQuickReply() *twitter.DirectMessageQuickReply {
	return &twitter.DirectMessageQuickReply{
		Type: "options",
		Options: []twitter.DirectMessageQuickReplyOption{{
			Label:       "/start",
			Description: "Start convo 🚀🚀🚀",
			Metadata:    "external_id_1",
		}},
	}
}

func wellcomeMessage(name string) string {
	return fmt.Sprintf(
		`Hello %v!
		You have been followed!
		Now you can use our app to have one on one convo with other stranger!
		Type /start to start convo with others!`, name)
}
