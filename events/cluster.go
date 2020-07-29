package events

import (
	"strconv"

	"github.com/opaas/capacity-worker/client"
	"github.com/opaas/capacity-worker/utils"
	"github.com/sirupsen/logrus"
)

type Cluster struct {
	SiteID                 string  `json:"SITE_ID"`
	Pod                    string  `json:"PODID"`
	Datacenter             string  `json:"DATACENTER"`
	EsxName                string  `json:"ESXNAME"`
	PoolName               string  `json:"POOL_NAME"`
	CPUTotal               int     `json:"VCPU_TOTAL"`
	CPURequested           int     `json:"VCPU_REQUESTED"`
	MemoryTotal            int     `json:"MEMORY_TOTAL_GB"`
	MemoryRequested        int     `json:"MEMORY_REQUESTED_GB"`
	CPURequestedPercent    float32 `json:"VCPU_TOTAL_REQUESTED_PCT"`
	CPUAvailablePercent    float32 `json:"VCPU_TOTAL_AVAILABLE_PCT"`
	MemoryRequestedPercent float32 `json:"MEMORY_TOTAL_REQUESTED_PCT"`
	MemoryAvailablePercent float32 `json:"MEMORY_TOTAL_AVAILABLE_PCT"`
	Version                string  `json:"VERSION"`
}

type ClusterEvent struct {
	StreamName string    `json:"streamName"`
	Clusters   []Cluster `json:"data"`
}

func (event ClusterEvent) Process(offset int64, opaasData *client.OpaasData, SlData []utils.SoftLayerHosts) {
	if event.StreamName == "xseries.resource_pool" {
		processResourcePools(offset, event, opaasData)
		return
	}
	processClusters(offset, event, opaasData)
}

func processResourcePools(offset int64, event ClusterEvent, opaasData *client.OpaasData) {
	for _, resourcePool := range event.Clusters {
		if is3x(resourcePool) {
			processResourcePool(offset, resourcePool, opaasData)
		}
	}
}

func processClusters(offset int64, event ClusterEvent, opaasData *client.OpaasData) {
	for _, cluster := range event.Clusters {
		if !is3x(cluster) {
			processCluster(offset, cluster, opaasData)
		}
	}
}

func is3x(cluster Cluster) bool {
	return cluster.Version == "CMS 3.x"
}

func processResourcePool(offset int64, resourcePool Cluster, opaasData *client.OpaasData) {
	resourcePool.SiteID = mapSites(resourcePool.SiteID)
	opaasCluster := findMatchingOpaasClusterWithResourcePool(resourcePool, opaasData)
	if opaasCluster != nil {
		// sendClusterSlackMessage(resourcePool, opaasCluster.Profile)
		patchClusterIfNecessary(resourcePool, opaasCluster)
	}
}

func processCluster(offset int64, cluster Cluster, opaasData *client.OpaasData) {
	cluster.SiteID = mapSites(cluster.SiteID)
	opaasCluster := findMatchingOpaasClusterWithCluster(cluster, opaasData.Clusters)
	if opaasCluster != nil {
		// sendClusterSlackMessage(cluster, opaasCluster.Profile)
		patchClusterIfNecessary(cluster, opaasCluster)
	}
}

func findMatchingOpaasClusterWithResourcePool(resourcePool Cluster, opaasData *client.OpaasData) *client.Cluster {
	logFields := logrus.Fields{
		"site":             resourcePool.SiteID,
		"pod":              resourcePool.Pod,
		"datacenter":       resourcePool.Datacenter,
		"resourcePoolName": resourcePool.PoolName,
	}
	logrus.WithFields(logFields).Info("Searching for cluster in opaas that matches vcenter resource pool")
	for _, opaasCluster := range opaasData.Clusters {
		if clusterMatchesResourcePool(resourcePool, opaasCluster) {
			logrus.WithFields(logFields).Info("Found cluster in opaas that matches vcenter resource pool")
			return &opaasCluster
		}
	}
	logrus.WithFields(logFields).Info("Unable to find cluster in opaas that matches vcenter resource pool")
	return nil
}

func findMatchingOpaasClusterWithCluster(cluster Cluster, opaasClusters []client.Cluster) *client.Cluster {
	logFields := logrus.Fields{
		"site":        cluster.SiteID,
		"datacenter":  cluster.Datacenter,
		"clusterName": cluster.EsxName,
	}
	logrus.WithFields(logFields).Info("Searching for cluster in opaas that matches vcenter cluster")
	for _, opaasCluster := range opaasClusters {
		if clusterMatchesCluster(opaasCluster, cluster) {
			logrus.WithFields(logFields).Info("Found cluster in opaas that matches vcenter cluster")
			return &opaasCluster
		}
	}
	logrus.WithFields(logFields).Info("Unable to find cluster in opaas that matches vcenter cluster")
	return nil
}

