package client

const storage_endpoint string = "storage"

type Storage struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	SizeAvailable       int    `json:"sizeAvailable"`
	SizeFree            int    `json:"sizeFree"`
	SizeConsumed        int    `json:"sizeConsumed"`
	Size                int    `json:"size"`
	InUseByOpaas        int    `json:"inUseByOpaas"`
	VCenterSizeConsumed int    `json:"vCenterSizeConsumed"`
}

func (opaasApi *OpaasApi) GetStorage() ([]Storage, error) {
	storageData := []Storage{}
	httpErr := opaasApi.get(storage_endpoint, &storageData)
	if httpErr != nil {
		return nil, httpErr
	}
	return storageData, nil
}

func (opaasApi *OpaasApi) PatchStorage(storageId string, patches []Patch) error {
	return opaasApi.patch(storage_endpoint, storageId, patches)
}
