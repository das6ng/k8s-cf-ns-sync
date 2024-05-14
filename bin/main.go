package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/das6ng/cfnssync"
	"k8s.io/apimachinery/pkg/watch"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	})))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nsList := strings.Split(strings.TrimSpace(os.Getenv("MONITOR_NS")), ",")
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
