package wlog

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
func NewRabbitMqGelf(rabbitGelfProtocolStr string, rabbitGelfHostStr string, rabbitGelfUserStr string, rabbitGelfPasswordStr string, caCertFile string, clientCertFile string, clientKeyFile string) *RabbitMqGelf {

	var conn *amqp.Connection
	var err error

	if rabbitGelfProtocolStr == "amqps" {
		// pour le tls : + fichiers dans ./etc/rabbitmqgelf du client
		// https://stackoverflow.com/questions/62436071/tls-handshake-failure-when-enabling-tls-for-rabbitmq-with-streadway-amqp
		// https://github.com/streadway/amqp/issues/455

		cert, err := tls.LoadX509KeyPair(".etc/rabbitmqgelf/client_certificate.pem", ".etc/rabbitmqgelf/client_key.pem")

		// Load CA cert
		caCert, err := ioutil.ReadFile(".etc/rabbitmqgelf/cacert.pem") // The same you configured in the rabbit MQ server
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert}, // from tls.LoadX509KeyPair
			RootCAs:      caCertPool,
			CipherSuites: []uint16{
				// openssl s_client -connect rabbitmq:5671 -tls1
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			},
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       true,
			MinVersion:               tls.VersionTLS10,
		}

		conn, err = amqp.DialTLS("amqps://"+rabbitGelfUserStr+":"+rabbitGelfPasswordStr+"@"+rabbitGelfHostStr+":5671/", tlsConfig)
	} else {
		conn, err = amqp.Dial("amqp://" + rabbitGelfUserStr + ":" + rabbitGelfPasswordStr + "@" + rabbitGelfHostStr + ":5672/")
	}

	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	// defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	// defer ch.Close()

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

		fmt.Println("error sent to Rabbit")

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
					true,           // durable
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

				errorToLogJSON, errJSON := json.Marshal(errorToLog)
				if errJSON == nil {
					err = logger.channel.Publish(
						"",     // exchange
						q.Name, // routing key
						false,  // mandatory
						false,  // immediate
						amqp.Publishing{
							ContentType: "text/plain",
							Body:        []byte(errorToLogJSON),
						})
					failOnError(err, "Failed to publish a message")
				}

			}
		}

	}

}
