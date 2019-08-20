package main

import (
	"./Config"
	"./GinHTMLRender"
	"./Logger"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
)

type contactInfo struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       int    `json:"phone"`
	OfficePhone int    `json:"office_phone"`
	City        string `json:"city"`
	State       string `json:"state"`
	Zip         string `json:"zip"`
	Enabled     bool   `json:"enabled"`
}

type appContext struct {
	DB           *sql.DB
	ConfigData   Config.Params
	Log          Logger.ErrorHandler
	contactsInfo []contactInfo
}

func (ac *appContext) DBErrorCheck(err error, query string) {
	ac.Log.LogMsg(0, "Query: "+query)

	switch err {
	case nil:
		ac.Log.LogMsg(0, "Query good")
	case sql.ErrNoRows:
		ac.Log.LogMsg(1, "Now rows returned")
	default:
		ac.Log.LogMsg(5, "DB Query failed: "+err.Error())
	}
}

func (ac *appContext) LoadAppDefaults() {
	// Load account types for sign up
	query :=
		"select json_agg(to_jsonb(r)) from (" +
			"select " +
			"id, first_name, last_name, phone, office_phone, " +
			"city, state, zip, enabled " +
			"from contacts " +
			"where " +
			"enabled)r"

	var jsondata string

	row := ac.DB.QueryRow(query)
	err := row.Scan(&jsondata)

	ac.Log.LogMsg(0, "json data is: "+jsondata)

	ac.DBErrorCheck(err, query)
	if err := json.Unmarshal([]byte(jsondata), &ac.contactsInfo); err != nil {
		panic(err)
	}

}

func main() {
	context := &appContext{
		DB:         nil,
		ConfigData: Config.LoadConfigFile("config.json"),
	}

	context.Log.InitLog(&context.ConfigData)

	context.Log.LogMsg(1, "Starting Advanced.ID web server ")

	InitDB(context)

	context.LoadAppDefaults()

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

	_ = r.Run(context.ConfigData.ListenIP + ":" + context.ConfigData.ListenPort)
}

func InitDB(c *appContext) {
	var err error

	connStr := "postgres://" + c.ConfigData.SQL.User + ":" + c.ConfigData.SQL.Password + "@" +
		c.ConfigData.SQL.Host + ":" + c.ConfigData.SQL.Port + "/" +
		c.ConfigData.SQL.DBname + "?sslmode=disable"

	c.Log.LogMsg(0, "PG Connection String: "+connStr)

	c.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		c.Log.LogMsg(5, err.Error())
	}

	err = c.DB.Ping()
	if err != nil {
		c.Log.LogMsg(5, "DB Ping failed: "+err.Error())
	}

	c.Log.LogMsg(0, "Successfully connected to database [ "+
		c.ConfigData.SQL.DBname+" ]")
}
