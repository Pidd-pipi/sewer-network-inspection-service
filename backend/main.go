package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := loadConfig()
	store := newInspectionStore()
	service := newOpsService(seedOpsRecords())
	metrics := newOpsMetrics()
	notifier := newOpsNotifier(8)
	worker := newOpsWorker(service, 30*time.Second, 2*time.Hour, "system")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Run(ctx)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = notifier.Dispatch(ctx, func(notice OpsNotice) error {
					log.Printf("notice record=%s channel=%s message=%s", notice.RecordID, notice.Channel, notice.Message)
					return nil
				})
			}
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/healthz", newRouter(store))
	mux.Handle("/api/", newRouter(store))
	mux.Handle("/ops/", opsMetricsMiddleware(metrics, newOpsRouter(service)))
	mux.Handle("/", staticHandler())
	if err := serveAddress(":"+cfg.Port, mux); err != nil {
		panic(err)
	}
	worker.Close()
}
