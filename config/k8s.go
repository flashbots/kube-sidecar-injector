package config

type K8S struct {
	Namespace                        string `yaml:"namespace,omitempty"`
	ServiceName                      string `yaml:"serviceName,omitempty"`
	ServicePortNumber                int32  `yaml:"servicePortNumber,omitempty"`
	MutatingWebhookConfigurationName string `yaml:"mutatingWebhookConfigurationName,omitempty"`
}
