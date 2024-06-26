package cfnssync

import (
	"context"
	"log/slog"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	nsAnnotionNameKey = "cf-ns-sync/name"
	nsAnnotionValKey  = "cf-ns-sync/value"
)

type IngressEvent struct {
	Type  watch.EventType
	Name  string
	NS    string
	Value string
}

func WatchIngress(ctx context.Context, ns string, ch chan<- IngressEvent) (err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		slog.ErrorContext(ctx, "init InClusterConfig failed", "err", err)
		return
	}
	cliSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.ErrorContext(ctx, "init client set failed", "err", err)
		return
	}

	watch, err := cliSet.NetworkingV1().Ingresses(ns).Watch(ctx, metav1.ListOptions{
		Watch: true,
	})
	if err != nil {
		slog.ErrorContext(ctx, "init watch failed", "ns", ns, "err", err)
		return
	}

	go func(ctx context.Context) {
		slog.InfoContext(ctx, "start watching", "ns", ns)
		for v := range watch.ResultChan() {
			ing := v.Object.(*netv1.Ingress)
			nsName := ing.Annotations[nsAnnotionNameKey]
			nsVal := ing.Annotations[nsAnnotionValKey]
			if nsName == "" || nsVal == "" {
				continue
			}
			slog.InfoContext(ctx, "event comming", "type", v.Type, "obj_name", ing.Name, "annotions", ing.GetAnnotations())
			ch <- IngressEvent{
				Type:  v.Type,
				Name:  nsName,
				NS:    nsVal,
				Value: nsVal,
			}
		}
	}(ctx)
	return
}
