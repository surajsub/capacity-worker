package events

import (
	"github.com/opaas/capacity-worker/client"
	"github.com/opaas/capacity-worker/utils"
	"github.com/sirupsen/logrus"
)

const instance_csv_file string = "output/Instances.csv"

type VM struct {
	STREAMNAME         string `json:"STREAM_NAME"`
	DOCID              string `json:"DOCID"`
	TS                 string `json:"TS"`
	SNAPSHOTID         int    `json:"SNAPSHOT_ID"`
	VMName             string `json:"VM_NAME"`
	CPU                int    `json:"CPU"`
	MEMORYREQUESTEDGB  int    `json:"MEMORY_REQUESTED_GB"`
	STORAGEREQUESTEDGB int    `json:"STORAGE_REQUESTED_GB"`
	PODID              string `json:"PODID"`
	SITEID             string `json:"SITE_ID"`
	DATACENTER         string `json:"DATACENTER"`
}

type VMEvent struct {
	StreamName string `json:"streamName"`
	VMs        []VM   `json:"data"`
}

func (event VMEvent) Process(offset int64, opaasData *client.OpaasData, SlData []utils.SoftLayerHosts) {
	vmCSVs := []utils.CSVInfo{}
	for _, vm := range event.VMs {
		vmCSV := processVM(vm, opaasData)
		vmCSVs = append(vmCSVs, vmCSV)
	}
	writeVMCSV(offset, vmCSVs)
}

func processVM(vm VM, opaasData *client.OpaasData) *utils.VMCSV {
	vm.SITEID = mapSites(vm.SITEID)
	vmCSV := createVMCSV(vm)
	opaasInstance := findMatchingVM(vm, opaasData)
	if opaasInstance != nil {
		addOpaasVMCSVInfo(opaasInstance, vmCSV)
	}
	return vmCSV
}

func findMatchingVM(vm VM, opaasData *client.OpaasData) *client.Instance {
	logFields := logrus.Fields{
		"hostname": vm.VMName,
	}
	logrus.WithFields(logFields).Info("Searching for matching vm")
	for _, opaasInstance := range opaasData.Instances {
		if vm.VMName == opaasInstance.Hostname {
			logrus.WithFields(logFields).Info("Found matching vm")
			return &opaasInstance
		}
	}
	logrus.WithFields(logFields).Info("Failed to find matching vm")
	return nil
}

func createVMCSV(vm VM) *utils.VMCSV {
	return &utils.VMCSV{
		Hostname:           vm.VMName,
		VcenterCPU:         vm.CPU,
		MEMORYREQUESTEDGB:  vm.MEMORYREQUESTEDGB,
		STORAGEREQUESTEDGB: vm.STORAGEREQUESTEDGB,
		Site:               vm.SITEID,
		Cdir:               "",
		Profile:            "",
		ResourceStatus:     "",
		WorkloadType:       "",
		RequestID:          "",
		Memory:             -1,
		CPU:                -1,
		Storage:            -1,
		StoragePoolName:    "",
	}
}

func addOpaasVMCSVInfo(opaasInstance *client.Instance, vmCSV *utils.VMCSV) {
	vmCSV.Cdir = opaasInstance.Cdir
	vmCSV.Profile = opaasInstance.Profile
	vmCSV.ResourceStatus = opaasInstance.ResourceStatus
	vmCSV.WorkloadType = opaasInstance.WorkloadType
	vmCSV.RequestID = opaasInstance.RequestID
	vmCSV.Memory = opaasInstance.Memory
	vmCSV.CPU = opaasInstance.CPU
	vmCSV.Storage = getTotalOpaasInstanceStorage(opaasInstance)
	if len(opaasInstance.Storage) != 0 {
		vmCSV.StoragePoolName = opaasInstance.Storage[0].PoolName

	} else {
		vmCSV.StoragePoolName = ""
	}
}

func getTotalOpaasInstanceStorage(opaasInstance *client.Instance) int {
	totalStorage := 0
	for _, storage := range opaasInstance.Storage {
		totalStorage += storage.Size
	}
	return totalStorage
}

func writeVMCSV(offset int64, vmCSVs []utils.CSVInfo) {
	logFields := logrus.Fields{
		"offset": offset,
	}
	logrus.WithFields(logFields).Info("Writting vm information to csv")
	csvErr := utils.WriteToCSV(instance_csv_file, vmCSVs)
	if csvErr != nil {
		logrus.WithFields(logrus.Fields{
			"offset": offset,
			"Error":  csvErr.Error(),
		}).Error("Failed to write to csv file")
	} else {
		logrus.WithFields(logFields).Info("Successfully wrote vm information to csv")
	}
}
