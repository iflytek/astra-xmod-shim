package log

import (
	"fmt"
	"os"
	"path/filepath"

	config "modserv-shim/internal/dto/config"

	"gopkg.in/natefinch/lumberjack.v2"

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

	// 2. 配置日志输出（文件+控制台）
	core := buildLogCore(cfg)

	// 3. 配置日志选项（调用行号等）
	options := buildZapOptions(cfg)

	// 4. 初始化全局Logger
	globalLogger = zap.New(core, options...)
	sugarLogger = globalLogger.Sugar()

	return nil
}

// 为缺失的配置项设置默认值
func setDefaultConfig(cfg *conf.LogConfig) error {
	if cfg.Level == "" {
		cfg.Level = "info" // 默认info级别
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // 默认单个文件100MB
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 7 // 默认保留7天
	}
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	return nil
}

// 构建日志核心（输出目标、编码、级别）
func buildLogCore(cfg *conf.LogConfig) zapcore.Core {
	// 日志文件输出（带轮转）
	fileWriter := &lumberjack.Logger{
		Filename:  cfg.OutputPath,
		MaxSize:   cfg.MaxSize,
		MaxAge:    cfg.MaxAge,
		Compress:  cfg.Compress,
		LocalTime: true,
	}
	fileSyncer := zapcore.AddSync(fileWriter)

	// 日志编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder, // 级别大写（INFO/WARN/ERROR）
		EncodeTime:     zapcore.ISO8601TimeEncoder,  // 时间格式：2006-01-02T15:04:05.000Z0700
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 调用者信息简写（如pkg/file.go:123）
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel // 非法级别默认info
	}

	// 构建核心（文件输出）
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 结构化JSON格式
		fileSyncer,
		level,
	)

	// 如果需要同时输出到控制台，添加控制台输出
	if cfg.EnableConsole {
		consoleSyncer := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig), // 控制台用易读格式
			consoleSyncer,
			level,
		)
		core = zapcore.NewTee(core, consoleCore) // 多输出源合并
	}

	return core
}

// 构建zap选项（调用行号等）
func buildZapOptions(cfg *conf.LogConfig) []zap.Option {
	var options []zap.Option
	if cfg.ShowLine {
		options = append(options, zap.AddCaller())      // 显示调用者信息
		options = append(options, zap.AddCallerSkip(1)) // 跳过当前包层级，显示真实调用位置
	}
	// 开发环境可添加堆栈跟踪（错误级别以上）
	// options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	return options
}

// 以下为常用日志方法封装（SugaredLogger，易用性优先）

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
