package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/das6ng/cfnssync"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/watch"
)

func init() {
	var logLv slog.Level
	if err := logLv.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		logLv = slog.LevelWarn
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: logLv <= slog.LevelInfo,
		Level:     logLv,
	})))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := cfnssync.InitCloudflare(ctx); err != nil {
		slog.ErrorContext(ctx, "init cloudflare api fail", "err", err.Error())
		return
	}

	nsList := lo.Filter(strings.Split(strings.TrimSpace(os.Getenv("MONITOR_NS")), ","), func(s string, _ int) bool {
		return s != ""
	})
	if len(nsList) == 0 {
		slog.WarnContext(ctx, "no ns specified, will monitor 'default' ns")
		nsList = append(nsList, "default")
	}
	ch := make(chan cfnssync.IngressEvent, 3)
	for _, ns := range nsList {
		if err := cfnssync.WatchIngress(ctx, ns, ch); err != nil {
			return
		}
	}

	for ev := range ch {
		if ev.Type != watch.Added && ev.Type != watch.Modified {
			continue
		}
		cfnssync.Sync2Cloudflare(ctx, ev.Name, ev.Value)
	}
}
