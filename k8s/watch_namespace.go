package k8s

import (
	"context"
	"log/slog"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func WatchNamespace(ctx context.Context, clientset *kubernetes.Clientset, exclude ...string) (ev <-chan Event, err error) {
	nsRes := clientset.CoreV1().Namespaces()
	list, err1 := nsRes.List(ctx, apimetav1.ListOptions{})
	if err1 != nil {
		err = err1
		slog.ErrorContext(ctx, "list namespaces fail", "err", err.Error())
		return
	}
	notif := make(chan Event, 8)
	ev = notif
	ex := lo.SliceToMap(exclude, func(ns string) (string, struct{}) {
		return ns, struct{}{}
	})

	go func() {
		for _, ns := range list.Items {
			if _, ok := ex[ns.Name]; ok {
				continue
			}
			notif <- Event{Type: EvList, Res: ResNamespace, Name: ns.Name}
		}
		slog.InfoContext(ctx, "list namespace finished", "count", len(list.Items))
	}()

	go func(ctx context.Context) {
		for {
			ev, err1 := nsRes.Watch(ctx, apimetav1.ListOptions{Watch: true})
			if err1 != nil {
				err = err1
				slog.ErrorContext(ctx, "watch namespaces fail, will try again", "err", err.Error())
				time.Sleep(100 * time.Millisecond)
				continue
			}
			evChan := ev.ResultChan()
		watchLoop:
			for {
				select {
				case <-ctx.Done():
					slog.InfoContext(ctx, "watch namespace canceled by context")
					close(notif)
					return
				case e := <-evChan:
					ns := e.Object.(*corev1.Namespace)
					if _, ok := ex[ns.Name]; !ok {
						continue watchLoop
					}
					switch e.Type {
					case watch.Added:
						notif <- Event{Type: EvAdded, Res: ResNamespace, Name: ns.Name}
					case watch.Deleted:
						notif <- Event{Type: EvDeleted, Res: ResNamespace, Name: ns.Name}
					case "":
						break watchLoop
					}
				}
			}
			slog.WarnContext(ctx, "watch namespcace got empty event type, will retry watch")
			ev.Stop()
			time.Sleep(100 * time.Millisecond)
		}
	}(ctx)
	return
}
