package logger

import (
	"os"

	"github.com/riotgames/key-conjurer/api/consts"

	"github.com/google/uuid"
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

	if consts.LogstashEndpoint != "" {
		logger.Hooks.Add(NewLogStashHook())
	}

	return logger.WithFields(logrus.Fields{
		"client":        client,
		"clientVersion": clientVersion,
		"uuid":          uuid.New().String()})
}
