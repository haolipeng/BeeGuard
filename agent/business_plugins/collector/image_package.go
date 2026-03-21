package main

import (
	"bufio"
	"context"
	"strings"
	"time"

	businessplugins "business_plugins/lib"
	"shared/datatype"

	"gitlab.myinterest.top/security/agent/business_plugins/collector/container"
	"gitlab.myinterest.top/security/agent/business_plugins/collector/engine"
	"go.uber.org/zap"
)

// ImagePackageHandler 镜像软件包采集处理器
// 通过 docker exec 进入容器内部，采集已安装的软件包列表（dpkg/rpm/apk）
type ImagePackageHandler struct{}

func (h *ImagePackageHandler) Name() string {
	return "image_package"
}

func (h *ImagePackageHandler) DataType() int {
	return datatype.ImagePackage
}

func (h *ImagePackageHandler) Handle(c *businessplugins.Client, cache *engine.Cache, seq string) {
	clients := container.NewClients()
	for _, client := range clients {
		containers, err := client.ListContainers(context.Background())
		if err != nil {
			zap.S().Warnf("Failed to list containers from %s: %v", client.Runtime(), err)
			client.Close()
			continue
		}

		// 按 ImageID 去重，每个镜像只需进入一个运行中的容器采集
		imageContainers := make(map[string]container.Container)
		for _, ctr := range containers {
			if ctr.State != "running" {
				continue
			}
			if _, ok := imageContainers[ctr.ImageID]; !ok {
				imageContainers[ctr.ImageID] = ctr
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		for imageID, ctr := range imageContainers {
			// 检测容器内 OS 版本
			osVersion := h.detectOSVersion(ctx, client, ctr.ID)

			// 采集软件包
			pkgType, packages := h.collectPackages(ctx, client, ctr.ID)
			if len(packages) == 0 {
				continue
			}

			for _, pkg := range packages {
				c.SendRecord(&businessplugins.Record{
					DataType:  int32(h.DataType()),
					Timestamp: time.Now().Unix(),
					Data: &businessplugins.Payload{
						Fields: map[string]string{
							"image_id":        imageID,
							"image_name":      ctr.ImageName,
							"container_id":    ctr.ID,
							"package_name":    pkg.name,
							"package_version": pkg.version,
							"package_type":    pkgType,
							"os_version":      osVersion,
							"package_seq":     seq,
						},
					},
				})
			}
			zap.S().Infof("Image package collection: image=%s os=%s type=%s packages=%d",
				ctr.ImageName, osVersion, pkgType, len(packages))
		}
		cancel()
		client.Close()
	}
}

type imgPkgInfo struct {
	name    string
	version string
}

// detectOSVersion 检测容器内 OS 版本（从 /etc/os-release 读取）
func (h *ImagePackageHandler) detectOSVersion(ctx context.Context, client container.Client, containerID string) string {
	output, err := client.Exec(ctx, containerID, "cat", "/etc/os-release")
	if err != nil {
		return ""
	}
	return parseOSRelease(string(output))
}

// parseOSRelease 解析 /etc/os-release 内容，返回 "distro version" 格式
// 例如: "ubuntu 22.04", "alpine 3.18.4", "debian 12"
func parseOSRelease(content string) string {
	var id, versionID string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			id = strings.Trim(strings.TrimPrefix(line, "ID="), "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
		}
	}
	if id == "" {
		return ""
	}
	if versionID != "" {
		return id + " " + versionID
	}
	return id
}

// collectPackages 检测容器内包管理器并列出已安装包
func (h *ImagePackageHandler) collectPackages(ctx context.Context, client container.Client, containerID string) (string, []imgPkgInfo) {
	// 尝试 dpkg（Debian/Ubuntu）
	if output, err := client.Exec(ctx, containerID, "dpkg-query", "-W", "-f", "${Package}\t${Version}\n"); err == nil {
		if packages := parseTSVPackages(output); len(packages) > 0 {
			return "dpkg", packages
		}
	}

	// 尝试 rpm（RedHat/CentOS）
	if output, err := client.Exec(ctx, containerID, "rpm", "-qa", "--queryformat", "%{NAME}\t%{VERSION}-%{RELEASE}\n"); err == nil {
		if packages := parseTSVPackages(output); len(packages) > 0 {
			return "rpm", packages
		}
	}

	// 尝试 apk（Alpine）
	if output, err := client.Exec(ctx, containerID, "apk", "list", "--installed"); err == nil {
		if packages := parseApkPackages(output); len(packages) > 0 {
			return "apk", packages
		}
	}

	return "", nil
}

// parseTSVPackages 解析 tab 分隔的 name\tversion 输出
func parseTSVPackages(data []byte) []imgPkgInfo {
	var packages []imgPkgInfo
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 || parts[0] == "" {
			continue
		}
		packages = append(packages, imgPkgInfo{name: parts[0], version: parts[1]})
	}
	return packages
}

// parseApkPackages 解析 apk list 输出
// 格式: "musl-1.2.4-r2 x86_64 {musl} (MIT) [installed]"
func parseApkPackages(data []byte) []imgPkgInfo {
	var packages []imgPkgInfo
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		spaceIdx := strings.Index(line, " ")
		if spaceIdx < 0 {
			continue
		}
		nameVer := line[:spaceIdx]
		name, version := splitApkNameVersion(nameVer)
		if name == "" {
			continue
		}
		packages = append(packages, imgPkgInfo{name: name, version: version})
	}
	return packages
}

// splitApkNameVersion 分割 "musl-1.2.4-r2" 为 ("musl", "1.2.4-r2")
// 从右向左找到第一个 "-数字" 的位置进行分割
func splitApkNameVersion(s string) (string, string) {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '-' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
