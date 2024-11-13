package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"quai-sync/config"
	"quai-sync/sync"
	"quai-sync/utils"

	"github.com/spf13/cobra"
)

var (
	configFile string
	Version    string
	rootCmd    = &cobra.Command{
		Use:     os.Args[0] + " [-c|--config /path/to/config.toml]",
		Short:   "Quai blockchain synchronization tool",
		RunE:    runCmd,
		Version: Version,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.toml", "config file path")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func runCmd(_ *cobra.Command, _ []string) error {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("load config file failed: %w", err)
	}
	utils.Json(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncer, err := sync.NewBlockSync(
		cfg.Node.URL,
		cfg.Database.DSN,
		cfg.Sync.DebugMode,
		cfg.Sync.BatchSize,
		cfg.GetLogLevel(),
	)
	if err != nil {
		return fmt.Errorf("create syncer failed: %w", err)
	}

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// 等待中断信号
		<-sigCh
		log.Println("received interrupt signal, shutting down...")
		cancel()
	}()

	// 启动同步器
	if err := syncer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("start syncer failed: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "program error: %s\n", err)
		os.Exit(1)
	}
}
