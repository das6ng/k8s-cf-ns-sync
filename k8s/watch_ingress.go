package k8s

import (
	"context"
	"errors"
	"log/slog"
	"time"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WatchIngress(ctx context.Context, clientset *kubernetes.Clientset, nsChan <-chan Event) (ev <-chan Event) {
	notif := make(chan Event, 8)
	ev = notif
	go func(ctx context.Context) {
		cancelMap := make(map[string]context.CancelCauseFunc)
		for {
			select {
			case <-ctx.Done():
				slog.InfoContext(ctx, "watch ingress canceled by context")
				close(notif)
				cause := errors.New("context canceled")
				for _, cancel := range cancelMap {
					cancel(cause)
				}
				return
			case e := <-nsChan:
				switch e.Type {
				case EvList, EvAdded:
					c1, cancel := context.WithCancelCause(ctx)
					cancelMap[e.Name] = cancel
					go doWatchIng(c1, clientset, e.Name, notif)
				case EvDeleted:
					cancel := cancelMap[e.Name]
					if cancel != nil {
						cancel(errors.New("namespace deleted"))
						delete(cancelMap, e.Name)
					}
				}
			}
		}
	}(ctx)
	return
}

func doWatchIng(ctx context.Context, clientset *kubernetes.Clientset, ns string, notif chan<- Event) {
	for {
		ingRes := clientset.NetworkingV1().Ingresses(ns)
		ev, err := ingRes.Watch(ctx, metav1.ListOptions{Watch: true})
		if err != nil {
			slog.ErrorContext(ctx, "init watch ingress resource failed, will retry", "ns", ns, "err", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		slog.InfoContext(ctx, "start watching ingress", "ns", ns)
		evChan := ev.ResultChan()
		for {
			select {
			case <-ctx.Done():
				cause := ctx.Err()
				slog.InfoContext(ctx, "watch ingress canceled by context", "ns", ns, "caused_by", cause.Error())
				return
			case e := <-evChan:
				ing := e.Object.(*netv1.Ingress)
				nsName := ing.Annotations[annotionKeyName]
				nsVal := ing.Annotations[annotionKeyValue]
				slog.InfoContext(ctx, "watch ingress event", "ns", ns, "ingress_name", ing.Name, "ns_name", nsName, "ns_value", nsVal, "event", e.Type)
				if nsName == "" || nsVal == "" {
					continue
				}
				switch e.Type {
				case watch.Added, watch.Modified:
					notif <- Event{Type: EvAdded, Res: ResDNSRecord, Name: nsName, Value: nsVal, NS: ns}
				case watch.Deleted:
					notif <- Event{Type: EvDeleted, Res: ResDNSRecord, Name: nsName, Value: nsVal, NS: ns}
				}
			}
		}
	}
}
