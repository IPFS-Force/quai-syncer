package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"quai-sync/sync"

	"gorm.io/gorm/logger"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	nodeURL := "https://rpc.quai.network/cyprus1/"
	dbURL := "postgresql://postgres:1234@localhost:5432/quai?connect_timeout=100&sslmode=disable&TimeZone=UTC"

	syncer, err := sync.NewBlockSync(nodeURL, dbURL, false, 5, logger.Info)
	if err != nil {
		log.Fatalf("Failed to create block syncer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("received interrupt signal, shutting down...")
		cancel()
	}()

	// 启动同步器
	if err := syncer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("Sync error: %v", err)
	}
}
