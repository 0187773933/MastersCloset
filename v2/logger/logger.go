// Package logger is a thin leveled logger over logrus. Unlike v1 it does not
// hardcode "./logs": the daily file lives under paths.LogDir() (~/.config/mct/
// logs), and the timezone used for timestamps is passed in from config rather
// than baked into a package-level var.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/0187773933/MastersCloset/v2/paths"
	logrus "github.com/sirupsen/logrus"
)

// Log is the process-wide logger. Init must be called once before use; GetLogger
// lazily initializes with defaults so package-level callers never get a nil.
var Log *logrus.Logger

var (
	location   = time.Local
	outputFile *os.File
	mu         sync.Mutex
)

const packageRoot = "github.com/0187773933/MastersCloset/"

type textFormatter struct{}

func (f *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ts := time.Now().In(location).Format("02Jan2006 === 15:04:05.000")
	ts = strings.ToUpper( ts )
	if entry.Caller != nil {
		fn := entry.Caller.Function
		if parts := strings.Split(fn, packageRoot); len(parts) > 1 {
			fn = parts[1]
		}
		return []byte(fmt.Sprintf("%s === %s():%d === %s\n", ts, fn, entry.Caller.Line, entry.Message)), nil
	}
	return []byte(fmt.Sprintf("%s === %s\n", ts, entry.Message)), nil
}

type teeWriter struct{ mu sync.Mutex }

func (w *teeWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := os.Stdout.Write(p)
	if outputFile != nil {
		outputFile.Write(p)
	}
	return n, err
}

// Init configures the logger. level is "debug" or anything else (info default);
// timeZone is an IANA name like "America/New_York" (falls back to local).
func Init(level string, timeZone string) {
	mu.Lock()
	defer mu.Unlock()
	if timeZone != "" {
		if loc, err := time.LoadLocation(timeZone); err == nil {
			location = loc
		}
	}
	Log = logrus.New()
	logFile := filepath.Join(paths.LogDir(), time.Now().In(location).Format("20060102")+".log")
	outputFile, _ = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if strings.ToLower(level) == "debug" {
		Log.SetReportCaller(true)
		Log.SetLevel(logrus.DebugLevel)
	} else {
		Log.SetReportCaller(false)
		Log.SetLevel(logrus.InfoLevel)
	}
	Log.SetFormatter(&textFormatter{})
	Log.SetOutput(&teeWriter{})
}

// GetLogger returns Log, initializing with defaults if Init was never called.
func GetLogger() *logrus.Logger {
	if Log == nil {
		Init(os.Getenv("LOG_LEVEL"), "")
	}
	return Log
}

// Close releases the daily log file handle.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if outputFile != nil {
		outputFile.Close()
		outputFile = nil
	}
}
