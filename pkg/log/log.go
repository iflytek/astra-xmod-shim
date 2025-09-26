// Package log 封装 zap 日志库，专为云原生环境设计：仅输出结构化 JSON 到 stdout
package log

import (
	"os"

	config "modserv-shim/internal/dto/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger
)

// Init 初始化日志系统（基于配置）
func Init(cfg *config.LogConfig) error {
	// 1. 验证并处理配置默认值
	if err := setDefaultConfig(cfg); err != nil {
		return err
	}

	// 2. 构建日志核心（仅输出到 stdout）
	core := buildLogCore(cfg)

	// 3. 配置日志选项（调用行号等）
	options := buildZapOptions(cfg)

	// 4. 初始化全局 Logger
	globalLogger = zap.New(core, options...)
	sugarLogger = globalLogger.Sugar()

	return nil
}

// setDefaultConfig 为缺失的配置项设置默认值
func setDefaultConfig(cfg *config.LogConfig) error {
	if cfg.Level == "" {
		cfg.Level = "info" // 默认 info 级别
	}
	// 云原生环境下无需文件路径、MaxSize、MaxAge 等配置
	return nil
}

// buildLogCore 构建日志核心：仅输出到 stdout，使用 JSON 编码
func buildLogCore(cfg *config.LogConfig) zapcore.Core {
	// 使用 stdout 作为输出目标
	consoleSyncer := zapcore.Lock(zapcore.AddSync(zapcore.Lock(os.Stdout)))

	// 日志编码配置（JSON 结构化日志）
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写级别（info/warn/error）
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 简写调用者（file.go:line）
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel // 非法级别默认 info
	}

	// 构建核心：仅输出到 stdout
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON 格式
		consoleSyncer,
		level,
	)
}

// buildZapOptions 构建 zap 选项（调用行号等）
func buildZapOptions(cfg *config.LogConfig) []zap.Option {
	var options []zap.Option
	if cfg.ShowLine {
		options = append(options, zap.AddCaller())      // 显示调用者信息
		options = append(options, zap.AddCallerSkip(1)) // 跳过当前包层级
	}
	return options
}

// 以下为常用日志方法封装（SugaredLogger）

func Debug(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
}

func Info(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
}

func Warn(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
}

func Error(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
}

func Fatal(template string, args ...interface{}) {
	sugarLogger.Fatalf(template, args...)
}

// Sync 刷新日志缓冲区（程序退出前调用）
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
