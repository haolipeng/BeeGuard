package main

import (
	"bufio"
	"os"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"go.uber.org/zap"
)

// KmodHandler 内核模块采集处理器
// 从 /proc/modules 文件读取已加载的内核模块信息
type KmodHandler struct{}

func (*KmodHandler) Name() string {
	return "kmod"
}

func (*KmodHandler) DataType() int {
	return datatype.Kmod
}

func (h *KmodHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	// 打开 /proc/modules 文件
	// 该文件包含当前已加载的内核模块信息
	f, err := os.Open("/proc/modules")
	if err != nil {
		zap.S().Errorf("Failed to open /proc/modules: %v", err)
		return
	}
	defer f.Close()

	// 逐行读取文件
	s := bufio.NewScanner(f)
	for s.Scan() {
		// 按空格分割字段
		fields := strings.Fields(s.Text())

		// /proc/modules 文件格式：
		// name size refcount used_by state addr
		// 例如：nvidia_uvm 1234567 2 - Live 0xffffffffc1234567
		if len(fields) > 5 {
			rec := &businessplugins.Record{
				DataType:  int32(h.DataType()),
				Timestamp: time.Now().Unix(),
				Data: &businessplugins.Payload{
					Fields: map[string]string{
						"name":        fields[0], // 模块名称
						"size":        fields[1], // 模块大小（字节）
						"refcount":    fields[2], // 引用计数
						"used_by":     fields[3], // 使用该模块的模块列表
						"state":       fields[4], // 状态（Live, Loading, Unloading）
						"addr":        fields[5], // 模块在内存中的地址
						"package_seq": seq,       // 包序列号（用于标识本次采集批次）
					},
				},
			}
			c.SendRecord(rec)
		}
	}

	if err := s.Err(); err != nil {
		zap.S().Errorf("Error reading /proc/modules: %v", err)
	}
}
