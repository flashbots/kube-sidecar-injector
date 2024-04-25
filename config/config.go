package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Inject []Inject `yaml:"inject,omitempty"`
	K8S    K8S      `yaml:"k8s"`
	Log    Log      `yaml:"log"`
	Server Server   `yaml:"server"`

	Version string
}

var (
	ErrConfigFailedToReadFromFile  = errors.New("failed to read configuration from file")
	ErrConfigurationFailedToDecode = errors.New("failed to decode configuration")
)

func ReadFrom(file string) (
	*Config, error,
) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w",
			ErrConfigFailedToReadFromFile, file, err,
		)
	}
	d := yaml.NewDecoder(f)
	d.KnownFields(true)
	var _cfg Config
	if err := d.Decode(&_cfg); err != nil {
		return nil, fmt.Errorf("%w: %s: %w",
			ErrConfigurationFailedToDecode, file, err,
		)
	}
	return &Config{
		Inject: _cfg.Inject,
	}, nil
}
