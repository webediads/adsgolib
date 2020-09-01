package wlog

import (
	"fmt"
	"net/http"
)

// Console is a stupid logger that outputs to stdout
type Console struct {
}

// NewConsole will instantiate our logger
func NewConsole() *Console {
	console := new(Console)
	return console
}

// Critical is used for errors that cannot be recovered
func (logger Console) Critical(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// Error is used for errors that cannot be recovered but we can still live with them
func (logger Console) Error(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// NotFound is used when a content or corresponding value was not found
func (logger Console) NotFound(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// Warning is used for errors that have been recovered
func (logger Console) Warning(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// Notice is mainly used internally for debugging to console
func (logger Console) Notice(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// Debug is mainly used internally for debugging to console
func (logger Console) Debug(msg string, w http.ResponseWriter, r *http.Request) error {
	fmt.Println(msg)
	return nil
}

// sendToGraylog formats and sends a message to graylog along with the filename, line number, etc
func (logger Console) sendToDestination(msg string, r *http.Request) {
	fmt.Println("not called")
}
