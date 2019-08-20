package Logger

import (
	"../Config"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

type ErrorHandler struct {
	Log      *logrus.Logger
	LogLevel int
}

func (e *ErrorHandler) SetLogLevel(level int) {

	switch level {
	case -1:
		e.Log.SetLevel(logrus.TraceLevel)
	case 0:
		e.Log.SetLevel(logrus.DebugLevel)
	case 1:
		e.Log.SetLevel(logrus.InfoLevel)
	case 2:
		e.Log.SetLevel(logrus.WarnLevel)
	case 3:
		e.Log.SetLevel(logrus.ErrorLevel)
	case 4:
		e.Log.SetLevel(logrus.FatalLevel)
	case 5:
		e.Log.SetLevel(logrus.PanicLevel)
	}
}

func (e *ErrorHandler) InitLog(c *Config.Params) {
	e.Log = logrus.New()

	logFile, err := os.OpenFile(c.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Failed to log to file, using default stderr")
	}

	e.Log.Out = logFile

	if c.LogFormat == "json" {
		e.Log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		e.Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	e.SetLogLevel(c.LogLevel)
}

func (e *ErrorHandler) LogMsg(levelNum int, message string) {
	if levelNum >= e.LogLevel {
		switch levelNum {
		case -1:
			e.Log.Trace(message)
		case 0:
			e.Log.Debug(message)
		case 1:
			e.Log.Info(message)
		case 2:
			e.Log.Warn(message)
		case 3:
			e.Log.Error(message)
		case 4:
			e.Log.Fatal(message)
		case 5:
			e.Log.Panic(message)
		}
	}
}

func (e *ErrorHandler) SlackAlert() {
	// TODO fill in with slack webhook to alert error
}
