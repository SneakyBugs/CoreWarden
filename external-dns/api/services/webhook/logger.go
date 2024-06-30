package webhook

import (
	"runtime"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapConversionHook(logger *zap.Logger) *ZapConversionHook {
	return &ZapConversionHook{
		Logger: logger,
	}
}

type ZapConversionHook struct {
	Logger *zap.Logger
}

func (h *ZapConversionHook) Fire(entry *logrus.Entry) error {
	fields := make([]zap.Field, 0, 10)

	for key, value := range entry.Data {
		if key == logrus.ErrorKey {
			fields = append(fields, zap.Error(value.(error)))
		} else {
			fields = append(fields, zap.Any(key, value))
		}
	}

	switch entry.Level {
	case logrus.PanicLevel:
		h.Write(zapcore.PanicLevel, entry.Message, fields, entry.Caller)
	case logrus.FatalLevel:
		h.Write(zapcore.FatalLevel, entry.Message, fields, entry.Caller)
	case logrus.ErrorLevel:
		h.Write(zapcore.ErrorLevel, entry.Message, fields, entry.Caller)
	case logrus.WarnLevel:
		h.Write(zapcore.WarnLevel, entry.Message, fields, entry.Caller)
	case logrus.InfoLevel:
		h.Write(zapcore.InfoLevel, entry.Message, fields, entry.Caller)
	case logrus.DebugLevel, logrus.TraceLevel:
		h.Write(zapcore.DebugLevel, entry.Message, fields, entry.Caller)
	}

	return nil
}

func (h *ZapConversionHook) Write(lvl zapcore.Level, msg string, fields []zap.Field, caller *runtime.Frame) {
	if ce := h.Logger.Check(lvl, msg); ce != nil {
		if caller != nil {
			ce.Caller = zapcore.NewEntryCaller(caller.PC, caller.File, caller.Line, caller.PC != 0)
		}
		ce.Write(fields...)
	}
}

func (h *ZapConversionHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
