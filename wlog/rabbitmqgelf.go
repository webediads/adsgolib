package wlog

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/streadway/amqp"
	"github.com/webediads/adsgolib/wcontext"
)

// RabbitMqGelf is our connection to Graylog
type RabbitMqGelf struct {
	channel *amqp.Channel
}

// NewRabbitMqGelf will instantiate our logger, setup the rabbitmq connection and channel
func NewRabbitMqGelf(rabbitGelfHostStr string, rabbitGelfUserStr string, rabbitGelfPasswordStr string) *RabbitMqGelf {

	// pour le tls : + fichiers dans ./etc/rabbitmqgelf du client
	// https://stackoverflow.com/questions/62436071/tls-handshake-failure-when-enabling-tls-for-rabbitmq-with-streadway-amqp
	// https://github.com/streadway/amqp/issues/455

	conn, err := amqp.Dial("amqps://" + rabbitGelfUserStr + ":" + rabbitGelfPasswordStr + "@" + rabbitGelfHostStr + ":5671/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	loggerRabbitMqGelf := new(RabbitMqGelf)
	loggerRabbitMqGelf.channel = ch
	return loggerRabbitMqGelf
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Critical is used for errors that cannot be recovered
func (logger *RabbitMqGelf) Critical(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// Error is used for errors that cannot be recovered but we can still live with them
func (logger *RabbitMqGelf) Error(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Software Failure. Press left mouse button to continue.\nGuru Meditation #00000025.65045338"))
	}
	return nil
}

// NotFound is used when a content or corresponding value was not found
func (logger *RabbitMqGelf) NotFound(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	if w != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - not found"))
	}
	return nil
}

// Warning is used for errors that have been recovered
func (logger *RabbitMqGelf) Warning(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Notice is mainly used internally for debugging to console
func (logger *RabbitMqGelf) Notice(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// Debug is mainly used internally for debugging to console
func (logger *RabbitMqGelf) Debug(msg string, w http.ResponseWriter, r *http.Request) error {
	logger.sendToDestination(msg, r)
	return nil
}

// sendToGraylog formats and sends a message to graylog along with the filename, line number, etc
func (logger *RabbitMqGelf) sendToDestination(msg string, r *http.Request) {
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

				q, err := logger.channel.QueueDeclare(
					"log-messages", // name
					false,          // durable
					false,          // delete when unused
					false,          // exclusive
					false,          // no-wait
					nil,            // arguments
				)
				failOnError(err, "Failed to declare a queue")

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
				fmt.Println(errorToLog)

				body := "Hello World!"
				err = logger.channel.Publish(
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(body),
					})
				failOnError(err, "Failed to publish a message")

			}
		}

	}

}
