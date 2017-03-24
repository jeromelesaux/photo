package logger

import (
	"fmt"
	"os"
	"time"
)

func Log(message string) {
	fmt.Fprintf(os.Stderr, "%s : %s\n", time.Now().Format(time.RFC3339), message)
}

func Logf(message string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s : "+message, time.Now().Format(time.RFC3339), args)
}
