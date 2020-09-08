package wlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	gelf "github.com/robertkowalski/graylog-golang"
	"github.com/webediads/adsgolib/wcontext"
)

// Graylog is our connection to Graylog
type Graylog struct {
	gelfConnection *gelf.Gelf
}

// NewGraylog will instantiate our logger, setup the graylog connection
func NewGraylog(graylogIPStr string, graylogPortStr string) *Graylog {
	graylogPort, err := strconv.Atoi(graylogPortStr)
	if err != nil {
		panic(err.Error())
	}
	loggerGraylog := new(Graylog)
	loggerGraylog.gelfConnection = gelf.New(gelf.Config{
		GraylogHostname: graylogIPStr,
		GraylogPort:     graylogPort,
	})
	return loggerGraylog
}

// Critical is used for errors that cannot be recovered
func (logger *Graylog) Critical(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// Error is used for errors that cannot be recovered but we can still live with them
func (logger *Graylog) Error(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// NotFound is used when a content or corresponding value was not found
func (logger *Graylog) NotFound(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - not found"))
	}
	return nil
}

// Warning is used for errors that have been recovered
func (logger *Graylog) Warning(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Notice is mainly used internally for debugging to console
func (logger *Graylog) Notice(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Debug is mainly used internally for debugging to console
func (logger *Graylog) Debug(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// sendToGraylog formats and sends a message to graylog along with the filename, line number, etc
func (logger *Graylog) sendToDestination(msg string, r *http.Request) {
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
					logIP, _ = ctx.Value(wcontext.ContextKeyRequestIP).(string)
					logReferer, _ = ctx.Value(wcontext.ContextKeyReferer).(string)
					logUserAgent, _ = ctx.Value(wcontext.ContextKeyUserAgent).(string)
					logURL = r.URL.RequestURI()
				}

				debugStack := debug.Stack()

				errorToLog := errorGelf{
					App:          Logger.appName,
					AppGroup:     Logger.appGroupName,
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

				errorToLogJSON, errJSON := json.Marshal(errorToLog)
				if errJSON == nil {
					logger.gelfConnection.Log(string(errorToLogJSON))
				}

			}
		}

	}

}
