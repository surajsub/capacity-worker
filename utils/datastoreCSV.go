package utils

import (
	"time"
)

type DatastoreCSV struct {
	Name         string `json:"name"`
	Site         string `json:"site"`
	Size         int    `json:"Size"`
	TOTALGB      int    `json:"TOTALGB"`
	SizeFree     int    `json:"sizeFree"`
	SizeConsumed int    `json:"sizeConsumed"`
	REQUESTEDGB  int    `json:"REQUESTEDGB"`
	COMMITTEDGB  int    `json:"COMMITTEDGB"`
}

func (datastoreCSV DatastoreCSV) getKeys() []string {
	return []string{
		"Name",
		"Site",
		"Size",
		"TOTALGB",
		"SizeConsumed",
		"REQUESTEDGB",
		"COMMITTEDGB",
		"SizeFree",
		"Timestamp",
	}
}

func (datastoreCSV DatastoreCSV) getValues() []string {
	return []string{
		datastoreCSV.Name,
		datastoreCSV.Site,
		customItoa(datastoreCSV.Size),
		customItoa(datastoreCSV.TOTALGB),
		customItoa(datastoreCSV.SizeConsumed),
		customItoa(datastoreCSV.REQUESTEDGB),
		customItoa(datastoreCSV.COMMITTEDGB),
		customItoa(datastoreCSV.SizeFree),
		time.Now().String(),
	}
}
