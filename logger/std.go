package logger

import (
	"io"

	"github.com/sirupsen/logrus"
)

// logrus 包的别名，方便后续替换日志库。
type (
	Level       = logrus.Level       // 日志级别
	FieldLogger = logrus.FieldLogger // 字段日志接口
	Logger      = logrus.Logger      // Logger 实例
	Entry       = logrus.Entry       // 日志条目
	Formatter   = logrus.Formatter   // 格式化器接口
	Hook        = logrus.Hook        // Hook 接口
	LevelHooks  = logrus.LevelHooks  // Level 到 Hooks 的映射
	Fields      = logrus.Fields      // 字段映射
	LogFunction = logrus.LogFunction // 按需构造日志的函数类型
)

// stdEntry 返回一个可直接调用完整 Entry API 的日志入口。
func stdEntry() *Entry {
	if entry, ok := GetFieldLogger().(*Entry); ok {
		return entry
	}
	return GetLogger().WithField("service_name", "unknown")
}

// StandardLogger 返回包内维护的全局 logger 实例。
func StandardLogger() *Logger {
	return GetLogger()
}

// SetLevel 设置全局日志级别。
func SetLevel(level Level) {
	GetLogger().SetLevel(level)
}

// GetLevel 返回全局日志级别。
func GetLevel() Level {
	return GetLogger().GetLevel()
}

// IsLevelEnabled 判断指定级别是否启用。
func IsLevelEnabled(level Level) bool {
	return GetLogger().IsLevelEnabled(level)
}

// SetFormatter 设置全局日志格式化器。
func SetFormatter(formatter Formatter) {
	GetLogger().SetFormatter(formatter)
}

// SetOutput 设置全局日志输出目标。
func SetOutput(output io.Writer) {
	GetLogger().SetOutput(output)
}

// SetReportCaller 控制是否输出调用方信息。
func SetReportCaller(reportCaller bool) {
	GetLogger().SetReportCaller(reportCaller)
}

// AddHook 为全局 logger 添加 hook。
func AddHook(hook Hook) {
	GetLogger().AddHook(hook)
}

// ReplaceHooks 替换全局 logger 的 hooks，并返回旧 hooks。
func ReplaceHooks(hooks LevelHooks) LevelHooks {
	return GetLogger().ReplaceHooks(hooks)
}

// WithField 创建带单个字段的日志 Entry。
func WithField(key string, value any) *Entry {
	return stdEntry().WithField(key, value)
}

// WithFields 创建带多个字段的日志 Entry。
func WithFields(fields Fields) *Entry {
	return stdEntry().WithFields(fields)
}

// WithError 创建带 error 字段的日志 Entry。
func WithError(err error) *Entry {
	return stdEntry().WithError(err)
}

// Log 以指定级别记录日志。
func Log(level Level, args ...any) {
	stdEntry().Log(level, args...)
}

// Trace 记录 trace 级别日志。
func Trace(args ...any) {
	stdEntry().Trace(args...)
}

// Debug 记录 debug 级别日志。
func Debug(args ...any) {
	stdEntry().Debug(args...)
}

// Info 记录 info 级别日志。
func Info(args ...any) {
	stdEntry().Info(args...)
}

// Print 记录 print 级别日志。
func Print(args ...any) {
	stdEntry().Print(args...)
}

// Warn 记录 warn 级别日志。
func Warn(args ...any) {
	stdEntry().Warn(args...)
}

// Warning 记录 warn 级别日志（Warn 的别名）。
func Warning(args ...any) {
	stdEntry().Warning(args...)
}

// Error 记录 error 级别日志。
func Error(args ...any) {
	stdEntry().Error(args...)
}

// Fatal 记录 fatal 级别日志并退出进程。
func Fatal(args ...any) {
	stdEntry().Fatal(args...)
}

// Panic 记录 panic 级别日志并触发 panic。
func Panic(args ...any) {
	stdEntry().Panic(args...)
}

// Logf 以格式化字符串在指定级别记录日志。
func Logf(level Level, format string, args ...any) {
	stdEntry().Logf(level, format, args...)
}

