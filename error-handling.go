package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
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

func (e *ErrorHandler) InitLog(c *Params) {
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

func (e *ErrorHandler) Msg(levelNum int, message string) {
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

func (ac *appContext) DBErrorCheck(err error, query string, c *gin.Context) bool {
	ac.Log.Msg(0, "Query: "+query)

	ac.Log.Msg(0, "-----------------------------")
	switch err {
	case nil:
		ac.Log.Msg(0, "Query good")
	case sql.ErrNoRows:
		ac.Log.Msg(1, "Now rows returned")
	default:
		ac.Log.Msg(0, "----------ERROR--------------")
		ac.Log.Msg(3, "DB Query failed: "+err.Error())
		ac.AbortMsg(500, err, c)
		return false
	}
	ac.Log.Msg(0, "-----------------------------")
	return true
}

func (ac *appContext) AbortMsg(code int, err error, c *gin.Context) bool {
	var errFile string
	var errMsg string

	if ac.ConfigData.Debug > 0 {
		errMsg = err.Error()
	}

	switch code {
	case http.StatusNotFound:
		errFile = "errors/404"
	case http.StatusNotAcceptable:
		errFile = "errors/verification"
	case http.StatusInternalServerError:
		errFile = "errors/500"
	}

	ac.Log.Msg(3, "Aborting: "+err.Error())

	c.HTML(code, errFile, gin.H{
		"error": errMsg,
	})

	c.Error(err)
	c.Abort()

	return false
}
