package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/channelserver/pkg/config"
	"github.com/rancher/channelserver/pkg/server"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
	URLs      = cli.StringSlice{
		"channels.yaml",
	}
	RefreshInterval      string
	ListenAddress        string
	SubKey               cli.StringSlice
	ChannelServerVersion string
)

func main() {
	app := cli.NewApp()
	app.Name = "Channel Server"
	app.Version = fmt.Sprintf("%s (%s)", Version, GitCommit)
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:   "url",
			EnvVar: "URL",
			Value:  &URLs,
		},
		cli.StringSliceFlag{
			Name:   "config-key",
			EnvVar: "SUBKEY",
			Value:  &SubKey,
		},
		cli.StringFlag{
			Name:        "refresh-interval",
			EnvVar:      "REFRESH_INTERVAL",
			Value:       "15m",
			Destination: &RefreshInterval,
		},
		cli.StringFlag{
			Name:        "listen-address",
			EnvVar:      "LISTEN_ADDRESS",
			Value:       "0.0.0.0:8080",
			Destination: &ListenAddress,
		},
		cli.StringFlag{
			Name:        "channel-server-version",
			EnvVar:      "CHANNEL_SERVER_VERSION",
			Destination: &ChannelServerVersion,
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	logrus.SetOutput(os.Stderr)
	ctx := signals.SetupSignalHandler(context.Background())

	intval, err := time.ParseDuration(RefreshInterval)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %s", RefreshInterval)
	}

	configs := map[string]*config.Config{}
	for _, key := range SubKey {
		cfg, err := config.NewConfig(ctx, key, intval, ChannelServerVersion, URLs...)
		if err != nil {
			return err
		}
		configs[key] = cfg
	}

	return server.ListenAndServe(ctx, ListenAddress, configs)
}
