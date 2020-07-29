package client

const cluster_endpoint string = "clusters"

type Cluster struct {
	ID                    string   `json:"id"`
	StorageIds            []string `json:"storageIds"`
	ResourcePoolName      string   `json:"resourcePoolName"`
	VMCount               int      `json:"vmCount"`
	PoolLocation          string   `json:"poolLocation"`
	Pod                   int      `json:"pod"`
	Datacenter            string   `json:"dataCenterName"`
	CPUInUseByOpaas       int      `json:"cpuInUseByOpaas"`
	VCenterCPUConsumed    int      `json:"vCenterCPUConsumed"`
	MemoryInUseByOpaas    int      `json:"memoryInUseByOpaas"`
	VCenterMemoryConsumed int      `json:"vCenterMemoryConsumed"`
	Profile               string   `json:"profile"`
	ClusterName           string   `json:"clusterName"`
	WorkloadTypes         []string `json:"workloadTypes"`
}

func (opaasApi *OpaasApi) GetClusters() ([]Cluster, error) {
	clusterData := []Cluster{}
	httpErr := opaasApi.get(cluster_endpoint, &clusterData)
	if httpErr != nil {
		return nil, httpErr
	}
	return clusterData, nil
}

func (opaasApi *OpaasApi) PatchCluster(clusterId string, patches []Patch) error {
	return opaasApi.patch(cluster_endpoint, clusterId, patches)
}
