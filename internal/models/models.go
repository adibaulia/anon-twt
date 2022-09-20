package models

import (
	"sync"

	"github.com/adibaulia/anon-twt/internal/models/status"
	"github.com/dghubble/go-twitter/twitter"
)

type (
	UsersConvo struct {
		sync.Mutex
		Users map[string]UserConvo
	}

	UserConvo struct {
		TwittID       string
		Username      string
		Name          string
		Status        status.Status
		TargetTwittID string
	}

	WebhookEvents struct {
		ForUserID           string                        `json:"for_user_id"`
		DirectMessageEvents *[]twitter.DirectMessageEvent `json:"direct_message_events"`
		FollowEvents        *[]FollowEvent                `json:"follow_events"`
	}

	Target struct {
		ID                             string `json:"id"`
		DefaultProfileImage            bool   `json:"default_profile_image"`
		ProfileBackgroundImageURL      string `json:"profile_background_image_url"`
		FriendsCount                   int    `json:"friends_count"`
		FavouritesCount                int    `json:"favourites_count"`
		ProfileLinkColor               int    `json:"profile_link_color"`
		ProfileBackgroundImageURLHTTPS string `json:"profile_background_image_url_https"`
		UtcOffset                      int    `json:"utc_offset"`
		ScreenName                     string `json:"screen_name"`
		IsTranslator                   bool   `json:"is_translator"`
		FollowersCount                 int    `json:"followers_count"`
		Name                           string `json:"name"`
		Lang                           string `json:"lang"`
		ProfileUseBackgroundImage      bool   `json:"profile_use_background_image"`
		CreatedAt                      string `json:"created_at"`
		ProfileTextColor               int    `json:"profile_text_color"`
		Notifications                  bool   `json:"notifications"`
		Protected                      bool   `json:"protected"`
		StatusesCount                  int    `json:"statuses_count"`
		URL                            string `json:"url"`
		ContributorsEnabled            bool   `json:"contributors_enabled"`
		DefaultProfile                 bool   `json:"default_profile"`
		ProfileSidebarBorderColor      int    `json:"profile_sidebar_border_color"`
		TimeZone                       string `json:"time_zone"`
		GeoEnabled                     bool   `json:"geo_enabled"`
		Verified                       bool   `json:"verified"`
		ProfileImageURL                string `json:"profile_image_url"`
		Following                      bool   `json:"following"`
		ProfileImageURLHTTPS           string `json:"profile_image_url_https"`
		ProfileBackgroundTile          bool   `json:"profile_background_tile"`
		ListedCount                    int    `json:"listed_count"`
		ProfileSidebarFillColor        int    `json:"profile_sidebar_fill_color"`
		Location                       string `json:"location"`
		FollowRequestSent              bool   `json:"follow_request_sent"`
		Description                    string `json:"description"`
		ProfileBackgroundColor         int    `json:"profile_background_color"`
	}
	FollowEvent struct {
		Type             string `json:"type"`
		CreatedTimestamp string `json:"created_timestamp"`
		Target           Target `json:"target"`
		Source           Source `json:"source"`
	}

	Source struct {
		ID                             string `json:"id"`
		DefaultProfileImage            bool   `json:"default_profile_image"`
		ProfileBackgroundImageURL      string `json:"profile_background_image_url"`
		FriendsCount                   int    `json:"friends_count"`
		FavouritesCount                int    `json:"favourites_count"`
		ProfileLinkColor               int    `json:"profile_link_color"`
		ProfileBackgroundImageURLHTTPS string `json:"profile_background_image_url_https"`
		UtcOffset                      int    `json:"utc_offset"`
		ScreenName                     string `json:"screen_name"`
		IsTranslator                   bool   `json:"is_translator"`
		FollowersCount                 int    `json:"followers_count"`
		Name                           string `json:"name"`
		Lang                           string `json:"lang"`
		ProfileUseBackgroundImage      bool   `json:"profile_use_background_image"`
		CreatedAt                      string `json:"created_at"`
		ProfileTextColor               int    `json:"profile_text_color"`
		Notifications                  bool   `json:"notifications"`
		Protected                      bool   `json:"protected"`
		StatusesCount                  int    `json:"statuses_count"`
		URL                            string `json:"url"`
		ContributorsEnabled            bool   `json:"contributors_enabled"`
		DefaultProfile                 bool   `json:"default_profile"`
		ProfileSidebarBorderColor      int    `json:"profile_sidebar_border_color"`
		TimeZone                       string `json:"time_zone"`
		GeoEnabled                     bool   `json:"geo_enabled"`
		Verified                       bool   `json:"verified"`
		ProfileImageURL                string `json:"profile_image_url"`
		Following                      bool   `json:"following"`
		ProfileImageURLHTTPS           string `json:"profile_image_url_https"`
		ProfileBackgroundTile          bool   `json:"profile_background_tile"`
		ListedCount                    int    `json:"listed_count"`
		ProfileSidebarFillColor        int    `json:"profile_sidebar_fill_color"`
		Location                       string `json:"location"`
		FollowRequestSent              bool   `json:"follow_request_sent"`
		Description                    string `json:"description"`
		ProfileBackgroundColor         int    `json:"profile_background_color"`
	}
)
