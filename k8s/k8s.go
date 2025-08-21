package k8s

import (
	"context"
	"log/slog"

	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	nsAnnotionNameKey = "cf-ns-sync/name"
	nsAnnotionValKey  = "cf-ns-sync/value"
)

type EvType string

const (
	EvAdded   EvType = EvType(watch.Added)
	EvDeleted EvType = EvType(watch.Deleted)
	EvList    EvType = EvType("LIST")
)

type ResType string

const (
	ResNamespace ResType = "NAMESPACE"
	ResIngress   ResType = "INGRESS"
	ResDNSRecord ResType = "DNSRECORD"
)

type Event struct {
	Type  EvType
	Res   ResType
	Name  string
	NS    string
	Value string
}

func NewClientSet(ctx context.Context) (clientset *kubernetes.Clientset, err error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		slog.ErrorContext(ctx, "init InClusterConfig failed", "err", err)
		return
	}
	clientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.ErrorContext(ctx, "init client set failed", "err", err)
		return
	}
	return
}
