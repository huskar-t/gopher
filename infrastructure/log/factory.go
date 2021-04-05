package log
import (
	"github.com/huskar-t/gopher/infrastructure/log/query"
	"github.com/sirupsen/logrus"
	"time"
)

type Message struct {
	Host      string        `json:"host"`
	Module    string        `json:"module"`
	Timestamp string        `json:"timestamp"`
	Message   string        `json:"message"`
	Error     string        `json:"error"`
	Tags      logrus.Fields `json:"tags"`
	Level     string        `json:"level"`
}

// LoggerFactory
type LoggerFactory interface {
	CreateHook() (logrus.Hook, error)
	Query(app, host, module, level, content string, from, to time.Time, offset, limit int, tagCond *query.Condition) (total int64, items []Message, err error)
}

var loggerFactory LoggerFactory

func SetLoggerFactory(f LoggerFactory) error {
	loggerFactory = f
	hook, err := f.CreateHook()
	if err != nil {
		return err
	}
	logger.AddHook(hook)
	return nil
}

func GetLoggerFactory() LoggerFactory {
	return loggerFactory
}

type noopHook struct {
}

func (n noopHook) Levels() []logrus.Level {
	return nil
}

func (n noopHook) Fire(entry *logrus.Entry) error {
	return nil
}

type noopLoggerFactory struct {
}

func (noopLoggerFactory) CreateHook() (logrus.Hook, error) {
	return &noopHook{}, nil
}
func (noopLoggerFactory) Query(app, host, module, level, content string, from, to time.Time, offset, limit int, tagCond *query.Condition) (total int64, items []Message, err error) {
	return
}

func init() {
	loggerFactory = &noopLoggerFactory{}
}
