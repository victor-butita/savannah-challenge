package config

import "github.com/spf13/viper"

type Config struct {
	ServerPort      string `mapstructure:"SERVER_PORT"`
	DatabaseURL     string `mapstructure:"DATABASE_URL"`
	OIDCProviderURL string `mapstructure:"OIDC_PROVIDER_URL"`
	OIDCClientID    string `mapstructure:"OIDC_CLIENT_ID"`
	ATUsername      string `mapstructure:"AT_USERNAME"`
	ATAPIKey        string `mapstructure:"AT_API_KEY"`
	ATEnv           string `mapstructure:"AT_ENV"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
