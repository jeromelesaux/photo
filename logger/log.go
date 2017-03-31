package logger

import (
	"fmt"
	"os"
	"time"
)

func Log(message string) (int, error) {
	return fmt.Fprintf(os.Stderr, "%s : %s\n", time.Now().Format(time.RFC3339), message)
}

func Logf(message string, args ...interface{}) (int, error) {
	formattedMessage := fmt.Sprintf(message, args...)
	return fmt.Fprintf(os.Stderr, "%s : "+formattedMessage, time.Now().Format(time.RFC3339))
}

func LogLn(a ...interface{}) (int, error) {
	currentTime := fmt.Sprintf("%s : ", time.Now().Format(time.RFC3339))
	return fmt.Fprintln(os.Stderr, currentTime, a)
}
