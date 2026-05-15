package logger

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	instance *logrus.Logger
	once     sync.Once
)

func Get() *logrus.Logger {
	once.Do(func() {
		instance = logrus.New()

		isDev := os.Getenv("APP_ENV") != "production"

		if isDev {
			instance.SetFormatter(&logrus.TextFormatter{
				DisableColors:   false,
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
			instance.SetOutput(os.Stdout)
			instance.SetLevel(logrus.DebugLevel)
		} else {
			instance.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02T15:04:05Z07:00",
			})
		}
	})
	return instance
}

// Shortcuts
func Info(args ...interface{})         { Get().Info(args...) }
func Infof(f string, a ...interface{}) { Get().Infof(f, a...) }
func Debug(args ...interface{})        { Get().Debug(args...) }
func Error(args ...interface{})        { Get().Error(args...) }
func Warn(args ...interface{})         { Get().Warn(args...) }
func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return Get().WithFields(logrus.Fields(fields))
}

func WithError(err error) *logrus.Entry {
	return Get().WithError(err)
}
