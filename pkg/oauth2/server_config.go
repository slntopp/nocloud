package oauth2

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
)

var ServerConfigLocation string

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("OAUTH2_SERVER_CONFIG_LOCATION", "oauth2_server_config.json")

	ServerConfigLocation = viper.GetString("OAUTH2_SERVER_CONFIG_LOCATION")
}

func ServerConfig() (Config, error) {
	var config Config
	conf, err := os.ReadFile(ServerConfigLocation)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(conf, &config)
	return config, err
}
