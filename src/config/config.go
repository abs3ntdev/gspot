package config

type Config struct {
	ClientId        string `yaml:"client_id"`
	ClientSecret    string `yaml:"client_secret"`
	ClientSecretCmd string `yaml:"client_secret_cmd"`
	Port            string `yaml:"port"`
	LogLevel        string `yaml:"log_level"         default:"info"`
	LogOutput       string `yaml:"log_output"        default:"stdout"`
}
