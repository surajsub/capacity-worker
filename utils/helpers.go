package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	slackTokenEnv     string = "CAPACITY_SLACK_TOKEN"
	slackChannelIdEnv string = "CAPACITY_SLACK_CHANNEL"
	opaasUrlEnv       string = "OPAAS_BASE_URL"
	opaasKeyEnv       string = "OPAAS_APIKEY"
	kafkaUsernameEnv  string = "CAP_KAFKA_USER"
	kafkaPasswordEnv  string = "CAP_KAFKA_PASSWORD"
	kafkaTopicEnv     string = "CAP_KAFKA_TOPIC"
	kafkaBrokersEnv   string = "CAP_KAFKA_BROKERS"
)

type GlobalHook struct {
}

type KafkaConfig struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Topic    string   `json:"topic"`
	Brokers  []string `json:"brokers"`
}

type SlackConfig struct {
	ChannelID string `json:"channelId"`
	Token     string `json:"token"`
}

func GetKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		Username: viper.GetString(kafkaUsernameEnv),
		Password: viper.GetString(kafkaPasswordEnv),
		Topic:    viper.GetString(kafkaTopicEnv),
		Brokers:  viper.GetStringSlice(kafkaBrokersEnv),
	}
}

func GetSlackConfig() *SlackConfig {
	return &SlackConfig{
		Token:     viper.GetString(slackTokenEnv),
		ChannelID: viper.GetString(slackChannelIdEnv),
	}
}

func InitEnv() error {
	if goDotErr := godotenv.Load(); goDotErr != nil {
		return goDotErr
	}

	requiredEnvVars := []string{
		slackTokenEnv,
		slackChannelIdEnv,
		opaasUrlEnv,
		opaasKeyEnv,
		kafkaUsernameEnv,
		kafkaPasswordEnv,
		kafkaTopicEnv,
		kafkaBrokersEnv,
	}

	for _, envVar := range requiredEnvVars {
		bindErr := viper.BindEnv(envVar)
		if bindErr != nil {
			return bindErr
		}
		if !viper.IsSet(envVar) {
			errMsg := fmt.Sprintf("%s env variable is not set", envVar)
			return errors.New(errMsg)
		}
	}

	return nil
}

func InitLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.AddHook(&GlobalHook{})
	logrus.SetOutput(os.Stdout)
}

func (h *GlobalHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *GlobalHook) Fire(e *logrus.Entry) error {
	hostname, errHostname := os.Hostname()

	if errHostname != nil {
		return errHostname
	}
	e.Data["hostname"] = hostname
	e.Data["pid"] = os.Getpid()
	e.Data["v"] = 1
	return nil
}