func clusterMatchesResourcePool(resourcePool Cluster, opaasCluster client.Cluster) bool {
	clusterPodInt, atoiErr := strconv.Atoi(resourcePool.Pod)
	if atoiErr != nil {
		logrus.WithFields(logrus.Fields{
			"pod":   resourcePool.Pod,
			"Error": atoiErr.Error(),
		}).Error("Unable to convert ascii pod value to integer")
		return false
	}
	return opaasCluster.PoolLocation == resourcePool.SiteID &&
		opaasCluster.Pod == clusterPodInt &&
		opaasCluster.Datacenter == resourcePool.Datacenter &&
		opaasCluster.ResourcePoolName == resourcePool.PoolName
}

func clusterMatchesCluster(opaasCluster client.Cluster, cluster Cluster) bool {
	return opaasCluster.PoolLocation == cluster.SiteID &&
		opaasCluster.Datacenter == cluster.Datacenter &&
		opaasCluster.ClusterName == cluster.EsxName
}

func sendClusterSlackMessage(cluster Cluster, profile string) {
	slackParams := &utils.SlackParams{
		EsxName:                cluster.EsxName,
		PoolName:               cluster.PoolName,
		Pod:                    cluster.Pod,
		Datacenter:             cluster.Datacenter,
		Site:                   cluster.SiteID,
		CPURequestedPercent:    cluster.CPURequestedPercent,
		CPUAvailablePercent:    cluster.CPUAvailablePercent,
		MemoryRequestedPercent: cluster.MemoryRequestedPercent,
		MemoryAvailablePercent: cluster.MemoryAvailablePercent,
		Profile:                profile,
	}
	logrus.WithFields(logrus.Fields{
		"slackParams": slackParams,
	}).Info("Sending message to slack for cluster")
	utils.SendSlackMessage(slackParams)
}

func patchClusterIfNecessary(cluster Cluster, opaasCluster *client.Cluster) {
	patches := createNecessaryClusterPatches(cluster, opaasCluster)
	logFields := logrus.Fields{
		"patches":          patches,
		"clusterId":        opaasCluster.ID,
		"resourcePoolName": cluster.PoolName,
	}
	if len(patches) == 0 {
		logrus.WithFields(logFields).Info("Cluster is up to date with vcenter")
		return
	}
	logrus.WithFields(logFields).Info("Patching cluster")
	patchCluster(opaasCluster.ID, patches)
}

func createNecessaryClusterPatches(cluster Cluster, opaasCluster *client.Cluster) []client.Patch {
	patches := []client.Patch{}
	if cpuPatchIsNecessary(cluster, opaasCluster) {
		patches = append(patches, createCPUPatch(cluster))
	}
	if memoryPatchIsNecessary(cluster, opaasCluster) {
		patches = append(patches, createMemoryPatch(cluster))
	}
	return patches
}

func cpuPatchIsNecessary(cluster Cluster, opaasCluster *client.Cluster) bool {
	return opaasCluster.CPUInUseByOpaas != cluster.CPURequested &&
		opaasCluster.VCenterCPUConsumed != cluster.CPURequested
}

func memoryPatchIsNecessary(cluster Cluster, opaasCluster *client.Cluster) bool {
	return opaasCluster.MemoryInUseByOpaas != cluster.MemoryRequested &&
		opaasCluster.VCenterMemoryConsumed != cluster.MemoryRequested
}

func createCPUPatch(cluster Cluster) client.Patch {
	return client.Patch{
		Op:    "replace",
		Path:  "/vCenterCpuConsumed",
		Value: cluster.CPURequested,
	}
}

func createMemoryPatch(cluster Cluster) client.Patch {
	return client.Patch{
		Op:    "replace",
		Path:  "/vCenterMemoryConsumed",
		Value: cluster.MemoryRequested,
	}
}

func patchCluster(clusterID string, patches []client.Patch) {
	opaasAPI := client.NewOpaasApi()
	clusterPatchErr := opaasAPI.PatchCluster(clusterID, patches)
	if clusterPatchErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": clusterPatchErr.Error(),
		}).Error("Failed to patch cluster")
	} else {
		logrus.WithFields(logrus.Fields{
			"clusterID": clusterID,
		}).Info("Successfully patched cluster")
	}
}
