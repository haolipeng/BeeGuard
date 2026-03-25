package main

import (
	"context"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/container"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/engine"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/process"
	"go.uber.org/zap"
)

type ContainerHandler struct{}

func (h *ContainerHandler) Name() string {
	return "container"
}

func (h *ContainerHandler) DataType() int {
	return datatype.Container
}

type Container struct {
	Id         string `mapstructure:"id"`
	Name       string `mapstructure:"name"`
	State      string `mapstructure:"state"`
	ImageId    string `mapstructure:"image_id"`
	ImageName  string `mapstructure:"image_name"`
	Pid        string `mapstructure:"pid"`
	Pns        string `mapstructure:"pns"`
	Runtime    string `mapstructure:"runtime"`
	CreateTime string `mapstructure:"create_time"`
}

func (h *ContainerHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	clients := container.NewClients()
	for _, client := range clients {
		containers, err := client.ListContainers(context.Background())
		client.Close()
		if err != nil {
			zap.S().Warnf("Failed to list containers from %s: %v", client.Runtime(), err)
			continue
		}
		for _, ctr := range containers {
			c.SendRecord(&businessplugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &businessplugins.Payload{
					Fields: map[string]string{
						"id":          ctr.ID,
						"name":        ctr.Name,
						"state":       ctr.State,
						"image_id":    ctr.ImageID,
						"image_name":  ctr.ImageName,
						"pid":         ctr.Pid,
						"pns":         ctr.Pns,
						"runtime":     ctr.Runtime,
						"create_time": ctr.CreateTime,
						"package_seq": seq,
					},
				},
			})
			if ctr.State == container.StateName[int32(container.RUNNING)] && ctr.Pns != "" && process.PnsDiffWithRpns(ctr.Pns) {
				cache.Put(h.DataType(), ctr.Pns, map[string]string{
					"container_id":   ctr.ID,
					"container_name": ctr.Name,
				})
			}
		}
	}
}
