package config

type Config struct {
	ListenBrainzEndpoint     string `yaml:"listenbrainz_endpoint" default:"https://api.listenbrainz.org"`
	ListenBrainzLabsEndpoint string `yaml:"listenbrainz_labs_endpoint" default:"https://labs.api.listenbrainz.org"`
	ListenBrainzUserToken    string `yaml:"listenbrainz_user_token"`
	ClientID                 string `yaml:"client_id"`
	ClientSecret             string `yaml:"client_secret"`
	ClientSecretCmd          string `yaml:"client_secret_cmd"`
	Port                     string `yaml:"port"`
	LogLevel                 string `yaml:"log_level"         default:"info"`
	LogOutput                string `yaml:"log_output"        default:"stdout"`
}
