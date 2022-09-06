package config

import (
	"log"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitterV2 "github.com/g8rswimmer/go-twitter/v2"
	"github.com/spf13/viper"
)

var (
	instance *Connection
)

type (
	Connection struct {
		TwtCliV1 *twitter.Client
		TwtCliV2 *twitterV2.Client
	}

	Configuration struct {
		APIKey            string `mapstructure:"APIKey"`
		APIKeySecret      string `mapstructure:"APIKeySecret"`
		BearerToken       string `mapstructure:"BearerToken"`
		AccessToken       string `mapstructure:"AccessToken"`
		AccessTokenSecret string `mapstructure:"AccessTokenSecret"`
	}
)

func init() {
	// if CONSUMER_KEY == "" || CONSUMER_KEY_SECRET == "" || ACCESS_TOKEN == "" || ACCESS_SECRET == "" {
	// 	log.Fatalf("Required Env not found")
	// }
	// oauth2 configures a client that uses app credentials to keep a fresh token
	// conf := GetConfigurationEnv()
	// config := &clientcredentials.Config{
	// 	ClientID:     conf.APIKey,
	// 	ClientSecret: conf.APIKeySecret,
	// 	TokenURL:     "https://api.twitter.com/oauth2/token",
	// }
	conf := GetConfigurationEnv()

	config := oauth1.NewConfig(conf.APIKey, conf.APIKeySecret)
	httpClient := config.Client(oauth1.NoContext, &oauth1.Token{
		Token:       conf.AccessToken,
		TokenSecret: conf.AccessTokenSecret,
	})

	// Twitter client
	client := twitter.NewClient(httpClient)

	clientV2 := &twitterV2.Client{
		Authorizer: &authorize{},
		Client:     httpClient,
		Host:       "https://api.twitter.com",
	}

	instance = &Connection{
		TwtCliV1: client,
		TwtCliV2: clientV2,
	}

}
func GetConfigurationEnv() *Configuration {

	viper.AddConfigPath(".")
	viper.SetConfigName("keys")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Cannot load configuration file: %v", err)
	}

	conf := &Configuration{}
	err = viper.Unmarshal(conf)
	if err != nil {
		log.Fatalf("Cannot unmarshall config: %v", err)
	}
	return conf
}

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {}

func GetInstance() *Connection {
	return instance
}
