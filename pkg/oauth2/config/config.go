package config

import (
	"encoding/json"
	"github.com/spf13/viper"
	"os"
)

var CONFIG_LOCATION string

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("CONFIG_LOCATION", "oauth2_config.json")

	CONFIG_LOCATION = viper.GetString("CONFIG_LOCATION")
}

type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
	AuthField    string   `json:"auth_field"`
	Scopes       []string `json:"scopes"`
}

func Config() (map[string]OAuth2Config, error) {
	var config map[string]OAuth2Config

	conf, err := os.ReadFile(CONFIG_LOCATION)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(conf, &config)
	return config, err
}
