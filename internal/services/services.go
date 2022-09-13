package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	log "github.com/sirupsen/logrus"
)

type (
	svc struct {
		TwtCli
	}

	Serve interface {
		SendWelcomeMessage(models.FollowEvent) error
		DirectMessagesEventProcessor(twitter.DirectMessageEvent) error
	}

	TwtCli interface {
		EventsNew(params *twitter.DirectMessageEventsNewParams) (*twitter.DirectMessageEvent, *http.Response, error)
	}
)

var Convos *models.UsersConvo

func NewService(twtCli TwtCli) *svc {
	Convos = &models.UsersConvo{
		Users: map[string]models.UserConvo{},
	}
	return &svc{twtCli}
}
func (s *svc) DirectMessagesEventProcessor(event twitter.DirectMessageEvent) error {
	senderID := event.Message.SenderID

	if senderID == SelfTwitterID {
		return nil
	}
	switch event.Message.Data.Text {
	case "/start":
		err, done := s.startPairing(senderID)
		if done {
			return err
		}
	case "/stop":
		s.stopProcess(senderID)
	default:
		s.routingDirectMessage(event, senderID)
	}

	return nil
}

func (s *svc) startPairing(senderID string) (error, bool) {
	Convos.Lock()
	users := Convos.Users
	curUser, found := users[senderID]
	if found {
		tarUser, foundTar := users[curUser.TargetTwittID]
		if foundTar {
			if curUser.TargetTwittID == tarUser.TwittID && tarUser.TargetTwittID == curUser.TwittID && tarUser.Status == models.InConvo {
				Convos.Unlock()
				return nil, true
			}
		}
	}
	Convos.Unlock()

	s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		Text: "[🤖] Searching stranger...",
	})
	done := make(chan bool)
	go s.pairingProcess(senderID, done)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	select {
	case <-done:
		Convos.Lock()
		users := Convos.Users
		curUser := users[senderID]
		Convos.Unlock()
		log.Print("pairing success user: %v", curUser.TwittID)
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "[🤖] Settled and ready to chat!",
			QuickReply: NewQRBuilder().CustomQuickRepy(twitter.DirectMessageQuickReplyOption{
				Label:       "Hi!",
				Description: "Say hello for stranger!",
				Metadata:    "external_id_2",
			}).StopButton().Build(),
		})
	case <-ctx.Done():
		Convos.Lock()
		defer Convos.Unlock()
		users := Convos.Users
		curUser := users[senderID]
		curUser.Status = models.End
		curUser.TargetTwittID = ""
		curUser.Timeout = true

		users[senderID] = curUser
		Convos.Users = users
		log.Print("timeout user: %v", curUser.TwittID)
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text:       "[🤖] Can't find stranger but you can start search again!",
			QuickReply: NewQRBuilder().StartButton().Build(),
		})
	}

	return nil, false
}

func (s *svc) pairingProcess(senderID string, done chan bool) {
	for {
		Convos.Lock()
		users := Convos.Users
		curUser := users[senderID]
		if !curUser.Timeout {
			curUser.Status = models.Ready
		}
		targetTwittID := ""
		for targetID, user := range users {
			if user.TwittID != senderID && (user.Status == models.Ready || (user.Status == models.InConvo && user.TargetTwittID == senderID)) && curUser.Status == models.Ready {
				curUser.Status = models.InConvo
				curUser.TargetTwittID = user.TwittID
				targetTwittID = targetID
			}
		}
		users[senderID] = curUser
		Convos.Users = users

		if curUser.TargetTwittID == targetTwittID && users[targetTwittID].TargetTwittID == senderID {
			Convos.Unlock()
			done <- true
			break
		}
		Convos.Unlock()
	}
}

func (s *svc) routingDirectMessage(event twitter.DirectMessageEvent, senderID string) {
	Convos.Lock()
	defer Convos.Unlock()
	text := event.Message.Data.Text
	users := Convos.Users
	curUser := users[senderID]
	if senderID == SelfTwitterID {
		return
	}
	if curUser.Status == models.InConvo {
		s.sendDirectMessage(curUser.TargetTwittID, &twitter.DirectMessageData{
			Text:       text,
			QuickReply: NewQRBuilder().StopButton().Build(),
		})
	} else {
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text:       "[🤖] You can use this bot by typing /start to start convo with stranger",
			QuickReply: NewQRBuilder().StartButton().Build(),
		})
	}
}

func (s *svc) stopProcess(senderID string) {
	Convos.Lock()
	defer Convos.Unlock()
	users := Convos.Users
	curUser := users[senderID]
	tarUser := users[curUser.TargetTwittID]

	curUser.Status = models.End
	tarUser.Status = models.End

	curUser.TargetTwittID = ""
	tarUser.TargetTwittID = ""

	users[senderID] = curUser
	users[tarUser.TwittID] = tarUser

	Convos.Users = users

	s.sendDirectMessage(tarUser.TwittID, &twitter.DirectMessageData{
		Text:       "[🤖] Your Partner stopped the convo",
		QuickReply: NewQRBuilder().StartButton().Build(),
	})
	s.sendDirectMessage(curUser.TwittID, &twitter.DirectMessageData{
		Text:       "[🤖] Convo stopped",
		QuickReply: NewQRBuilder().StartButton().Build(),
	})
}

func (s *svc) SendWelcomeMessage(event models.FollowEvent) error {
	if event.Type == Follow {
		user := event.Target
		log.Infof("Welcome Message triggered")

		users := Convos.Users
		if _, found := users[user.ID]; !found {
			func() {
				Convos.Lock()
				defer Convos.Unlock()
				users[user.ID] = models.UserConvo{
					TwittID:  user.ID,
					Name:     user.Name,
					Username: user.ScreenName,
					Status:   models.End,
				}
				Convos.Users = users
			}()
		}
		s.sendDirectMessage(user.ID, &twitter.DirectMessageData{
			Text:       welcomeMessage(user.Name),
			QuickReply: NewQRBuilder().StartButton().Build(),
		})
	}

	return nil
}

func (s *svc) sendDirectMessage(targetID string, directMessageData *twitter.DirectMessageData) {
	s.EventsNew(&twitter.DirectMessageEventsNewParams{
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
