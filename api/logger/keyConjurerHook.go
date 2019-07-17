package logger

import (
	"fmt"
	"keyconjurer-lambda/consts"
	"net"

	"github.com/sirupsen/logrus"
)

type LogStashHook struct {
	socket    net.Conn
	formatter *KeyConjurerFormatter
}

func NewLogStashHook() *LogStashHook {
	conn, err := net.Dial("tcp", consts.LogstashEndpoint)
	if err != nil {
		fmt.Println("Unable to connect to endpoint. Only logging to Stdout")
		fmt.Println(err.Error())
		conn = nil
	}
	return &LogStashHook{
		socket:    conn,
		formatter: &KeyConjurerFormatter{}}
}

func (k *LogStashHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (k *LogStashHook) Fire(entry *logrus.Entry) error {
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
