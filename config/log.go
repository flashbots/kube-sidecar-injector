package config

type Log struct {
	Level string `yaml:"level,omitempty"`
	Mode  string `yaml:"mode,omitempty"`
}
