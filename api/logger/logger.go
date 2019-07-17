package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger(client, clientVersion string, level logrus.Level) *logrus.Entry {
	logger := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &KeyConjurerFormatter{},
		Hooks:        make(logrus.LevelHooks),
		ReportCaller: true,
		Level:        level,
		ExitFunc:     os.Exit}
	// This creates a log entry that enables us to pass around logrus.Entry type
	//  to customize the outputted fields
	return logger.WithFields(logrus.Fields{
		"client":        client,
		"clientVersion": clientVersion})
}
