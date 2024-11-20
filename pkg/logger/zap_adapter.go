package logger

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
)

type ZapLogSink struct {
	logger *zap.SugaredLogger
}

func NewZapLogger(logger *zap.Logger) logr.Logger {
	return logr.New(&ZapLogSink{
		logger: logger.Sugar(),
	})
}

func (z *ZapLogSink) Init(info logr.RuntimeInfo) {
}

func (z *ZapLogSink) Enabled(level int) bool {
	return true
}

func (z *ZapLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
	z.logger.Infow(msg, append(keysAndValues, "log_level", level)...)
}

func (z *ZapLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
	z.logger.Errorw(msg, append(keysAndValues, "error", err)...)
}

func (z *ZapLogSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return &ZapLogSink{
		logger: z.logger.With(keysAndValues...),
	}
}

func (z *ZapLogSink) WithName(name string) logr.LogSink {
	return &ZapLogSink{
		logger: z.logger.With("logger_name", name),
	}
}
