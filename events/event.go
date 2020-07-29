package events

import (
	"github.com/opaas/capacity-worker/client"
	"github.com/opaas/capacity-worker/utils"
)

type Event interface {
	Process(offset int64, opaasData *opaas.OpaasData, SlData []internal.SoftLayerHosts)
}

func mapSites(site string) string {
	if site == "POK1E" {
		return "POK02"
	}
	if site == "DAL1E" {
		return "DAL00"
	}
	return site
}

