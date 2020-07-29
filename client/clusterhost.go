package client

const clusterhost_endpoint string = "cluster-hosts"

type Clusterhost struct {
	ID            string   `json:"id"`
	Name          string   `json:"hostName"`
	ServerID      string   `json:"serverId"`
	ClusterID     string   `json:"clusterId"`
	WorkloadTypes []string `json:"workloadTypes"`
}

func (opaasApi *OpaasApi) GetClusterhosts() ([]Clusterhost, error) {
	clusterhostData := []Clusterhost{}
	httpErr := opaasApi.get(clusterhost_endpoint, &clusterhostData)
	if httpErr != nil {
		return nil, httpErr
	}
	return clusterhostData, nil
}

