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
		DirectMessagesEventProcessor(twitter.DirectMessageEvent) error
	}
)

var convos *models.UsersConvo

func NewService(client *config.Connection) *svc {
	convos = &models.UsersConvo{
		Users: map[string]models.UserConvo{},
	}
	return &svc{client}
}
func (s *svc) DirectMessagesEventProcessor(event twitter.DirectMessageEvent) error {
	senderID := event.Message.SenderID

	if senderID == SelfTwitterID {
		return nil
	}
	switch event.Message.Data.Text {
	case "/start":
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "Searching...",
		})
		for {
			convos.Lock()
			users := convos.Users
			curUser := users[senderID]
			curUser.Status = models.Ready
			targetTwittID := ""
			for targetID, user := range users {
				if user.TwittID != senderID && (user.Status == models.Ready || (user.Status == models.InConvo && user.TargetTwittID == senderID)) && curUser.Status == models.Ready {
					curUser.Status = models.InConvo
					curUser.TargetTwittID = user.TwittID
					targetTwittID = targetID
				}
			}
			users[senderID] = curUser
			convos.Users = users

			if curUser.TargetTwittID == targetTwittID && users[targetTwittID].TargetTwittID == senderID {
				convos.Unlock()
				break
			}
			convos.Unlock()
		}
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "Settled and ready to chat!",
		})

	default:
		text := event.Message.Data.Text
		convos.Lock()
		users := convos.Users
		curUser := users[senderID]
		if curUser.Status == models.InConvo {
			s.sendDirectMessage(curUser.TargetTwittID, &twitter.DirectMessageData{
				Text: text,
			})
		}
		convos.Unlock()
	}

	return nil
}

func (s *svc) SendWelcomeMessage(event models.FollowEvent) error {
	if event.Type == Follow {
		user := event.Target
		log.Infof("Welcome Message triggered")

		users := convos.Users
		if _, found := users[user.ID]; !found {
			func() {
				convos.Lock()
				defer convos.Unlock()
				users[user.ID] = models.UserConvo{
					TwittID:  user.ID,
					Name:     user.Name,
					Username: user.ScreenName,
					Status:   models.End,
				}
				convos.Users = users
			}()
		}
		s.sendDirectMessage(user.ID, &twitter.DirectMessageData{
			Text:       wellcomeMessage(user.Name),
			QuickReply: defaultQuickReply(),
		})
	}

	return nil
}

func (s *svc) sendDirectMessage(targetID string, directMessageData *twitter.DirectMessageData) {
	s.client.TwtCliV1.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
		Event: &twitter.DirectMessageEvent{
			Type: "message_create",
			Message: &twitter.DirectMessageEventMessage{
				Target: &twitter.DirectMessageTarget{
					RecipientID: targetID,
				},
				Data: directMessageData,
			},
		},
	})
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
