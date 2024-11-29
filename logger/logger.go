package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"strings"
)

var l *zap.SugaredLogger

type Config struct {
	// Path   string
	// Rotate string
	Format string
	Level  string
	Output io.Writer
}

func Build(c Config) {
	// encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var encoder zapcore.Encoder
	if strings.ToLower(c.Format) == "json" {
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.TimeKey = "ts"
		encoderCfg.CallerKey = "caller"
		encoderCfg.MessageKey = "msg"
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}
	if c.Output == nil {
		c.Output = os.Stdout
	}
	syncer := zapcore.AddSync(c.Output)

	// default level: info = 0
	level, _ := zapcore.ParseLevel(c.Level)
	core := zapcore.NewCore(encoder, syncer, level)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l = log.Sugar()
}

func Debug(args ...interface{}) {
	l.Debug(args...)
}
func Debugf(template string, args ...interface{}) {
	l.Debugf(template, args...)
}
func Info(args ...interface{}) {
	l.Info(args...)
}
func Infof(template string, args ...interface{}) {
	l.Infof(template, args...)
}
func Warn(args ...interface{}) {
	l.Warn(args...)
}
func Warnf(template string, args ...interface{}) {
	l.Warnf(template, args...)
}
func Error(args ...interface{}) {
	l.Error(args...)
}
func Errorf(template string, args ...interface{}) {
	l.Errorf(template, args...)
}
func DPanic(args ...interface{}) {
	l.DPanic(args...)
}
func DPanicf(template string, args ...interface{}) {
	l.DPanicf(template, args...)
}
func Panic(args ...interface{}) {
	l.Panic(args...)
}
func Panicf(template string, args ...interface{}) {
	l.Panicf(template, args...)
}
func Fatal(args ...interface{}) {
	l.Fatal(args...)
}
func Fatalf(template string, args ...interface{}) {
	l.Fatalf(template, args...)
}

func Sync() error {
	return l.Sync()
}
