package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	middleware "git.webedia-group.net/tools/adsgolib/middleware"

	gelf "github.com/robertkowalski/graylog-golang"
)

// Wrapper is a struct containing the required resources for logging to screen/logfile/remote syslog
type wrapper struct {
	realLogger   log.Logger
	graylog      *gelf.Gelf
	appName      string
	appGroupName string
}

// Logger is our application Wrapper object
var Logger = &wrapper{}

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

// Init will instantiate our logger, setup the graylog connection
func (logger *wrapper) Init(graylogIPStr string, graylogPortStr string, appName string, appGroupName string) error {
	graylogPort, err := strconv.Atoi(graylogPortStr)
	if err != nil {
		return err
	}
	logger.graylog = gelf.New(gelf.Config{
		GraylogHostname: graylogIPStr,
		GraylogPort:     graylogPort,
	})
	logger.appName = appName
	logger.appGroupName = appGroupName
	return nil
}

// Critical is used for errors that cannot be recovered
func (logger *wrapper) Critical(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// Error is used for errors that cannot be recovered but we can still live with them
func (logger *wrapper) Error(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// NotFound is used when a content or corresponding value was not found
func (logger *wrapper) NotFound(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - not found"))
	}
	return nil
}

// Warning is used for errors that have been recovered
func (logger *wrapper) Warning(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	return nil
}

// Notice is mainly used internally for debugging to console
func (logger *wrapper) Notice(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	return nil
}

// Debug is mainly used internally for debugging to console
func (logger *wrapper) Debug(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToGraylog(msg, r)
	return nil
}

// sendToGraylog formats and sends a message to graylog along with the filename, line number, etc
func (logger *wrapper) sendToGraylog(msg string, r *http.Request) {
	pc, _, _, ok1 := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok1 && details != nil {

		fmt.Println("error sent to Graylog")

		// details.Name() contient le nom de la méthode appelée juste avant
		regex, _ := regexp.Compile(".[a-z]+$")
		previousFunction := regex.FindString(details.Name())
		if previousFunction != "" {
			previousFunction = strings.ToLower(previousFunction)
			logLevel, levelFound := logLevels[previousFunction]
			if !levelFound {
				logLevel = 7
			}

			_, fileName, lineNumber, ok2 := runtime.Caller(2)

			var ok3 = true
			if ok2 {
				if strings.HasSuffix(fileName, "recoverer.go") {
					_, fileName, lineNumber, ok3 = runtime.Caller(5)
				}
			}
			if ok3 {

				// default values
				logIP := "unknown"
				logReferer := "unknown"
				logUserAgent := "unknown"
				logURL := "unknown"

				if r != nil {
					ctx := r.Context()
					logIP, _ = ctx.Value(middleware.ContextKeyRequestIP).(string)
					logReferer, _ = ctx.Value(middleware.ContextKeyReferer).(string)
					logUserAgent, _ = ctx.Value(middleware.ContextKeyUserAgent).(string)
					logURL = r.URL.RequestURI()
				}

				debugStack := debug.Stack()

				errorToLog := errorGelf{
					App:          logger.appName,
					AppGroup:     logger.appGroupName,
					ShortMessage: msg,
					FullMessage:  string(debugStack),
					IPAddress:    logIP,
					Level:        logLevel,
					Line:         lineNumber,
					Source:       fileName,
					URL:          logURL,
					URLReferer:   logReferer,
					UserAgent:    logUserAgent,
				}

				fmt.Println("sent to graylog maybe")

				errorToLogJSON, errJSON := json.Marshal(errorToLog)
				if errJSON == nil {
					logger.graylog.Log(string(errorToLogJSON))
				}

			}
		}

	}

}
