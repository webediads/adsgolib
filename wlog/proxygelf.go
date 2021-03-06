package wlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/webediads/adsgolib/wcontext"
)

// ProxyGelf is our connection to Graylog
type ProxyGelf struct {
	url string
}

// NewProxyGelf will instantiate our logger
func NewProxyGelf(url string) *ProxyGelf {

	loggerProxyGelf := new(ProxyGelf)
	loggerProxyGelf.url = url
	return loggerProxyGelf
}

// Critical is used for errors that cannot be recovered
func (logger *ProxyGelf) Critical(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// Error is used for errors that cannot be recovered but we can still live with them
func (logger *ProxyGelf) Error(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// NotFound is used when a content or corresponding value was not found
func (logger *ProxyGelf) NotFound(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - not found"))
	}
	return nil
}

// Warning is used for errors that have been recovered
func (logger *ProxyGelf) Warning(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Notice is mainly used internally for debugging to console
func (logger *ProxyGelf) Notice(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Debug is mainly used internally for debugging to console
func (logger *ProxyGelf) Debug(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// sendToGraylog formats and sends a message to graylog along with the filename, line number, etc
func (logger *ProxyGelf) sendToDestination(msg string, r *http.Request) {
	pc, _, _, ok1 := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if !ok1 || details == nil {
		// fmt.Println("problem ok1 or details")
		return
	}

	// details.Name() contient le nom de la méthode appelée juste avant
	regex, _ := regexp.Compile(".[a-z]+$")
	previousFunction := regex.FindString(details.Name())
	if previousFunction == "" {
		// fmt.Println("problem previousFunction")
		return
	}

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
	if !ok3 {
		// fmt.Println("problem ok3")
		return
	}

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

	requestBody, err := json.Marshal(map[string]string{
		"app":          Logger.appName,
		"app_group":    Logger.appGroupName,
		"message":      msg,
		"level":        strconv.Itoa(logLevel),
		"full_message": strings.Replace(string(debugStack), "[", "", -2), // api aime pas le caractère [, ça casse son json_decode
		"ip_address":   logIP,
		"line":         strconv.Itoa(lineNumber),
		"file":         fileName,
		"url":          logURL,
		"url_referer":  logReferer,
		"user_agent":   logUserAgent,
	})
	if err != nil {
		fmt.Println("error json.Marshal")
		return
	}
	// fmt.Println(string(requestBody))

	transport := http.Transport{
		Dial: dialTimeout,
	}
	client := http.Client{
		Transport: &transport,
	}
	_, err = client.Post(logger.url, "application/x-www-form-urlencoded", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("error post")
	} else {
		fmt.Println("msg send " + msg)
	}
	transport.CloseIdleConnections()

}

func dialTimeout(network, addr string) (net.Conn, error) {
	var timeout = time.Duration(3 * time.Second)
	return net.DialTimeout(network, addr, timeout)
}
