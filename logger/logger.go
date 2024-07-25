package logger

import (
	"fmt"
	"os"
	"time"

	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	"github.com/sirupsen/logrus"
)

const (
	// LogStashFormatter is constant used to format logs as logstash format
	LogStashFormatter = "logstash"
	// TextFormatter is constant used to format logs as simple text format
	TextFormatter = "text"
)

// InitLog initializes the logrus logger
func InitLog(logLevel, formatter string) error {

	switch formatter {
	case LogStashFormatter:
		logrus.SetFormatter(&logrustash.LogstashFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})

	}

	logrus.SetOutput(os.Stdout)

	if level, err := logrus.ParseLevel(logLevel); err != nil {
		logrus.SetLevel(logrus.DebugLevel)
		fmt.Fprintf(os.Stderr, "Error with error : "+err.Error())
		return err
	} else {
		logrus.SetLevel(level)
		return nil
	}
}
