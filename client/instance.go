package client

const instances_endpoint string = "instances"

type instanceStorage struct {
	Usage    string `json:"usage"`
	Size     int    `json:"size"`
	PoolName string `json:"poolName"`
}

type Instance struct {
	Hostname       string            `json:"hostname"`
	Site           string            `json:"site"`
	Profile        string            `json:"profile"`
	Cdir           string            `json:"cdir"`
	ResourceStatus string            `json:"resourceStatus"`
	WorkloadType   string            `json:"workloadType"`
	RequestID      string            `json:"requestId"`
	Memory         int               `json:"memory"`
	CPU            float32           `json:"cpu"`
	Storage        []instanceStorage `json:"storage"`
}

func (opaasApi *OpaasApi) GetInstances() ([]Instance, error) {
	instanceData := []Instance{}
	httpErr := opaasApi.get(instances_endpoint, &instanceData)
	if httpErr != nil {
		return nil, httpErr
	}
	return instanceData, nil
}
