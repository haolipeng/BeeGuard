package main

import (
	"time"

	businessplugins "business_plugins/lib"

	"github.com/go-viper/mapstructure/v2"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/port"
	"go.uber.org/zap"
)

// PortHandler 端口采集处理器
// 第一次移植：只使用 procfs 方法，不包含进程关联功能
type PortHandler struct{}

func (h PortHandler) Name() string {
	return "port"
}

func (h PortHandler) DataType() int {
	return 5051 // 端口采集的数据类型
}

func (h *PortHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	// 调用 port.ListeningPorts() 获取所有监听端口
	ports, err := port.ListeningPorts()
	if err != nil {
		zap.S().Errorf("Failed to get listening ports: %v", err)
		return
	}

	// 遍历端口，发送记录
	for _, port := range ports {
		rec := &businessplugins.Record{
			DataType:  int32(h.DataType()),
			Timestamp: time.Now().Unix(),
			Data: &businessplugins.Payload{
				Fields: make(map[string]string, 15),
			},
		}

		// 使用 mapstructure 将 Port 结构体转换为 map[string]string
		// 这样可以自动填充所有字段（Family, Protocol, State, Sport, Dport, Sip, Dip, Uid, Inode, Username）
		err := mapstructure.Decode(port, &rec.Data.Fields)
		if err != nil {
			zap.S().Warnf("Failed to decode port: %v", err)
			continue
		}

		// 添加包序列号（用于标识本次采集批次）
		rec.Data.Fields["package_seq"] = seq

		// 发送记录到 agent
		c.SendRecord(rec)
	}

	zap.S().Infof("Port collection completed, sent %d port records", len(ports))
}
