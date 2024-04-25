package main

import (
	"fmt"
	"os"

	"github.com/flashbots/kube-sidecar-injector/config"
	"github.com/flashbots/kube-sidecar-injector/global"
	"github.com/flashbots/kube-sidecar-injector/logutils"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"go.uber.org/zap"
)

const (
	envPrefix = "KUBE_SIDECAR_INJECTOR_"
)

var (
	version = "development"
)

func main() {
	defaultConfigFile := "/etc/" + global.AppName + "/config.yaml"

	cfg := &config.Config{
		Version: version,
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			EnvVars: []string{envPrefix + "CONFIG_FILE"},
			Name:    "config-file",
			Usage:   "config file",
			Value:   defaultConfigFile,
		},

		&cli.StringFlag{
			Destination: &cfg.Log.Level,
			EnvVars:     []string{envPrefix + "LOG_LEVEL"},
			Name:        "log-level",
			Usage:       "logging level",
			Value:       "info",
		},

		&cli.StringFlag{
			Destination: &cfg.Log.Mode,
			EnvVars:     []string{envPrefix + "LOG_MODE"},
			Name:        "log-mode",
			Usage:       "logging mode",
			Value:       "prod",
		},
	}

	commands := []*cli.Command{
		CommandServe(cfg, flags),
		CommandDumpConfig(cfg, flags),
	}

	app := &cli.App{
		Name:    global.AppName,
		Usage:   "Inject sidecar containers into k8s pods",
		Version: version,

		Flags:          flags,
		Commands:       commands,
		DefaultCommand: commands[0].Name,

		Before: func(clictx *cli.Context) error {
			f := clictx.String("config-file")
			if _, err := os.Stat(f); err == nil {
				// read non-CLI config from the file
				_cfg, err := config.ReadFrom(f)
				if err != nil {
					return err
				}
				cfg.Inject = _cfg.Inject
				// read the rest of config
				if err := altsrc.InitInputSourceWithContext(
					flags,
					altsrc.NewYamlSourceFromFlagFunc("config-file"),
				)(clictx); err != nil {
					return err
				}
			}

			// setup logger
			l, err := logutils.NewLogger(&cfg.Log)
			if err != nil {
				return err
			}
			zap.ReplaceGlobals(l)

			return nil
		},

		Action: func(clictx *cli.Context) error {
			return cli.ShowAppHelp(clictx)
		},
	}

	defer func() {
		zap.L().Sync() //nolint:errcheck
	}()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed with error:\n\n%s\n\n", err.Error())
		os.Exit(1)
	}
}
