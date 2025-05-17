package utils

import (
	"fmt"
	"log"
	"os"
	"time"
	"github.com/mbbgs/rook/consts"
	"path/filepath"
)

// ANSI color codes
const (
	Red       = "\033[31m"
	Orange    = "\033[38;5;208m"
	NeonGreen = "\033[38;5;46m"
	Reset     = "\033[0m"
)

var logFile *os.File

func InitLogger() error {
	dir, err := GetSessionDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, consts.ROOK_LOG)
	logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("[ROOK]: ")
	return nil
}

// CloseLogger should be deferred in main()
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func colorLogger(color, label, message string, silent bool) {
	now := time.Now().Format("15:04:05")
	formatted := fmt.Sprintf("%s[%s %s]%s %s\n", color, label, now, Reset, message)
	
	// Only print to console if not silent
	if !silent {
		fmt.Print(formatted)
	}

	log.Printf("[%s] %s", label, message)
}

// Quick helpers
func Error(message string) {
	colorLogger(Red, "ERROR", message, false)
}

func ErrorE(err error) {
	if err != nil {
		colorLogger(Red, "ERROR", err.Error(), false)
	}
}

func Warn(message string) {
	colorLogger(Orange, "WARNING", message, false)
}

func Done(message string) {
	colorLogger(NeonGreen, "DONE", message, false)
}

func SilentDone(message string) {
	colorLogger(NeonGreen, "DONE", message, true)
}
