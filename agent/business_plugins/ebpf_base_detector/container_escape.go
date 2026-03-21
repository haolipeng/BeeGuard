package main

import (
	"strings"

	"ebpf_base_detector/events"
)

// ContainerEscapeDetector 容器逃逸检测器
type ContainerEscapeDetector struct{}

// EscapeResult 逃逸检测结果
type EscapeResult struct {
	RuleName    string // 规则名称
	Severity    string // 严重级别
	Description string // 描述
	DevName     string // 触发告警的设备路径
	DirName     string // 挂载目标
}

// 宿主机块设备前缀列表
var blockDevicePrefixes = []string{
	"/dev/sd",   // SCSI/SATA 设备
	"/dev/vd",   // VirtIO 设备
	"/dev/nvme", // NVMe 设备
	"/dev/xvd",  // Xen 设备
	"/dev/hd",   // IDE 设备
}

// NewContainerEscapeDetector 创建容器逃逸检测器
func NewContainerEscapeDetector() *ContainerEscapeDetector {
	return &ContainerEscapeDetector{}
}

// DetectMountEscape 检测 mount 事件是否为容器逃逸
func (d *ContainerEscapeDetector) DetectMountEscape(evt *events.MountEvent) *EscapeResult {
	// 条件1: 必须在容器内（mntns_id != root_mntns_id）
	if !IsContainer(evt.MntnsID, evt.RootMntnsID) {
		return nil
	}

	devName := cstring(evt.DevName[:])

	// 条件2: 挂载源是宿主机块设备
	isBlockDevice := false
	for _, prefix := range blockDevicePrefixes {
		if strings.HasPrefix(devName, prefix) {
			isBlockDevice = true
			break
		}
	}

	if isBlockDevice {
		return &EscapeResult{
			RuleName:    "container_escape_mount_device",
			Severity:    SeverityCritical,
			Description: "容器内挂载宿主机块设备，疑似容器逃逸",
			DevName:     devName,
			DirName:     cstring(evt.DirName[:]),
		}
	}

	return nil
}
