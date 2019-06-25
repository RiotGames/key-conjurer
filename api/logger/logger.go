package logger

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

	"keyconjurer-lambda/consts"

	"github.com/google/uuid"
)

// LogLevel is used to control log verbosity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	CRITICAL
	ERROR
	SILENT
)

// Logger is used to write logs to infosec's Kibana stash
type Logger struct {
	JSONEvent     string
	UUID          string
	Username      string
	Client        string
	ClientVersion string
	socket        net.Conn
	Level         LogLevel
}

// Event is a set of information for a log
type Event struct {
	Level         string `json:"level"`
	Time          string `json:"time"`
	Username      string `json:"username"`
	Metadata      string `json:"metadata"`
	UUID          string `json:"uuid"`
	Client        string `json:"client"`
	ClientVersion string `json:"clientVersion"`
	APIVersion    string `json:"apiVersion"`
}

// LogMessage is used to structure an event into something loggable
type LogMessage struct {
	JSONEvent string `json:"jsonEvent"`
	Event     Event  `json:"keyconjurerEvent"`
}

func (l *LogLevel) getLevel() string {
	switch *l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case CRITICAL:
		return "CRITICAL"
	case ERROR:
		return "ERROR"
	case SILENT:
		return "SILENT"
	default:
		return ""
	}
}

// NewLogger creates a logger that will send its data via TCP to the host
//  on the designated port, level is used to control the verbosity
func NewLogger(eventName, client, clientVersion string, level LogLevel) *Logger {
	conn, err := net.Dial("tcp", consts.LoggerEndpoint)
	if level != SILENT && err != nil {
		log.Println("Only logging to std err")
		log.Println(err)
		conn = nil
	}
	return &Logger{
		JSONEvent:     eventName,
		Level:         level,
		UUID:          uuid.New().String(),
		Client:        client,
		ClientVersion: clientVersion,
		socket:        conn}
}

func (l *Logger) newLogMessage(metadata []string, level LogLevel) *LogMessage {
	return &LogMessage{
		JSONEvent: l.JSONEvent,
		Event: Event{
			Time:          time.Now().Format(time.RFC3339),
			Level:         level.getLevel(),
			Username:      l.Username,
			Metadata:      strings.Join(metadata, ","),
			Client:        l.Client,
			ClientVersion: l.ClientVersion,
			APIVersion:    consts.Version,
			UUID:          l.UUID}}
}

// Debug logs message with DEBUG verbosity
func (l *Logger) Debug(metadata ...string) {
	l.Log(DEBUG, metadata...)
}

// Info logs message with INFO verbosity
func (l *Logger) Info(metadata ...string) {
	l.Log(INFO, metadata...)
}

// Warn logs message with WARN verbosity
func (l *Logger) Warn(metadata ...string) {
	l.Log(WARN, metadata...)
}

// Critical logs message with CRITICAL verbosity
func (l *Logger) Critical(metadata ...string) {
	l.Log(CRITICAL, metadata...)
}

// Error logs message with ERROR verbosity
func (l *Logger) Error(metadata ...string) {
	l.Log(ERROR, metadata...)
}

// Log sends the message across the network si the socket is available
//  and the loggers level isn't SILENT
func (l *Logger) Log(level LogLevel, metadata ...string) {
	if l.Level != SILENT && level >= l.Level {
		msg, _ := json.Marshal(l.newLogMessage(metadata, level))
		log.Println(string(msg))
		if l.socket != nil {
			l.socket.Write(append(msg, 0x0A))
		}
	}
}

// SetUsername sets the username within the logger
func (l *Logger) SetUsername(username string) {
	l.Username = username
}
