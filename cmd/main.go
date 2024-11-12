package main

import (
	"context"
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.toml", "配置文件路径")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func runCmd(_ *cobra.Command, _ []string) error {
	// 加载配置文件
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}
	utils.Json(cfg)

	// 创建上下文和取消函数
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建同步器
	syncer, err := sync.NewBlockSync(
		cfg.Node.URL,
		cfg.Database.DSN,
		cfg.Sync.DebugMode,
		cfg.Sync.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("创建同步器失败: %w", err)
	}

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// 等待中断信号
		<-sigCh
		fmt.Println("\n收到中断信号，正在关闭...")
		cancel()
	}()

	// 启动同步器
	if err := syncer.Start(ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("启动同步器失败: %w", err)
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "程序运行出错: %s\n", err)
		os.Exit(1)
	}
}
