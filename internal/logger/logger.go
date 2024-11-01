package logger

import "github.com/sirupsen/logrus"

type Logger struct {
	*logrus.Logger
}

// New creates a new logger instance
func New() *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return &Logger{Logger: log}
}

// Field represents a log fiels
type Field struct {
	Key   string
	Value interface{}
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields ...Field) *Logger {
	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}

	return &Logger{Logger: l.Logger.WithFields(logrusFields).Logger}
}
