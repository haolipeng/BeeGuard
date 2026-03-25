package main

import (
	"context"
	"strconv"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/container"
	"github.com/haolipeng/BeeGuard/agent/business_plugins/collector/engine"
	"go.uber.org/zap"
)

// ImageHandler 镜像资产采集处理器
type ImageHandler struct{}

func (h *ImageHandler) Name() string {
	return "image"
}

func (h *ImageHandler) DataType() int {
	return datatype.Image
}

func (h *ImageHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	clients := container.NewClients()
	for _, client := range clients {
		images, err := client.ListImages(context.Background())
		client.Close()
		if err != nil {
			zap.S().Warnf("Failed to list images from %s: %v", client.Runtime(), err)
			continue
		}
		for _, img := range images {
			c.SendRecord(&businessplugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &businessplugins.Payload{
					Fields: map[string]string{
						"image_id":         img.ID,
						"image_name":       img.Name,
						"image_version":    img.Version,
						"image_size":       img.Size,
						"container_count":  strconv.Itoa(img.ContainerCount),
						"image_build_time": img.CreateTime,
						"runtime":          img.Runtime,
						"package_seq":      seq,
					},
				},
			})
		}
		zap.S().Infof("Image collection completed from %s, sent %d image records", client.Runtime(), len(images))
	}
}
