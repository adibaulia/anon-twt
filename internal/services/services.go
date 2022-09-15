package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/dghubble/go-twitter/twitter"
	twitterV2 "github.com/g8rswimmer/go-twitter/v2"
	log "github.com/sirupsen/logrus"
)

type (
	Svc struct {
		TwtCliV1
		TwtCliV2
	}

	Serve interface {
		SendWelcomeMessage(models.FollowEvent) error
		DirectMessagesEventProcessor(twitter.DirectMessageEvent) error
	}

	TwtCliV1 interface {
		EventsNew(params *twitter.DirectMessageEventsNewParams) (*twitter.DirectMessageEvent, *http.Response, error)
	}
	TwtCliV2 interface {
		UserLookup(ctx context.Context, ids []string, opts twitterV2.UserLookupOpts) (*twitterV2.UserLookupResponse, error)
	}
)

var Convos *models.UsersConvo

func NewService(twtCli TwtCliV1, v2 TwtCliV2) *Svc {
	Convos = &models.UsersConvo{
		Users: map[string]models.UserConvo{},
	}
	return &Svc{twtCli, v2}
}

func (s *Svc) SendWelcomeMessage(event models.FollowEvent) error {
	if event.Type == Follow {
		user := event.Target
		log.Infof("Welcome Message triggered")
		s.sendDirectMessage(user.ID, &twitter.DirectMessageData{
			Text:       welcomeMessage(user.Name),
			QuickReply: NewQRBuilder().RegisterButton().Build(),
		})
	}

	return nil
}

func (s *Svc) DirectMessagesEventProcessor(event twitter.DirectMessageEvent) error {
	senderID := event.Message.SenderID

	if senderID == SelfTwitterID {
		return nil
	}
	incomingMessage := event.Message.Data.Text
	switch incomingMessage {
	case "/register":
		s.createNewUser(senderID)
	case "/start":
		err := s.startPairing(senderID)
		if err != nil {
			s.sendDirectMessage(senderID, &twitter.DirectMessageData{
				Text:       "[🤖] Make sure you have register first!",
				QuickReply: NewQRBuilder().RegisterButton().Build(),
			})
			return nil
		}
	case "/stop":
		s.stopProcess(senderID)
	default:
		//commands := []string{"/register", "start", "/stop"}
		//for _, command := range commands {
		//	if incomingMessage == command {
		//		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		//			Text: "[🤖] Searching stranger...",
		//		})
		//		return nil
		//	}
		//}
		s.routingDirectMessage(event, senderID)
	}

	return nil
}

func (s *Svc) startPairing(senderID string) error {
	users := Convos.Users
	curUser, found := users[senderID]
	if found {
		Convos.Lock()
		tarUser, foundTar := users[curUser.TargetTwittID]
		if foundTar {
			if s.isInvalidUser(curUser, tarUser) {
				Convos.Unlock()
				return fmt.Errorf("invalid user")
			}
		}
		Convos.Unlock()
	} else {
		return fmt.Errorf("invalid user")
	}

	s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		Text: "[🤖] Searching stranger...",
	})

	successPair := make(chan bool)
	go s.pairingProcess(senderID, successPair)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	select {
	case <-successPair:
		s.successPairing(senderID)
	case <-ctx.Done():
		s.timeoutPairing(senderID)
	}

	return nil
}

func (s *Svc) createNewUser(senderID string) {
	Convos.Lock()
	defer Convos.Unlock()
	users := Convos.Users
	_, found := users[senderID]
	if found {
		s.sendDirectMessage(senderID, &twitter.DirectMessageData{
			Text: "[🤖] You Already registered!",
		})
		return
	}

	s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		Text: "[🤖] Registering some value from you",
	})
	user := users[senderID]
	userResp, err := s.UserLookup(context.Background(), []string{senderID}, twitterV2.UserLookupOpts{})
	if err != nil {
		log.Printf("user lookup error: %v", err)
	}
	dict := userResp.Raw.UserDictionaries()
	u, found := dict[senderID]
	if found {
		user = models.UserConvo{
			Name:     u.User.Name,
			Username: u.User.UserName,
		}
	}
	user.TwittID = senderID
	user.Status = models.End
	users[senderID] = user
	Convos.Users = users

	s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		Text:       "[🤖] Registered! Now you can start searching convo!",
		QuickReply: NewQRBuilder().StartButton().Build(),
	})
}

func (s *Svc) timeoutPairing(senderID string) {
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

func (s *Svc) successPairing(senderID string) {
	Convos.Lock()
	defer Convos.Unlock()
	users := Convos.Users
	curUser := users[senderID]

	log.Print("pairing success user: %v", curUser.TwittID)

	s.sendDirectMessage(senderID, &twitter.DirectMessageData{
		Text: "[🤖] Settled and ready to chat!",
		QuickReply: NewQRBuilder().CustomQuickRepy(twitter.DirectMessageQuickReplyOption{
			Label:       "Hi!",
			Description: "Say hello for stranger!",
			Metadata:    "external_id_2",
		}).StopButton().Build(),
	})
}

func (s *Svc) isInvalidUser(curUser models.UserConvo, tarUser models.UserConvo) bool {
	return curUser.TargetTwittID == tarUser.TwittID && tarUser.TargetTwittID == curUser.TwittID && tarUser.Status == models.InConvo
}

func (s *Svc) pairingProcess(senderID string, done chan bool) {
	for {
		Convos.Lock()
		users := Convos.Users
		curUser := users[senderID]
		if !curUser.Timeout {
			curUser.Status = models.Ready
		}
		targetTwittID := ""
		for targetID, user := range users {
			if s.isValidUserReadyToPair(senderID, user, curUser) {
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

func (s *Svc) isValidUserReadyToPair(senderID string, user models.UserConvo, curUser models.UserConvo) bool {
	return user.TwittID != senderID && (user.Status == models.Ready || (user.Status == models.InConvo && user.TargetTwittID == senderID)) && curUser.Status == models.Ready
}

func (s *Svc) routingDirectMessage(event twitter.DirectMessageEvent, senderID string) {
	if senderID == SelfTwitterID {
		return
	}
	Convos.Lock()
	defer Convos.Unlock()
	text := event.Message.Data.Text
	users := Convos.Users
	curUser := users[senderID]
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

func (s *Svc) stopProcess(senderID string) {
	Convos.Lock()
	defer Convos.Unlock()
	users := Convos.Users
	curUser := users[senderID]
	tarUser := users[curUser.TargetTwittID]

	if curUser.Status != models.InConvo && tarUser.Status != models.InConvo {
		s.sendDirectMessage(curUser.TwittID, &twitter.DirectMessageData{
			Text:       "[🤖] Invalid command! You must register or start first for stopping convo",
			QuickReply: NewQRBuilder().RegisterButton().StartButton().Build(),
		})
		return
	}

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

func (s *Svc) sendDirectMessage(targetID string, directMessageData *twitter.DirectMessageData) {
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
		Type /register to register you and /start convo!`, name)
}
