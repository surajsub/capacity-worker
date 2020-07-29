package utils


import "time"

type VMCSV struct {
	Hostname           string  `json:"hostname"`
	Site               string  `json:"site"`
	StoragePoolName    string  `json:"storagePoolName"`
	Profile            string  `json:"Profile"`
	Cdir               string  `json:"cdir"`
	ResourceStatus     string  `json:"resourceStatus"`
	WorkloadType       string  `json:"workloadType"`
	RequestID          string  `json:"requestId"`
	Memory             int     `json:"memory"`
	CPU                float32 `json:"cpu"`
	Storage            int     `json:"storage"`
	VcenterCPU         int     `json:"vcetnerCPU"`
	MEMORYREQUESTEDGB  int     `json:"MEMORY_REQUESTED_GB"`
	STORAGEREQUESTEDGB int     `json:"STORAGEREQUESTEDGB"`
}

func (vmCSV VMCSV) getKeys() []string {
	return []string{
		"Hostname",
		"Site",
		"Profile",
		"Cdir",
		"ResourceStatus",
		"WorkloadType",
		"RequestID",
		"StoragePoolName",
		"Memory",
		"MEMORYREQUESTEDGB",
		"CPU",
		"VcenterCPU",
		"Storage",
		"STORAGEREQUESTEDGB",
		"Timestamp",
	}
}

func (vmCSV VMCSV) getValues() []string {
	return []string{
		vmCSV.Hostname,
		vmCSV.Site,
		vmCSV.Profile,
		vmCSV.Cdir,
		vmCSV.ResourceStatus,
		vmCSV.WorkloadType,
		vmCSV.RequestID,
		vmCSV.StoragePoolName,
		customItoa(vmCSV.Memory),
		customItoa(vmCSV.MEMORYREQUESTEDGB),
		customFloat32ToAsci(vmCSV.CPU),
		customItoa(vmCSV.VcenterCPU),
		customItoa(vmCSV.Storage),
		customItoa(vmCSV.STORAGEREQUESTEDGB),
		time.Now().String(),
	}
}
