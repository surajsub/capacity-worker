package utils

import (
	"fmt"
	"github.com/slack-go/slack"
)

const (
	newClusterHost string = "Attention! New Clusterhost Created"
	changedServerId string = "Attention! Clusterhost Has a new ServerID"
)

var (
	profileToEmoji = make(map[string]string)
)

func init() {
	profileToEmoji["sei"] = ":vmw2:"
	profileToEmoji["seix"] = ":lizard:"
	profileToEmoji["3x"] = ":vmware:"
	profileToEmoji["uma"] = ":floppy_disk:"
}

type SlackParams struct {
	EsxName                string  `json:"esxName"`
	PoolName               string  `json:"poolName"`
	Pod                    string  `json:"pod"`
	Datacenter             string  `json:"datacenter"`
	Site                   string  `json:"site"`
	CPURequestedPercent    float32 `json:"cpuRequestedPercent"`
	CPUAvailablePercent    float32 `json:"cpuAvailablePercent"`
	MemoryRequestedPercent float32 `json:"memoryRequestedPercent"`
	MemoryAvailablePercent float32 `json:"memoryAvailablePercent"`
	Profile                string  `json:"profile"`
}

type SlackCHParams struct {
	Hostname      string   `json:"hostName"`
	ServerId      string   `json:"serverId"`
	OldServerId   string   `json:"oldServerId"`
	WorkloadTypes []string `json:"workloadTypes"`
	ClusterId     string   `json:"clusterId"`
	Profile       string   `json:"profile"`
}

