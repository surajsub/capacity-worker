package events

import (
	"errors"
	"strconv"
        "github.com/sirupsen/logrus"
        "github.com/opaas/capacity-worker/client"
        "github.com/opaas/capacity-worker/utils"
)

type ClusterHostEvent struct {
	StreamName    string         `json:"streamName"`
        Data          []ClusterHost  `json:"data"`
}

type ClusterHost struct {
	STREAMNAME    string `json:"STREAM_NAME"`
	HOSTNAME      string `json:"HOSTNAME"`
	POD           string `json:"PODID"`
	CLUSTERNAME   string `json:"ESXNAME"`
	DATACENTER    string `json:"DATACENTER"`
}

func (event ClusterHostEvent) Process(offset int64, opaasData *opaas.OpaasData, SlData []internal.SoftLayerHosts) {
	for _, clusterhost := range event.Data {
		processClusterhost(clusterhost, opaasData, SlData)
	}
}

func processClusterhost(clusterhost ClusterHost, opaasData *opaas.OpaasData, SlData []internal.SoftLayerHosts) {
	cluster := findCluster(clusterhost, opaasData)
	if cluster == nil {
		logFields := logrus.Fields{
			"clusterHost": clusterhost.HOSTNAME,
			"ClusterName": clusterhost.CLUSTERNAME,
			"Datacenter":  clusterhost.DATACENTER,
			"Pod":         clusterhost.POD,
		}
		logrus.WithFields(logFields).Info("Cannot find matching clustername in Opaas")
		return
	}
	if cluster.Profile != "3x" {
		logFields := logrus.Fields{
			"clusterHost": clusterhost.HOSTNAME,
			"Profile":     cluster.Profile,
		}
		logrus.WithFields(logFields).Info("Clusterhost is not 3x")
		return
	}
	serverIdErr, serverId := findServerID(clusterhost, SlData)
	if serverIdErr != nil {
		logFields := logrus.Fields{
			"clusterHost": clusterhost.HOSTNAME,
		}
		logrus.WithFields(logFields).Info("Cannot find matching hostname and serverID from SoftLayer")
		return 
	}
	clusterHost := findClusterhost(clusterhost, opaasData)
	if clusterHost == nil {
		addNewClusterhost(clusterhost, cluster, serverId)
	} else {
		if clusterHost.ServerID != serverId {
			sendNewServerIDSlackMessage(clusterHost, cluster, serverId)
		}
	}
}

func findClusterhost(clusterhost ClusterHost, opaasData *opaas.OpaasData) *opaas.Clusterhost {
	for _, hostrecord := range opaasData.Clusterhosts {
		if hostrecord.Name == clusterhost.HOSTNAME {
			return &hostrecord
		}
	}
	return nil
}

func addNewClusterhost(clusterhost ClusterHost, cluster *opaas.Cluster, serverID string) {
// code to build new cluster-host payload and send to opaas goes here
	sendAddClusterHostSlackMessage(cluster, clusterhost, serverID)
}

func findCluster(clusterhost ClusterHost, opaasData *opaas.OpaasData) *opaas.Cluster {
	for _, cluster := range opaasData.Clusters {
		if cluster.ClusterName == clusterhost.CLUSTERNAME {
			if cluster.Datacenter == clusterhost.DATACENTER {
				if strconv.Itoa(cluster.Pod) == clusterhost.POD {
					return &cluster
				}
			}
		}
	}
	return nil
}

func findServerID(clusterhost ClusterHost, SlData []internal.SoftLayerHosts) (error, string) {
	var err error = nil
	for _, slHost := range SlData {
		if slHost.FQDN == clusterhost.HOSTNAME {
			return err, strconv.Itoa(slHost.ID)
		}
	}
	err = errors.New("No match in Slack Data for clusterhost.HOSTNAME")
	return err, clusterhost.HOSTNAME
}

func sendNewServerIDSlackMessage(clusterHost *opaas.Clusterhost, cluster *opaas.Cluster, serverId string) {
	slackCHParams := &internal.SlackCHParams{
		Hostname:    clusterHost.Name,
		ServerId:    serverId,
		OldServerId: clusterHost.ServerID,
		Profile:     cluster.Profile,
	}
	internal.SendNewServerIdSlackMessage(slackCHParams)
}

func sendAddClusterHostSlackMessage(cluster *opaas.Cluster, clusterhost ClusterHost, serverID string) {
	slackCHParams := &internal.SlackCHParams{
		Hostname:      clusterhost.HOSTNAME,
		ServerId:      serverID,
		WorkloadTypes: cluster.WorkloadTypes,
		ClusterId:     cluster.ID,
		Profile:       cluster.Profile,
	}
	internal.SendAddClusterHostSlackMessage(slackCHParams)
}
