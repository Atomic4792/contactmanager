package main

import (
	"./GinHTMLRender"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
)

type appContext struct {
	DB         *sql.DB
	ConfigData Params
	Log        ErrorHandler
}

func main() {
	context := &appContext{
		DB:         nil,
		ConfigData: LoadConfigFile("config.json"),
	}

	context.Log.InitLog(&context.ConfigData)

	context.Log.Msg(1, "Starting Advanced.ID web server ")

	InitDB(context)

	// context.LoadAppDefaults()

	r := gin.Default()
	htmlRender := GinHTMLRender.New()
	htmlRender.Debug = gin.IsDebugging()
	htmlRender.Layout = "layouts/default"
	htmlRender.TemplatesDir = "templates/" // default
	htmlRender.Ext = ".html"               // default

	// tell gin to use our render
	r.HTMLRender = htmlRender.Create()

	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	r.StaticFS("/assets", http.Dir("./assets"))

	r.GET("/", context.ShowIndex)
	r.GET("/index", context.ShowIndex)
	r.GET("/index.html", context.ShowIndex)
	r.POST("/formData", context.uploadContact)
	r.POST("/saveUpdate", context.saveContact)
	r.POST("/deleteContact", context.deleteContact)
	r.POST("/editContact", context.editContact)

	_ = r.Run(context.ConfigData.ListenIP + ":" + context.ConfigData.ListenPort)
}

func InitDB(c *appContext) {
	var err error

	connStr := "postgres://" + c.ConfigData.SQL.User + ":" + c.ConfigData.SQL.Password + "@" +
		c.ConfigData.SQL.Host + ":" + c.ConfigData.SQL.Port + "/" +
		c.ConfigData.SQL.DBname + "?sslmode=disable"

	c.Log.Msg(0, "PG Connection String: "+connStr)

	c.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		c.Log.Msg(5, err.Error())
	}

	err = c.DB.Ping()
	if err != nil {
		c.Log.Msg(5, "DB Ping failed: "+err.Error())
	}

	c.Log.Msg(0, "Successfully connected to database [ "+
		c.ConfigData.SQL.DBname+" ]")
}
