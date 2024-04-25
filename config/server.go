package config

type Server struct {
	ListenAddress   string `yaml:"listenAddress,omitempty"`
	PathHealthcheck string `yaml:"patchHealthcheck,omitempty"`
	PathWebhook     string `yaml:"patchWebhook,omitempty"`
}
