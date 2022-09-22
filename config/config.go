package config

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
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
		Port     string
	}

	Configuration struct {
		APIKey            string `mapstructure:"API_KEY"`
		APIKeySecret      string `mapstructure:"API_KEY_SECRET"`
		BearerToken       string `mapstructure:"BEARER_TOKEN"`
		AccessToken       string `mapstructure:"ACCESS_TOKEN"`
		AccessTokenSecret string `mapstructure:"ACCESS_TOKEN_SECRET"`
		LogPath           string `mapstructure:"LOG_PATH"`
		LogLevel          string `mapstructure:"LOG_LEVEL"`
		Port              string `mapstructure:"PORT" env:"PORT"`
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
		Port:     conf.Port,
	}

}
func getConf() *Configuration {

	conf := &Configuration{}
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	envKeysMap := &map[string]interface{}{}
	if err := mapstructure.Decode(conf, &envKeysMap); err != nil {
		log.Fatal(err)
	}
	for k := range *envKeysMap {
		if bindErr := viper.BindEnv(k); bindErr != nil {
			log.Fatal(bindErr)
		}
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Cannot load configuration file: %v", err)
		log.Errorf("Reading from given env: %v", err)
	}

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
