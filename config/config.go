package config

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	twitterV2 "github.com/g8rswimmer/go-twitter/v2"
	"github.com/spf13/viper"
)

var (
	instance *Connection
	conf     *Configuration
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
		logLevel          string `mapstructure:"logLevel"`
		logPath           string `mapstructure:"logPath"`
	}
)

func init() {

	conf = getConf()
	initLog()
	config := oauth1.NewConfig(conf.APIKey, conf.APIKeySecret)
	token := oauth1.NewToken(conf.AccessToken, conf.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	//oauth2 configures a client that uses app credentials to keep a fresh token
	// config := &clientcredentials.Config{
	// 	ClientID:     conf.APIKey,
	// 	ClientSecret: conf.APIKeySecret,
	// 	TokenURL:     "https://api.twitter.com/oauth2/token",
	// }
	// http.Client will automatically authorize Requests
	// httpClient := config.Client(context.Background())

	// Twitter clientApiV1
	clientApiV1 := twitter.NewClient(httpClient)

	clientApiV2 := &twitterV2.Client{
		Authorizer: &authorize{},
		Client:     httpClient,
		Host:       "https://api.twitter.com",
	}
	instance = &Connection{
		TwtCliV1: clientApiV1,
		TwtCliV2: clientApiV2,
	}

}
func getConf() *Configuration {

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

func GetConf() *Configuration {
	return conf
}
