package events

import (
	"github.com/opaas/capacity-worker/client"
	"github.com/opaas/capacity-worker/utils"
	"github.com/sirupsen/logrus"
)

const datastore_csv_file string = "output/Datastores.csv"

type DatastoreEvent struct {
	StreamName string      `json:"streamName"`
	Data       []Datastore `json:"data"`
}

type Datastore struct {
	STREAMNAME    string `json:"STREAM_NAME"`
	DOCID         string `json:"DOCID"`
	TS            string `json:"TS"`
	SNAPSHOTID    int    `json:"SNAPSHOT_ID"`
	DATASTORENAME string `json:"DATASTORE_NAME"`
	PODID         string `json:"PODID"`
	DATACENTER    string `json:"DATACENTER"`
	TOTALGB       int    `json:"TOTAL_GB"`
	REQUESTEDGB   int    `json:"REQUESTED_GB"`
	COMMITTEDGB   int    `json:"COMMITTED_GB"`
	CDATE         string `json:"CDATE"`
	SITEID        string `json:"SITE_ID"`
}

func (event DatastoreEvent) Process(offset int64, opaasData *client.OpaasData, SlData []utils.SoftLayerHosts) {
	datastoreCSVs := []utils.CSVInfo{}
	for _, datastore := range event.Data {
		datastoreCSV := processDatastore(datastore, opaasData)
		datastoreCSVs = append(datastoreCSVs, datastoreCSV)
	}
	writeDatastoreCSV(offset, datastoreCSVs)
}

func processDatastore(datastore Datastore, opaasData *client.OpaasData) *utils.DatastoreCSV {
	datastore.SITEID = mapSites(datastore.SITEID)
	datastoreCSV := createDatastoreCSV(datastore)
	opaasStorage := findAppropriateStorage(datastore, opaasData)
	if opaasStorage != nil {
		addOpaasStorageCSVInfo(opaasStorage, datastoreCSV)
		patchDatastoreIfNecessary(datastore, opaasStorage)
	}
	return datastoreCSV
}

func findAppropriateStorage(datastore Datastore, opaasData *client.OpaasData) *client.Storage {
	storage := findMatchingStorage(datastore, opaasData)
	if storage == nil {
		return nil
	}
	if !clusterExistsAtCorrectSite(datastore, storage, opaasData.Clusters) {
		return nil
	}
	return storage
}

func findMatchingStorage(datastore Datastore, opaasData *client.OpaasData) *client.Storage {
	logFields := logrus.Fields{
		"datastoreName": datastore.DATASTORENAME,
	}
	logrus.WithFields(logFields).Info("Searching for matching storage")
	for _, storage := range opaasData.Storage {
		if storage.Name == datastore.DATASTORENAME {
			logrus.WithFields(logFields).Info("Found matching storage")
			return &storage
		}
	}
	logrus.WithFields(logFields).Info("Failed to find matching storage")
	return nil
}

func clusterExistsAtCorrectSite(datastore Datastore, storage *client.Storage, clusters []client.Cluster) bool {
	for _, cluster := range clusters {
		if cluster.PoolLocation == datastore.SITEID && storageIsAssociatedWithCluster(storage, &cluster) {
			return true
		}
	}
	return false
}

func storageIsAssociatedWithCluster(opaasStorage *client.Storage, cluster *client.Cluster) bool {
	for _, id := range cluster.StorageIds {
		if id == opaasStorage.ID {
			return true
		}
	}
	return false
}

func createDatastoreCSV(datastore Datastore) *utils.DatastoreCSV {
	return &utils.DatastoreCSV{
		Name:         datastore.DATASTORENAME,
		TOTALGB:      datastore.TOTALGB,
		REQUESTEDGB:  datastore.REQUESTEDGB,
		COMMITTEDGB:  datastore.COMMITTEDGB,
		Site:         datastore.SITEID,
		Size:         -1,
		SizeFree:     -1,
		SizeConsumed: -1,
	}
}

func addOpaasStorageCSVInfo(opaasStorage *client.Storage, datastoreCSV *utils.DatastoreCSV) {
	datastoreCSV.Size = opaasStorage.Size
	datastoreCSV.SizeFree = opaasStorage.SizeFree
	datastoreCSV.SizeConsumed = opaasStorage.SizeConsumed
}

func patchDatastoreIfNecessary(dataStore Datastore, opaasStorage *client.Storage) {
	patches := createNecessaryDatastorePatches(dataStore, opaasStorage)
	logFields := logrus.Fields{
		"patches":      patches,
		"storageId":    opaasStorage.ID,
		"datstoreName": dataStore.DATASTORENAME,
	}
	if len(patches) == 0 {
		logrus.WithFields(logFields).Info("Storage is up-to-date with vcenter")
		return
	}
	logrus.WithFields(logFields).Info("Patching storage")
	patchStorage(opaasStorage.ID, patches)
}

func createNecessaryDatastorePatches(datastore Datastore, opaasStorage *client.Storage) []client.Patch {
	patches := []client.Patch{}
	if storagePatchIsNecessary(datastore, opaasStorage) {
		patches = append(patches, createStoragePatch(datastore))
	}
	return patches
}

func storagePatchIsNecessary(datastore Datastore, opaasStorage *client.Storage) bool {
	return opaasStorage.InUseByOpaas != datastore.REQUESTEDGB &&
		opaasStorage.VCenterSizeConsumed != datastore.REQUESTEDGB
}

func createStoragePatch(datastore Datastore) client.Patch {
	return client.Patch{
		Op:    "replace",
		Path:  "/vCenterSizeConsumed",
		Value: datastore.REQUESTEDGB,
	}
}

func patchStorage(storageID string, patches []client.Patch) {
	opaasAPI := client.NewOpaasApi()
	storagePatchErr := opaasAPI.PatchStorage(storageID, patches)
	if storagePatchErr != nil {
		logrus.WithFields(logrus.Fields{
			"Error": storagePatchErr.Error(),
		}).Error("Failed to patch storages")
	} else {
		logrus.WithFields(logrus.Fields{
			"storageId": storageID,
		}).Info("Successfully patched storage")
	}
}

func writeDatastoreCSV(offset int64, datastoreCSV []utils.CSVInfo) {
	logFields := logrus.Fields{
		"offset": offset,
	}
	logrus.WithFields(logFields).Info("Writting datastore information to csv")
	csvErr := utils.WriteToCSV(datastore_csv_file, datastoreCSV)
	if csvErr != nil {
		logrus.WithFields(logrus.Fields{
			"offset": offset,
			"Error":  csvErr.Error(),
		}).Error("Failed to write to csv file")
	} else {
		logrus.WithFields(logFields).Info("Successfully wrote datastore information to csv")
	}
}
