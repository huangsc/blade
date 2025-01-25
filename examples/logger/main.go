package main

import (
	"context"
	"errors"
	"time"

	"github.com/huangsc/blade/logger"
)

func main() {
	// 创建开发环境的日志实例
	log := logger.NewZapLogger(
		logger.WithLevel(logger.DebugLevel),
		logger.WithDevelopment(true),
		logger.WithFields(
			logger.String("service", "example"),
			logger.String("version", "v1.0.0"),
		),
		logger.WithCaller(1),
		logger.WithStacktrace(logger.ErrorLevel),
	)

	// 使用上下文
	ctx := context.Background()
	log = log.WithContext(ctx)

	// 输出不同级别的日志
	log.Debug("这是一条调试日志",
		logger.String("key", "value"),
		logger.Int("count", 1),
	)

	log.Info("这是一条信息日志",
		logger.Time("now", time.Now()),
		logger.Duration("elapsed", time.Second),
	)

	log.Warn("这是一条警告日志",
		logger.Bool("warning", true),
	)

	// 使用 WithFields 添加字段
	log = log.WithFields(
		logger.String("component", "database"),
		logger.String("operation", "query"),
	)

	// 输出错误日志（包含堆栈信息）
	err := func() error {
		return errors.New("数据库连接失败")
	}()
	log.Error("操作失败",
		logger.Error(err),
		logger.Int("retry", 3),
	)

	// Fatal 日志会导致程序退出
	// log.Fatal("程序遇到致命错误", logger.Error(err))
}
