package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/opaas/capacity-worker/client"
	"github.com/opaas/capacity-worker/events"
	"github.com/opaas/capacity-worker/kafka"
	"github.com/opaas/capacity-worker/utils"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

const (
	batch_size           int           = 300
	message_read_timeout time.Duration = 10 * time.Second
)

func init() {
	utils.InitLogger()
	envVarErr := utils.InitEnv()
	if envVarErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": envVarErr,
		}).Fatal()
	}
}

func main() {
	for {
		messageBatch := readMessageBatchOrTimeout()
		processMessageBatch(messageBatch)
	}
}

func readMessageBatchOrTimeout() []kafkaGo.Message {
	logrus.Info("Reading messages from kafka with timeout value of ", message_read_timeout.String())
	kafkaReader := kafka.NewKafkaReader()
	//defer kafkaReader.Close()
	defer kafka.CloseKafkaReader(kafkaReader)
	contextWithTimeout, cancelTimeout := context.WithTimeout(context.Background(), message_read_timeout)
	defer cancelTimeout()
	var messageBatch []kafkaGo.Message
	for i := 0; i < batch_size; i++ {
		message, readMessageErr := kafkaReader.ReadMessage(contextWithTimeout)
		//message, readMessageErr := kafkaReader.ReadMessage(context.Background())
		if readMessageErr != nil {
			logrus.WithFields(logrus.Fields{
				"Error": readMessageErr.Error(),
			}).Info()
			break
		}
		messageBatch = append(messageBatch, message)
	}
	return messageBatch
}

func processMessageBatch(messageBatch []kafkaGo.Message) {
	if len(messageBatch) == 0 {
		logrus.Info("No new messages read from kafka")
		return
	}
	logrus.WithFields(logrus.Fields{
		"messageBatchLength": len(messageBatch),
	}).Info("Processing message batch")
	batchOpaasData := getOpaasData()
	SlData := utils.GetSLData()
	for _, message := range messageBatch {
		logrus.Info("Printing the message %s ", message.Topic)
		processMessage(message, batchOpaasData, SlData)
		offsetErr := utils.WriteOffset(message.Offset)
		if offsetErr != nil {
			logrus.Fatal("Unable to write offset to file")
		}
	}
}

func getOpaasData() *client.OpaasData {
	opaasAPI := client.NewOpaasApi()
	instances, instErr := opaasAPI.GetInstances()
	if instErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": instErr.Error(),
		}).Fatal("Unable to retrieve instances")
	}
	storage, storageErr := opaasAPI.GetStorage()
	if storageErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": storageErr.Error(),
		}).Fatal("Unable to retrieve storage")
	}
	clusters, clustErr := opaasAPI.GetClusters()
	if clustErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": clustErr.Error(),
		}).Fatal("Unable to retrieve clusters")
	}
	clusterhosts, clusterhostErr := opaasAPI.GetClusterhosts()
	if clusterhostErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": clusterhostErr.Error(),
		}).Fatal("Unable to retrieve clusterhosts")
	}
	return &client.OpaasData{
		Instances:    instances,
		Storage:      storage,
		Clusters:     clusters,
		Clusterhosts: clusterhosts,
	}
}

func processMessage(message kafkaGo.Message, batchOpaasData *client.OpaasData, SlData []utils.SoftLayerHosts) {
	offset := message.Offset
	event, conversionErr := convertMessageToEvent(message)
	if conversionErr != nil {
		logrus.WithFields(logrus.Fields{
			"offset": offset,
			"Error":  conversionErr.Error(),
		}).Error("Problem occured while converting kafka message to usable event")
		return
	}
	logrus.WithFields(logrus.Fields{
		"offset": offset,
	}).Info("Successfully converted message to event. Beginning to process event")
	event.Process(offset, batchOpaasData, SlData)
}

func convertMessageToEvent(message kafkaGo.Message) (events.Event, error) {
	unknownEvent := kafka.KafkaEvent{}
	unknownUnmarshallErr := json.Unmarshal(message.Value, &unknownEvent)
	if unknownUnmarshallErr != nil {
		return nil, unknownUnmarshallErr
	}
	event, findCorrectEventTypeErr := findCorrectEventType(unknownEvent)
	if findCorrectEventTypeErr != nil {
		return nil, findCorrectEventTypeErr
	}
	eventUnmarshalErr := json.Unmarshal(message.Value, event)
	return event, eventUnmarshalErr
}

func findCorrectEventType(unknownEvent kafka.KafkaEvent) (events.Event, error) {
	var event events.Event = nil
	var err error = nil
	switch unknownEvent.StreamName {
	case "xseries.datastore":
		event = &events.DatastoreEvent{}
	case "xseries.vminfo":
		event = &events.VMEvent{}
	case "xseries.esx_cluster":
		event = &events.ClusterEvent{}
	case "xseries.resource_pool":
		event = &events.ClusterEvent{}
	case "xseries.esx_host":
		event = &events.ClusterHostEvent{}
	default:
		err = errors.New("No operation defined for provided StreamName")
	}
	return event, err
}
