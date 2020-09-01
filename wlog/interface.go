package wlog

import "net/http"

type ILogger interface {
	Critical(msg string, w http.ResponseWriter, r *http.Request) error
	Error(msg string, w http.ResponseWriter, r *http.Request) error
	NotFound(msg string, w http.ResponseWriter, r *http.Request) error
	Warning(msg string, w http.ResponseWriter, r *http.Request) error
	Notice(msg string, w http.ResponseWriter, r *http.Request) error
	Debug(msg string, w http.ResponseWriter, r *http.Request) error
	sendToDestination(msg string, r *http.Request)
}
