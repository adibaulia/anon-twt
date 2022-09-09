package services

import (
	"context"
	"fmt"

	"github.com/adibaulia/anon-twt/config"
	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	twitterV2 "github.com/g8rswimmer/go-twitter/v2"
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

	opts := twitterV2.UserLookupOpts{
		Expansions: []twitterV2.Expansion{twitterV2.ExpansionPinnedTweetID},
	}

	userResponse, err := s.client.TwtCliV2.UserLookup(context.Background(), []string{senderID}, opts)
	if err != nil {
		log.Panicf("user lookup error: %v", err)
	}

	dictionaries := userResponse.Raw.UserDictionaries()

	userData := dictionaries[senderID]

	if senderID == SelfTwitterID {
		return nil
	}
	switch event.Message.Data.Text {
	case "/start":
		func() {
			convos.Lock()
			defer convos.Unlock()
			for {
				users := convos.Users
				if _, found := users[senderID]; !found {
					func() {
						users[senderID] = models.UserConvo{
							TwittID:  senderID,
							Name:     userData.User.Name,
							Username: userData.User.UserName,
							Status:   models.Ready,
						}
						convos.Users = users
					}()
					continue
				} else {
					found := false
					for id, targetUser := range users {
						if targetUser.Status == models.Ready && id != senderID {
							func() {
								user := users[senderID]
								user.TargetTwittID = targetUser.TwittID
								user.Status = models.InConvo
								users[senderID] = user
								convos.Users = users
							}()
							found = true
						}
					}
					if !found {
						continue
					} else {
						break
					}
				}
			}
		}()
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "ready to go",
		})
	}
	return nil
}

func (s *svc) SendWelcomeMessage(event models.FollowEvent) error {
	if event.Type == Follow {
		follower := event.Target
		log.Infof("Welcome Message triggered")
		s.sendDirectMessage(follower.ID, &twitter.DirectMessageData{
			Text:       wellcomeMessage(follower.Name),
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
