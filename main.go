package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	kei "github.com/RandomLemon/kei/pkg/kei"

	// 空导入决定哪些适配器/插件被编译进二进制，是否启用由 config.yaml 决定。
	_ "github.com/RandomLemon/kei-plugin-chat-agent"
	_ "github.com/RandomLemon/kei/adapters/mock"
	_ "github.com/RandomLemon/kei/adapters/onebot"

	// storage.type: sqlite 需要显式启用 SQL 方言（GORM + cgo）。
	_ "github.com/RandomLemon/kei/pkg/kei/sql"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := kei.Run(ctx, kei.Options{ConfigFile: *configPath}); err != nil {
		fmt.Fprintln(os.Stderr, "xia:", err)
		os.Exit(1)
	}
}
