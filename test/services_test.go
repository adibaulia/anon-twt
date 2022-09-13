package test

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/adibaulia/anon-twt/internal/models"
	"github.com/adibaulia/anon-twt/internal/services"
	"github.com/dghubble/go-twitter/twitter"
)

type (
	MockTwtCli struct{}
)

var (
	directMessageEvent = `{"for_user_id":"1566612593162088448","direct_message_events":[{"created_timestamp":"1663039900427","id":"1569529216978550790","type":"message_create","message_create":{"sender_id":"1453000592033472525","target":{"recipient_id":"1566612593162088448"},"message_data":{"text":"/start","entities":{"hashtags":[],"media":null,"urls":[],"user_mentions":[],"symbols":[],"polls":null}}}}],"follow_events":null}`
	followEvent        = `{"for_user_id":"1566612593162088448","follow_events":[{"type":"follow","created_timestamp":"1663039656087","target":{"id":"oke","default_profile_image":false,"profile_background_image_url":"","friends_count":331,"favourites_count":2588,"profile_link_color":-1,"profile_background_image_url_https":"","utc_offset":0,"screen_name":"syncting","is_translator":false,"followers_count":205,"name":"Loid Forger","lang":"","profile_use_background_image":false,"created_at":"Tue Oct 26 14:08:38 +0000 2021","profile_text_color":-1,"notifications":false,"protected":false,"statuses_count":8035,"url":"","contributors_enabled":false,"default_profile":true,"profile_sidebar_border_color":-1,"time_zone":"","geo_enabled":false,"verified":false,"profile_image_url":"http://pbs.twimg.com/profile_images/1561015056069902336/lYej0rCo_normal.jpg","following":false,"profile_image_url_https":"https://pbs.twimg.com/profile_images/1561015056069902336/lYej0rCo_normal.jpg","profile_background_tile":false,"listed_count":0,"profile_sidebar_fill_color":-1,"location":"","follow_request_sent":false,"description":"INFJ | software engineer","profile_background_color":-1}},{"type":"follow","created_timestamp":"1663039656087","target":{"id":"mantap","default_profile_image":false,"profile_background_image_url":"","friends_count":331,"favourites_count":2588,"profile_link_color":-1,"profile_background_image_url_https":"","utc_offset":0,"screen_name":"syncting","is_translator":false,"followers_count":205,"name":"Loid Forger","lang":"","profile_use_background_image":false,"created_at":"Tue Oct 26 14:08:38 +0000 2021","profile_text_color":-1,"notifications":false,"protected":false,"statuses_count":8035,"url":"","contributors_enabled":false,"default_profile":true,"profile_sidebar_border_color":-1,"time_zone":"","geo_enabled":false,"verified":false,"profile_image_url":"http://pbs.twimg.com/profile_images/1561015056069902336/lYej0rCo_normal.jpg","following":false,"profile_image_url_https":"https://pbs.twimg.com/profile_images/1561015056069902336/lYej0rCo_normal.jpg","profile_background_tile":false,"listed_count":0,"profile_sidebar_fill_color":-1,"location":"","follow_request_sent":false,"description":"INFJ | software engineer","profile_background_color":-1}}]}`
)

func (m *MockTwtCli) EventsNew(params *twitter.DirectMessageEventsNewParams) (*twitter.DirectMessageEvent, *http.Response, error) {
	log.Printf("Sended DM to twittID: %v, with value: %v", params.Event.Message.Target, params.Event.Message.Data)
	return nil, nil, nil
}

func TestAll(t *testing.T) {

	followEvents := []models.FollowEvent{
		{
			Type: "follow",
			Target: models.Target{
				ID: "oke",
			},
		},
		{
			Type: "follow",
			Target: models.Target{
				ID: "mantap",
			},
		},
		{
			Type: "follow",
			Target: models.Target{
				ID: "banget",
			},
		},
		{
			Type: "follow",
			Target: models.Target{
				ID: "sip",
			},
		},
	}

	svc := services.NewService(&MockTwtCli{})

	for _, v := range followEvents {
		svc.SendWelcomeMessage(v)
	}

	time.Sleep(1 * time.Second)

	DMEvents := []twitter.DirectMessageEvent{
		{
			Message: &twitter.DirectMessageEventMessage{
				SenderID: "mantap",
				Data: &twitter.DirectMessageData{
					Text: "/start",
				},
			},
		},
		{
			Message: &twitter.DirectMessageEventMessage{
				SenderID: "oke",
				Data: &twitter.DirectMessageData{
					Text: "/start",
				},
			},
		},
		{
			Message: &twitter.DirectMessageEventMessage{
				SenderID: "banget",
				Data: &twitter.DirectMessageData{
					Text: "/start",
				},
			},
		},
		{
			Message: &twitter.DirectMessageEventMessage{
				SenderID: "sip",
				Data: &twitter.DirectMessageData{
					Text: "/start",
				},
			},
		},
	}

	wg := &sync.WaitGroup{}

	for _, v := range DMEvents {
		wg.Add(1)
		v := v
		go func() {
			svc.DirectMessagesEventProcessor(v)
			wg.Done()
		}()

	}

	wg.Wait()
	oke := services.Convos
	fmt.Printf("%v", oke)

}
