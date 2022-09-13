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
		convos.Lock()
		users := convos.Users
		curUser, found := users[senderID]
		if found {
			tarUser, foundTar := users[curUser.TargetTwittID]
			if foundTar {
				if curUser.TargetTwittID == tarUser.TwittID && tarUser.TargetTwittID == curUser.TwittID && tarUser.Status == models.InConvo {
					convos.Unlock()
					return nil
				}
			}
		}
		convos.Unlock()

		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "[🤖] Searching stranger...",
		})
		for {
			convos.Lock()
			users := convos.Users
			curUser := users[senderID]
			//if curUser.Status != models.InConvo {
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
			//}
			convos.Unlock()
		}
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "[🤖] Settled and ready to chat!",
			QuickReply: NewQRBuilder().CustomQuickRepy(twitter.DirectMessageQuickReplyOption{
				Label:       "Hi!",
				Description: "Say hello for stranger!",
				Metadata:    "external_id_2",
			}).StopButton().GetQuickReply(),
		})

	case "/stop":
		s.stopProcess(senderID)
	default:
		s.routingDirectMessage(event, senderID)
	}

	return nil
}

func (s *svc) routingDirectMessage(event twitter.DirectMessageEvent, senderID string) {
	convos.Lock()
	defer convos.Unlock()
	text := event.Message.Data.Text
	users := convos.Users
	curUser := users[senderID]
	if senderID == SelfTwitterID {
		return
	}
	if curUser.Status == models.InConvo {
		s.sendDirectMessage(curUser.TargetTwittID, &twitter.DirectMessageData{
			Text:       text,
			QuickReply: NewQRBuilder().StopButton().GetQuickReply(),
		})
	} else {
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text:       "[🤖] You can use this bot by typing /start to start convo with stranger",
			QuickReply: NewQRBuilder().StartButton().GetQuickReply(),
		})
	}
}

func (s *svc) stopProcess(senderID string) {
	convos.Lock()
	defer convos.Unlock()
	users := convos.Users
	curUser := users[senderID]
	tarUser := users[curUser.TargetTwittID]

	curUser.Status = models.End
	tarUser.Status = models.End

	curUser.TargetTwittID = ""
	tarUser.TargetTwittID = ""

	users[senderID] = curUser
	users[tarUser.TwittID] = tarUser

	convos.Users = users

	s.sendDirectMessage(tarUser.TwittID, &twitter.DirectMessageData{
		Text:       "[🤖] Your Partner stopped the convo",
		QuickReply: NewQRBuilder().StartButton().GetQuickReply(),
	})
	s.sendDirectMessage(curUser.TwittID, &twitter.DirectMessageData{
		Text:       "[🤖] Convo stopped",
		QuickReply: NewQRBuilder().StartButton().GetQuickReply(),
	})
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
			Text:       welcomeMessage(user.Name),
			QuickReply: NewQRBuilder().StartButton().GetQuickReply(),
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

func welcomeMessage(name string) string {
	return fmt.Sprintf(
		`[🤖][🤖][🤖]
		Hello %v!
		You have been followed!
		Now you can use our app to have one on one convo with other stranger!
		Type /start to start convo with others!`, name)
}
