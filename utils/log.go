package utils

import (
	"os"

	"github.com/dixonwille/wlog"
	"github.com/spf13/viper"
)

var (
	ui       wlog.UI
	levelLog = errorLevel
)

const (
	errorLevel = iota // display when logLevel < 0
	warningLevel      // display when logLevel < 1
	successLevel      // display when logLevel < 2
	infoLevel         // display when logLevel < 3
	logLevel          // display when logLevel < 4
)

// Init logging configuration
func Init() {
	ui = wlog.New(os.Stdin, os.Stdout, os.Stderr)
	ui = wlog.AddPrefix("?", "[ERROR]", "[INFO]", "[LOG]", "", "[RUNNING]", "[SUCCESS]", "[WARNING]", ui)
	ui = wlog.AddColor(wlog.BrightBlue, wlog.BrightRed, wlog.Cyan, wlog.Black, wlog.Black, wlog.Black, wlog.Magenta, wlog.BrightGreen, wlog.BrightYellow, ui)

	levelLog = viper.GetInt("sniffer.log_level")
}

// Log is for logging debug variable or call trace
func Log(message string) {
	if levelLog > logLevel {
		ui.Log(message)
	}
}

// Info is for logging information (like return of function)
func Info(message string) {
	if levelLog > infoLevel {
		ui.Info(message)
	}
}

// Success if for logging when something goes good
func Success(message string) {
	if levelLog > successLevel {
		ui.Success(message)
	}
}

// Warning is for logging when something goes wrong, but code handle this case
func Warning(message string) {
	if levelLog > warningLevel {
		ui.Warn(message)
	}
}

// Error is for logging when something goes wrong and code doesn't handle this
func Error(message string) {
	if levelLog > errorLevel {
		ui.Error(message)
	}
}
