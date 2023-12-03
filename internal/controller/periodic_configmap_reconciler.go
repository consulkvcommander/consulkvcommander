package controller

import (
	"context"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"time"
)

func SetupPeriodicConfigMapReconciler(k8sClient client.Client) chan event.GenericEvent {
	periodicReconcilerChan := make(chan event.GenericEvent)
	go func(ch chan event.GenericEvent) {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ticker.C:
				broadcastReconcile(k8sClient, ch)
			}
		}
	}(periodicReconcilerChan)
	return periodicReconcilerChan
}

func broadcastReconcile(k8sClient client.Client, ch chan event.GenericEvent) {
	list := &sascomv1.KVGroupList{}
	listOpts := []client.ListOption{
		client.InNamespace(""),
	}

	ctx := context.Background()

	if err := k8sClient.List(ctx, list, listOpts...); err != nil {
		return
	}
	for _, item := range list.Items {
		item := item
		go func() {
			ch <- event.GenericEvent{Object: &item}
		}()

	}
}
