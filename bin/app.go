package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/das6ng/cfnssync/cf"
	"github.com/das6ng/cfnssync/k8s"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

var version = "v0.1.7"

const (
	flagMode                   = "mode"
	flagExclude                = "exclude"
	flagInclude                = "include"
	flagCloudflareZone         = "cloudflare-zone"
	flagCloudflareAPIToken     = "cloudflare-api-token"
	flagCloudflarePullInterval = "cloudflare-pull-interval"
)

var app = &cli.Command{
	Name:    "cf-ns-sync",
	Usage:   "Keep k8s ingress noted DNS synced to Cloudflare.com",
	Version: version,
	Action:  appMain,
	Authors: []any{"Dash Wong [dashengyeah@hotmail.com]"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  flagMode,
			Value: "exclude",
			Usage: "[exclude]|include",
			Validator: func(s string) error {
				return lo.If(!lo.Contains([]string{
					"exclude", "include",
				}, s), errors.New("invalid mode")).Else(nil)
			},
		},
		&cli.StringSliceFlag{
			Name:  flagExclude,
			Value: []string{"kube-system", "kube-public", "kube-node-lease"},
			Usage: "monitor all namespace except specified by this flag",
		},
		&cli.StringSliceFlag{
			Name:  flagInclude,
			Value: []string{"default"},
			Usage: "monitor namespace specified by this flag",
		},
		&cli.StringFlag{
			Name:     flagCloudflareZone,
			Required: true,
			Usage:    "DNS zone name managed by Cloudflare.com",
		},
		&cli.StringFlag{
			Name:     flagCloudflareAPIToken,
			Required: true,
			Usage:    "api-token of Cloudflare.com, need ZONE-EDIT ZONE-READ access to specified Zone",
		},
		&cli.StringFlag{
			Name:  flagCloudflarePullInterval,
			Value: "10m",
			Usage: "pull remote cloudflare DNS name list interval (5s/2m3s/3h..., default:10m)",
		},
	},
}

func appMain(ctx context.Context, c *cli.Command) (err error) {
	cfZone := c.String(flagCloudflareZone)
	cfToken := c.String(flagCloudflareAPIToken)
	pullInterval := c.String(flagCloudflarePullInterval)
	cfStatus, err := cf.NewZone(ctx, cfZone, cfToken, pullInterval)
	if err != nil {
		slog.ErrorContext(ctx, "connect to cloudflare failed", "zone", cfZone, "err", err.Error())
		os.Exit(1)
	}

	excludeNameSpaces := c.StringSlice(flagExclude)
	clientset, err := k8s.NewClientSet(ctx)
	var nsEv <-chan k8s.Event
	switch c.String(flagMode) {
	case "exclude":
		nsEv, err = k8s.WatchNamespace(ctx, clientset, excludeNameSpaces...)
		if err != nil {
			slog.ErrorContext(ctx, "watch namespace got error", "err", err.Error())
			os.Exit(1)
		}
	case "include":
		nsList := c.StringSlice(flagInclude)
		ev := make(chan k8s.Event, 2)
		nsEv = ev
		go func() {
			for _, ns := range nsList {
				ev <- k8s.Event{Type: k8s.EvList, Res: k8s.ResNamespace, Name: ns}
			}
		}()
	default:
		err = errors.New("invalid mode")
		slog.ErrorContext(ctx, "invalid mode", "mode", c.String("mode"))
		return
	}

	dnsEv := k8s.WatchIngress(ctx, clientset, nsEv)
	for ev := range dnsEv {
		slog.InfoContext(ctx, "in-cluster DNS record event", "ns", ev.NS, "name", ev.Name, "content", ev.Value, "event", ev.Type)
		if ev.Type != k8s.EvAdded {
			continue
		}
		cfStatus.Sync(ctx, ev.Name, ev.Value)
	}
	return
}