// Tracef 记录 trace 级别格式化日志。
func Tracef(format string, args ...any) {
	stdEntry().Tracef(format, args...)
}

// Debugf 记录 debug 级别格式化日志。
func Debugf(format string, args ...any) {
	stdEntry().Debugf(format, args...)
}

// Infof 记录 info 级别格式化日志。
func Infof(format string, args ...any) {
	stdEntry().Infof(format, args...)
}

// Printf 记录 print 级别格式化日志。
func Printf(format string, args ...any) {
	stdEntry().Printf(format, args...)
}

// Warnf 记录 warn 级别格式化日志。
func Warnf(format string, args ...any) {
	stdEntry().Warnf(format, args...)
}

// Warningf 记录 warn 级别格式化日志（Warnf 的别名）。
func Warningf(format string, args ...any) {
	stdEntry().Warningf(format, args...)
}

// Errorf 记录 error 级别格式化日志。
func Errorf(format string, args ...any) {
	stdEntry().Errorf(format, args...)
}

// Fatalf 记录 fatal 级别格式化日志并退出进程。
func Fatalf(format string, args ...any) {
	stdEntry().Fatalf(format, args...)
}

// Panicf 记录 panic 级别格式化日志并触发 panic。
func Panicf(format string, args ...any) {
	stdEntry().Panicf(format, args...)
}

// Logln 以换行风格在指定级别记录日志。
func Logln(level Level, args ...any) {
	stdEntry().Logln(level, args...)
}

// Traceln 记录 trace 级别换行日志。
func Traceln(args ...any) {
	stdEntry().Traceln(args...)
}

// Debugln 记录 debug 级别换行日志。
func Debugln(args ...any) {
	stdEntry().Debugln(args...)
}

// Infoln 记录 info 级别换行日志。
func Infoln(args ...any) {
	stdEntry().Infoln(args...)
}

// Println 记录 print 级别换行日志。
func Println(args ...any) {
	stdEntry().Println(args...)
}

// Warnln 记录 warn 级别换行日志。
func Warnln(args ...any) {
	stdEntry().Warnln(args...)
}

// Warningln 记录 warn 级别换行日志（Warnln 的别名）。
func Warningln(args ...any) {
	stdEntry().Warningln(args...)
}

// Errorln 记录 error 级别换行日志。
func Errorln(args ...any) {
	stdEntry().Errorln(args...)
}

// Fatalln 记录 fatal 级别换行日志并退出进程。
func Fatalln(args ...any) {
	stdEntry().Fatalln(args...)
}

// Panicln 记录 panic 级别换行日志并触发 panic。
func Panicln(args ...any) {
	stdEntry().Panicln(args...)
}

// TraceFn 按需构造并记录 trace 级别日志。
func TraceFn(fn LogFunction) {
	GetLogger().TraceFn(fn)
}

// DebugFn 按需构造并记录 debug 级别日志。
func DebugFn(fn LogFunction) {
	GetLogger().DebugFn(fn)
}

// InfoFn 按需构造并记录 info 级别日志。
func InfoFn(fn LogFunction) {
	GetLogger().InfoFn(fn)
}

// PrintFn 按需构造并记录 print 级别日志。
func PrintFn(fn LogFunction) {
	GetLogger().PrintFn(fn)
}

// WarnFn 按需构造并记录 warn 级别日志。
func WarnFn(fn LogFunction) {
	GetLogger().WarnFn(fn)
}

// WarningFn 按需构造并记录 warn 级别日志（WarnFn 的别名）。
func WarningFn(fn LogFunction) {
	GetLogger().WarningFn(fn)
}

// ErrorFn 按需构造并记录 error 级别日志。
func ErrorFn(fn LogFunction) {
	GetLogger().ErrorFn(fn)
}

// FatalFn 按需构造并记录 fatal 级别日志并退出进程。
func FatalFn(fn LogFunction) {
	GetLogger().FatalFn(fn)
}

// PanicFn 按需构造并记录 panic 级别日志并触发 panic。
func PanicFn(fn LogFunction) {
	GetLogger().PanicFn(fn)
}

// Exit 触发 logrus 退出流程并退出进程。
func Exit(code int) {
	GetLogger().Exit(code)
}
