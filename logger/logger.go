package logger

import "github.com/sirupsen/logrus"

func GetLogger(prefix string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{"app": prefix})
}