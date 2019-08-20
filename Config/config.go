package Config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
)

type Params struct {
	Debug              int     `json:"Debug"`       // enable extra debugging beyond log level
	LogFile            string  `json:"LogFile"`     // log file
	LogFormat          string  `json:"LogFormat"`   // text or json
	LogLevel           int     `json:"LogLevel"`    // Min log level to log to file
	APIkey             string  `json:"APIkey"`      // API key for advanced insight switch API
	ListenIP           string  `json:"ListenIP"`    // IP to listen on
	ListenPort         string  `json:"ListenPort"`  // Port to bind to
	EpochWindow        int     `json:"EpochWindow"` // range of secs for allowing an api query
	DefaultRatePerUser float32 `json:"DefaultRatePerUser"`
	SessionHours       int     `json:"SessionHours"`             // how long a session should last
	SessionMaintenance int     `json:"SessionMaintenance"`       // interval of hours to run session clean up
	SlackChannel       string  `json:"SlackChannel"`             // where to alarm to
	SlackHook          string  `json:"SlackHook"`                // slack hook URI
	MaxCallsEscalate   int64   `json:"MaxCallReportsToEscalate"` // how many before triggering an escalation with the switch API
	SMS                struct {
		Secret string `json:"Secret"` // set in telnyx portal
		URL    string `json:"URL"`    // endpoint for outbound messaging
		From   string `json:"From"`   // originating DID from telnyx
	} `json:"SMS"`
	Redis struct {
		Host string `json:"Host"` // redis host
		Port string `json:"Port"` // redis port
		Size string `json:"Size"` // redis cluster size
	} `json:"Redis"`
	SQL struct {
		Host     string `json:"Host"`     // pgsql host
		Port     string `json:"Port"`     // pgsql port
		DBname   string `json:"DBname"`   // pgsql database name
		User     string `json:"User"`     // pgsql user
		Password string `json:"Password"` // pgsql pwd
	} `json:"SQL"`
}

func LoadConfigFile(file string) Params {
	var c = Params{}

	if _, err := os.Stat(file); err != nil {
		log.Panic(err.Error())
	}

	// var config Params
	configFile, err := os.Open(file)
	defer configFile.Close()

	if err != nil {
		log.Panic(err.Error())
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&c); err != nil {
		log.Panic(err.Error())
	}

	return c
}
