package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/cloudflare/cloudflare-go"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelWarn,
	})))
}

const (
	nsAnnotionNameKey = "cf-ns-sync/name"
	nsAnnotionValKey  = "cf-ns-sync/value"
)

func main() {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("init InClusterConfig failed", "err", err)
		return
	}
	cliSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.Error("init client set failed", "err", err)
		return
	}

	watch, err := cliSet.NetworkingV1().Ingresses("default").Watch(context.Background(), metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		slog.Error("init watch failed", "err", err)
		return
	}

	for v := range watch.ResultChan() {
		ing := v.Object.(*netv1.Ingress)
		nsName := ing.Annotations[nsAnnotionNameKey]
		nsVal := ing.Annotations[nsAnnotionValKey]
		if nsName == "" || nsVal == "" {
			continue
		}
		// ing.Spec.
		slog.Warn("event comming", "type", v.Type, "obj_name", ing.Name, "annotions", ing.GetAnnotations())
	}
}

func sync2Cloudflare(ctx context.Context, name, val string) {
	// Construct a new API object using a global API key
	api, err := cloudflare.New(os.Getenv("CLOUDFLARE_API_KEY"), os.Getenv("CLOUDFLARE_API_EMAIL"))
	// alternatively, you can use a scoped API token
	// api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	_ = api
	// api.GetDNSRecord(ctx, &cloudflare.ResourceContainer{
	// 	Level:      cloudflare.ZoneRouteLevel,
	// 	Identifier: "",
	// 	Type:       "A",
	// })
}
