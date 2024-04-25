package main

import (
	"fmt"

	"github.com/flashbots/kube-sidecar-injector/config"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func CommandDumpConfig(cfg *config.Config, globalFlags []cli.Flag) *cli.Command {
	cmd := CommandServe(cfg, globalFlags)

	cmd.Name = "dump-config"
	cmd.Usage = "dump the effective configuration"

	cmd.Action = func(_ *cli.Context) error {
		bytes, err := yaml.Marshal(cfg)
		if err != nil {
			return err
		}

		fmt.Printf("---\n\n%s\n", string(bytes))
		return nil
	}

	return cmd
}
