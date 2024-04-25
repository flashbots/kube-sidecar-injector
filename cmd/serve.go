package main

import (
	"fmt"
	"slices"

	"github.com/flashbots/kube-sidecar-injector/config"
	"github.com/flashbots/kube-sidecar-injector/global"
	"github.com/flashbots/kube-sidecar-injector/server"
	"github.com/urfave/cli/v2"
)

const (
	categoryDebug  = "DEBUG:"
	categoryK8S    = "KUBERNETES:"
	categoryServer = "SERVER:"
)

func CommandServe(cfg *config.Config, globalFlags []cli.Flag) *cli.Command {
	var rawServicePortNumber int64

	debugFlags := []cli.Flag{}

	k8sFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryK8S,
			Destination: &cfg.K8S.Namespace,
			EnvVars:     []string{envPrefix + "NAMESPACE"},
			Name:        "namespace",
			Usage:       "namespace in which the injector will run",
			Value:       "default",
		},

		&cli.StringFlag{
			Category:    categoryK8S,
			Destination: &cfg.K8S.ServiceName,
			EnvVars:     []string{envPrefix + "SERVICE_NAME"},
			Name:        "service-name",
			Usage:       "`name` of service to use",
			Value:       global.AppName,
		},

		&cli.Int64Flag{
			Category:    categoryK8S,
			Destination: &rawServicePortNumber,
			EnvVars:     []string{envPrefix + "SERVICE_PORT_NUMBER"},
			Name:        "service-port-number",
			Usage:       "the port `number` on which the k8s service listens on",
			Value:       8443,
		},

		&cli.StringFlag{
			Category:    categoryK8S,
			Destination: &cfg.K8S.MutatingWebhookConfigurationName,
			EnvVars:     []string{envPrefix + "MUTATING_WEBHOOK_CONFIGURATION_NAME"},
			Name:        "mutating-webhook-configuration-name",
			Usage:       "`name` of mutating webhook configuration to use",
			Value:       global.AppName,
		},
	}

	serverFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.ListenAddress,
			EnvVars:     []string{envPrefix + "LISTEN_ADDRESS"},
			Name:        "listen-address",
			Usage:       "`host:port` for the server to listen on",
			Value:       "0.0.0.0:8443",
		},

		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.PathHealthcheck,
			EnvVars:     []string{envPrefix + "PATH_HEALTHCHECK"},
			Name:        "path-healthcheck",
			Usage:       "`path` at which to serve the healthcheck",
			Value:       "/",
		},

		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.PathWebhook,
			EnvVars:     []string{envPrefix + "PATH_WEBHOOK"},
			Name:        "path-webhook",
			Usage:       "`path` at which to serve the webhook",
			Value:       "/mutate",
		},
	}

	flags := slices.Concat(
		debugFlags,
		k8sFlags,
		serverFlags,
	)

	return &cli.Command{
		Name:  "serve",
		Usage: "run the monitor server",
		Flags: flags,

		Before: func(clictx *cli.Context) error {
			for _, i := range cfg.Inject {
				if i.LabelSelector != nil {
					if _, err := i.LabelSelector.LabelSelector(); err != nil {
						return err
					}
				}
				for _, c := range i.Containers {
					if _, err := c.Container(); err != nil {
						return fmt.Errorf("invalid config for container '%s': %w",
							c.Name, err,
						)
					}
				}
			}

			if rawServicePortNumber > 65535 {
				return fmt.Errorf("invalid port service port number: %d", rawServicePortNumber)
			}
			cfg.K8S.ServicePortNumber = int32(rawServicePortNumber)

			return nil
		},

		Action: func(_ *cli.Context) error {
			s, err := server.New(cfg)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}
}
