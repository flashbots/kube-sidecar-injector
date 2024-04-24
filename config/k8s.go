package config

type K8S struct {
	Namespace                        string `yaml:"namespace"`
	ServiceName                      string `yaml:"service_name"`
	ServicePortNumber                int32  `yaml:"service_port_number"`
	MutatingWebhookConfigurationName string `yaml:"mutating_webhook_configuration_name"`
}
