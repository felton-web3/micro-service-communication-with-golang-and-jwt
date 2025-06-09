package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure : "port"`
	} `mapstructure : "server"`

	JWT struct {
		Issuer                 string        `mapstructure :"issuer"`
		AccessTokenDuration    time.Duration `mapstructure : "access_token_duration_min"`
		RefreshTokenDuration   time.Duration `mapstructure :"refresh_token_duration_hours"`
		PrivateKeyFile         string
		PublicKeyFile          string
		AuthServerPublicKeyURL string
	} `mapstructure :"jwt"`
}

var AppConfig *Config

func Load() {
	AppConfig = &Config{}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		slog.Warn("config.yml not found, using defaults and environment variables")
	}
	if err := v.Unmarshal(AppConfig); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	//AppConfig.JWT.PrivateKeyFile = v.GetString("JWT_PRIVATE_KEY_FILE")
	//AppConfig.JWT.PublicKeyFile = v.GetString("JWT_PUBLIC_KEY_FILE")

	AppConfig.JWT.PrivateKeyFile = "keys/private.pem"
	AppConfig.JWT.PublicKeyFile = "keys/public.pem"
	//AppConfig.JWT.AuthServerPublicKeyURL =

	//AppConfig.JWT.AuthServerPublicKeyURL = v.GetString("JWT_AUTH_SERVICE_PUBLIC_KEY_URL") //example: http://localhost:8080/api/public-key
	AppConfig.JWT.AuthServerPublicKeyURL = "http://localhost:8080/api/public-key"

	AppConfig.JWT.AccessTokenDuration = 15 * time.Minute
	//AppConfig.JWT.AccessTokenDuration *= time.Minute

	AppConfig.JWT.RefreshTokenDuration = 24 * time.Hour

}
