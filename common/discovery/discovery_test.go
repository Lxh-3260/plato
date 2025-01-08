package discovery

import (
	"context"
	"testing"
	"time"
)

func TestServiceDiscovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() // 30s后取消子协程，防止死循环

	ser := NewServiceDiscovery(ctx)
	defer ser.Close()

	ser.WatchService("/web/", func(key, value string) {}, func(key, value string) {})
	ser.WatchService("/gRPC/", func(key, value string) {}, func(key, value string) {})

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Perform some checks or assertions here if needed
		case <-ctx.Done():
			return
		}
	}
}
