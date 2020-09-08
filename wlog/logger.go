package wlog

import (
	"log"
	"net/http"
)

// Wrapper is a struct containing the required resources for logging to screen/logfile/remote syslog
type Wrapper struct {
	destination  ILogger
	realLogger   log.Logger
	appName      string
	appGroupName string
}

// Logger is our application Wrapper object
var Logger = &Wrapper{}

/*
0 critical : l'app ne peut pas ou plus fonctionner
3 error : http 500, impossible de répondre au client
5 warning : on a catché l'erreur
6 notice : hash déjà utilisé, tout ce qui n'empêche pas de continuer ou qui n'est pas une erreur en soi
7 debug : pour nous, pour comprendre ce qui se passe dans un algo en fonction des paramètres par exemple
*/
var logLevels = map[string]int{
	"critical": 0,
	"error":    3,
	"warning":  5,
	"notice":   6,
	"debug":    7,
}

// ILogger is our interface for matching our logger systems
type ILogger interface {
	Critical(msg string, w http.ResponseWriter, r *http.Request) error
	Error(msg string, w http.ResponseWriter, r *http.Request) error
	NotFound(msg string, w http.ResponseWriter, r *http.Request) error
	Warning(msg string, w http.ResponseWriter, r *http.Request) error
	Notice(msg string, w http.ResponseWriter, r *http.Request) error
	Debug(msg string, w http.ResponseWriter, r *http.Request) error
	sendToDestination(msg string, r *http.Request)
}

type errorGelf struct {
	App          string `json:"app"`
	AppGroup     string `json:"app_group"`
	FullMessage  string `json:"full_message"`
	ShortMessage string `json:"short_message"`
	IPAddress    string `json:"ip_address"`
	Level        int    `json:"level"`
	Line         int    `json:"line"`
	Source       string `json:"source"`
	URL          string `json:"url"`
	URLReferer   string `json:"url_referer"`
	UserAgent    string `json:"user_agent"`
}

// SetLogger sets the destination which is a type ILogger
func SetLogger(destination ILogger, appName string, appGroupName string) {
	Logger.destination = destination
	Logger.appName = appName
	Logger.appGroupName = appGroupName
}

// GetLogger returns the destination for quick access
func GetLogger() ILogger {
	return Logger.destination
}
