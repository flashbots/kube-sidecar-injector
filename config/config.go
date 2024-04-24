package config

type Config struct {
	K8S    K8S    `yaml:"k8s"`
	Log    Log    `yaml:"log"`
	Server Server `yaml:"server"`

	Version string
}
