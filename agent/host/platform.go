package host

import (
	"github.com/shirou/gopsutil/v3/host"
)

var (
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	KernelVersion   string
	Arch            string
)

func init() {
	// 获取内核版本
	KernelVersion, _ = host.KernelVersion()
	// 获取平台信息
	Platform, PlatformFamily, PlatformVersion, _ = host.PlatformInformation()
	// 获取架构信息
	Arch, _ = host.KernelArch()
}
