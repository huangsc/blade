package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
	level  Level
	fields []Field
	ctx    context.Context
}

// NewZapLogger 创建一个基于 zap 的日志实现
func NewZapLogger(opts ...Option) Logger {
	options := &Options{
		Level:            InfoLevel,
		TimeFormat:       "2006-01-02 15:04:05.000",
		EnableCaller:     true,
		CallerSkip:       1,
		Development:      false,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	for _, opt := range opts {
		opt(options)
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(options.TimeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建 zap 配置
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(convertLevel(options.Level)),
		Development:      options.Development,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      options.OutputPaths,
		ErrorOutputPaths: options.ErrorOutputPaths,
	}

	if options.Development {
		config.Encoding = "console"
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建 zap logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(options.CallerSkip),
		zap.AddStacktrace(convertLevel(options.StacktraceLevel)),
	)
	if err != nil {
		panic(err)
	}

	// 添加全局字段
	if len(options.Fields) > 0 {
		fields := make([]zap.Field, 0, len(options.Fields))
		for _, field := range options.Fields {
			fields = append(fields, zap.Any(field.Key, field.Value))
		}
		logger = logger.With(fields...)
	}

	return &zapLogger{
		logger: logger,
		level:  options.Level,
		fields: options.Fields,
	}
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	if l.level > DebugLevel {
		return
	}
	l.logger.Debug(msg, convertFields(fields...)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	if l.level > InfoLevel {
		return
	}
	l.logger.Info(msg, convertFields(fields...)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	if l.level > WarnLevel {
		return
	}
	l.logger.Warn(msg, convertFields(fields...)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	if l.level > ErrorLevel {
		return
	}
	l.logger.Error(msg, convertFields(fields...)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	if l.level > FatalLevel {
		return
	}
	l.logger.Fatal(msg, convertFields(fields...)...)
	os.Exit(1)
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	return &zapLogger{
		logger: l.logger,
		level:  l.level,
		fields: l.fields,
		ctx:    ctx,
	}
}

func (l *zapLogger) WithFields(fields ...Field) Logger {
	newLogger := l.logger.With(convertFields(fields...)...)
	return &zapLogger{
		logger: newLogger,
		level:  l.level,
		fields: append(l.fields, fields...),
		ctx:    l.ctx,
	}
}

func (l *zapLogger) WithLevel(level Level) Logger {
	return &zapLogger{
		logger: l.logger,
		level:  level,
		fields: l.fields,
		ctx:    l.ctx,
	}
}

// convertLevel 转换日志级别
func convertLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// convertFields 转换字段
func convertFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}
