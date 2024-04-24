package config

type Server struct {
	ListenAddress   string `yaml:"listen_address"`
	PathHealthcheck string `yaml:"patch_healthcheck"`
	PathWebhook     string `yaml:"patch_webhook"`
}
