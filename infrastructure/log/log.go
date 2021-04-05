package log

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

const ModuleKey = "module"

// GetLogger 系统日志接口
func GetLogger(module string) logrus.FieldLogger {
	return logger.WithField(ModuleKey, module)
}

func AddHook(hook logrus.Hook) {
	logger.AddHook(hook)
}

func SetLevel(level logrus.Level) {
	logger.SetLevel(level)
}

func SetFormatter(formatter logrus.Formatter) {
	logger.SetFormatter(formatter)
}

func init() {
	logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: false, FullTimestamp: true, ForceColors: true, TimestampFormat: "2006-01-02 15:04:05"})
}
