package wconnectors

import (
	"log"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

var kafkaWriterConnections map[string]*kafka.Writer
var kafkaWriterOnce map[string]bool
var kafkaWriterOnceMutex sync.Mutex

// KafkaWriterSettings is the struct that is used for registering a connection
type KafkaWriterSettings struct {
	Topic   string
	Brokers []string
}

var allKafkaWriterSettings = make(map[string]KafkaWriterSettings)

// RegisterKafkaWriter registers a db connection
func RegisterKafkaWriter(topicName string, settings KafkaWriterSettings) {
	allKafkaWriterSettings[topicName] = settings
}

// KafkaWriter return the writer to the topic name
func KafkaWriter(topicName string) *kafka.Writer {

	if topicName == "" {
		log.Println("KafkaWriter's topicName cannot be empty")
		return nil
	}

	if len(kafkaWriterOnce) == 0 {
		kafkaWriterOnce = make(map[string]bool, 15)
		kafkaWriterConnections = make(map[string]*kafka.Writer, 15)
	}

	kafkaWriterOnceMutex.Lock()
	if !kafkaWriterOnce[topicName] {
		kafkaWriterOnce[topicName] = true

		kkConnection := kafka.NewWriter(kafka.WriterConfig{
			Brokers:      allKafkaWriterSettings[topicName].Brokers,
			Topic:        allKafkaWriterSettings[topicName].Topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
		})
		kafkaWriterConnections[topicName] = kkConnection
		kafkaWriterOnceMutex.Unlock()
	} else {
		kafkaWriterOnceMutex.Unlock()
	}
	return kafkaWriterConnections[topicName]
}
