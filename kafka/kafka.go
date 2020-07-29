package kafka

import (

	"crypto/tls"
	"github.com/opaas/capacity-worker/utils"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/sirupsen/logrus"
	"github.com/opaas/capacity-worker"

)

type KafkaEvent struct {
	StreamName string `json:"streamName"`
}

func NewKafkaReader() *kafkaGo.Reader {
	kafkaReaderConfig := createKafkaReaderConfig()
	kafkaReader := kafkaGo.NewReader(kafkaReaderConfig)
	currentOffset := getCurrentOffset()
	kafkaReader.SetOffset(currentOffset)
	return kafkaReader
}

func createKafkaReaderConfig() kafkaGo.ReaderConfig {
	kafkaConfig := utils.GetKafkaConfig()
	kafkaDialer := createKafkaDialer(kafkaConfig)
	return kafkaGo.ReaderConfig{
		Brokers: kafkaConfig.Brokers,
		Topic:   kafkaConfig.Topic,
		MaxWait: 500 * time.Millisecond,
		Dialer:  kafkaDialer,
	}
}

func createKafkaDialer(kafkaConfig *utils.KafkaConfig) *kafkaGo.Dialer {
	return &kafkaGo.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		SASLMechanism: plain.Mechanism{
			Username: kafkaConfig.Username,
			Password: kafkaConfig.Password,
		},
		TLS: &tls.Config{},
	}
}

func getCurrentOffset() int64 {
	lastOffsetRecorded, offsetErr := ioUtils.ReadOffset()
	if offsetErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": offsetErr.Error(),
		}).Fatal("Unable to retrieve offset")
	}
	currentOffset := lastOffsetRecorded + 1
	return currentOffset
}

func CloseKafkaReader(kafkaReader *kafkaGo.Reader) {
	readerCloseErr := kafkaReader.Close()
	if readerCloseErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": readerCloseErr.Error(),
		}).Info("Error closing consumer")
	}
}