// TODO: Add logrus logging and error handling
func SendSlackMessage(slackParams *SlackParams) {
	blocks := constructSlackBlocks(slackParams)
	slackConfig := GetSlackConfig()
	api := slack.New(slackConfig.Token)
	_, _, err := api.PostMessage(slackConfig.ChannelID, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}

func SendAddClusterHostSlackMessage(slackCHParams *SlackCHParams) {
	blocks := constructCHSlackBlocks(slackCHParams, newClusterHost)
	slackConfig := GetSlackConfig()
	api := slack.New(slackConfig.Token)
	_, _, err := api.PostMessage(slackConfig.ChannelID, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}

func SendNewServerIdSlackMessage(slackCHParams *SlackCHParams) {
	blocks := constructCHSlackBlocks(slackCHParams, changedServerId)
	slackConfig := GetSlackConfig()
	api := slack.New(slackConfig.Token)
	_, _, err := api.PostMessage(slackConfig.ChannelID, slack.MsgOptionBlocks(blocks...))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}


func constructSlackBlocks(slackParams *SlackParams) []slack.Block {
	headerBlock := constructHeaderBlock(slackParams)
	locationBlock := constructLocationBlock(slackParams)
	fieldsBlock := constructFieldsBlock(slackParams)
	dividerBlock := slack.NewDividerBlock()
	return []slack.Block{
		headerBlock,
		locationBlock,
		fieldsBlock,
		dividerBlock,
	}
}

func constructHeaderBlock(slackParams *SlackParams) *slack.SectionBlock {
	emoji := profileToEmoji[slackParams.Profile]
	headerText := fmt.Sprintf("[%s *%s*] *Cluster Name: %s* *Resource Pool Name: %s*", emoji, slackParams.Profile, slackParams.EsxName, slackParams.PoolName)
	headerTextBlockObj := slack.NewTextBlockObject("mrkdwn", headerText, false, false)
	return slack.NewSectionBlock(headerTextBlockObj, nil, nil)
}

func constructLocationBlock(slackParams *SlackParams) *slack.SectionBlock {
	locationText := fmt.Sprintf("*Site:  %s Datacenter: %s Pod: %s*", slackParams.Site, slackParams.Datacenter, slackParams.Pod)
	locationTextBlockObj := slack.NewTextBlockObject("mrkdwn", locationText, false, false)
	return slack.NewSectionBlock(locationTextBlockObj, nil, nil)
}

func constructFieldsBlock(slackParams *SlackParams) *slack.SectionBlock {
	cpuRequestedField := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*CPU Percent Requested: %0.2f%%*", slackParams.CPURequestedPercent), false, false)
	cpuAvailableField := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*CPU Percent Available: %0.2f%%*", slackParams.CPUAvailablePercent), false, false)
	memoryRequestedField := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Memory Percent Requested: %0.2f%%*", slackParams.MemoryRequestedPercent), false, false)
	memoryAvailableField := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*Memory Percent Available: %0.2f%%*", slackParams.MemoryAvailablePercent), false, false)
	fields := []*slack.TextBlockObject{
		cpuRequestedField,
		cpuAvailableField,
		memoryRequestedField,
		memoryAvailableField,
	}
	return slack.NewSectionBlock(nil, fields, nil)
}

func constructCHSlackBlocks(slackCHParams *SlackCHParams, title string) []slack.Block {
	if title == newClusterHost {
		ncHeaderBlock := constructCHHeaderBlock(slackCHParams, newClusterHost)
		ncHostnameBlock := constructCHNameBlock(slackCHParams)
		ncFieldsBlock := constructCHFieldsBlock(slackCHParams, newClusterHost)
		ncWorkloadBlock := constructCHWorkloadBlock(slackCHParams)
		ncDividerBlock := slack.NewDividerBlock()
		return []slack.Block{
			ncHeaderBlock,
			ncHostnameBlock,
			ncFieldsBlock,
			ncWorkloadBlock,
			ncDividerBlock,
		}
	} else {
		headerBlock := constructCHHeaderBlock(slackCHParams, changedServerId)
		hostnameBlock := constructCHNameBlock(slackCHParams)
		chFieldsBlock := constructCHFieldsBlock(slackCHParams, changedServerId)
		dividerBlock := slack.NewDividerBlock()
		return []slack.Block{
			headerBlock,
			hostnameBlock,
			chFieldsBlock,
			dividerBlock,
		}
	}
}

func constructCHHeaderBlock(slackCHParams *SlackCHParams, title string) *slack.SectionBlock {
	emoji := profileToEmoji[slackCHParams.Profile]
//	headerText := fmt.Sprintf("[%s *%s*] <!here> *%s*", emoji, slackCHParams.Profile, title)
	headerText := fmt.Sprintf("[%s *%s*] *%s*", emoji, slackCHParams.Profile, title)
	headerTextBlockObj := slack.NewTextBlockObject("mrkdwn", headerText, false, false)
	return slack.NewSectionBlock(headerTextBlockObj, nil, nil)
}

func constructCHNameBlock(slackCHParams *SlackCHParams) *slack.SectionBlock {
	nameText := fmt.Sprintf("*Clusterhost Name: %s*", slackCHParams.Hostname)
	nameTextBlockObj := slack.NewTextBlockObject("mrkdwn", nameText, false, false)
	return slack.NewSectionBlock(nameTextBlockObj, nil, nil)
}

func constructCHFieldsBlock(slackCHParams *SlackCHParams, title string) *slack.SectionBlock {
	if title == newClusterHost {
		ncFieldsText := fmt.Sprintf("*ServerID:  %s ClusterID: %s*", slackCHParams.ServerId, slackCHParams.ClusterId)
		ncFieldsTextBlockObj := slack.NewTextBlockObject("mrkdwn", ncFieldsText, false, false)
		return slack.NewSectionBlock(ncFieldsTextBlockObj, nil, nil)
	} else {
		csFieldsText := fmt.Sprintf("*New ServerID:  %s    Old ServerID: %s*", slackCHParams.ServerId, slackCHParams.OldServerId)
		csFieldsTextBlockObj := slack.NewTextBlockObject("mrkdwn", csFieldsText, false, false)
		return slack.NewSectionBlock(csFieldsTextBlockObj, nil, nil)
	}
}

func constructCHWorkloadBlock(slackCHParams *SlackCHParams) *slack.SectionBlock {
	workloadText := fmt.Sprintf("*WorkLoad Types: %s*", slackCHParams.WorkloadTypes)
	workloadTextBlockObj := slack.NewTextBlockObject("mrkdwn", workloadText, false, false)
	return slack.NewSectionBlock(workloadTextBlockObj, nil, nil)
}
