package keyconjurer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/riotgames/key-conjurer/api/consts"
	"github.com/sirupsen/logrus"
)

type loggerSettings struct {
	Level            logrus.Level
	LogstashEndpoint string
}

func newLogger(settings loggerSettings) *logrus.Entry {
	logger := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    &keyConjurerFormatter{},
		Hooks:        make(logrus.LevelHooks),
		ReportCaller: true,
		Level:        settings.Level,
		ExitFunc:     os.Exit,
	}

	if settings.LogstashEndpoint != "" {
		logger.Hooks.Add(newLogStashHook(settings.LogstashEndpoint))
	}

	return logger.WithFields(logrus.Fields{"uuid": uuid.New().String()})
}

type keyConjurerFormatter struct{}

func (*keyConjurerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Most of this code came from https://github.com/sirupsen/logrus/blob/a6c0064cfaf982707445a1c90368f956421ebcf0/json_formatter.go
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	data["time"] = entry.Time.Format(time.RFC3339)
	data["level"] = entry.Level.String()
	data["metadata"] = entry.Message
	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		data["func"] = funcVal
		data["file"] = fileVal
	}
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	output := map[string]interface{}{
		"jsonEvent":        "keyConjurer",
		"keyConjurerEvent": data}
	encoder := json.NewEncoder(b)
	if err := encoder.Encode(output); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return b.Bytes(), nil
}

type logStashHook struct {
	socket    net.Conn
	formatter *keyConjurerFormatter
}

func newLogStashHook(endpoint string) *logStashHook {
	timeoutDialer := &net.Dialer{
		Timeout: time.Second * consts.HttpTimeout,
	}

	conn, err := timeoutDialer.Dial("tcp", endpoint)
	if err != nil {
		fmt.Println("Unable to connect to endpoint. Only logging to Stdout")
		fmt.Println(err.Error())
		conn = nil
	}

	return &logStashHook{
		socket:    conn,
		formatter: &keyConjurerFormatter{},
	}
}

func (*logStashHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (k *logStashHook) Fire(entry *logrus.Entry) error {
	if k.socket == nil {
		return nil
	}

	log, err := k.formatter.Format(entry)

	// Appending 0x0A is necessary for logstash to find the end of the log
	if err == nil {
		k.socket.Write(append(log, 0x0A))
	}

	return nil
}
