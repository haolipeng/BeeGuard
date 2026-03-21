package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ContainerMeta 容器元数据
type ContainerMeta struct {
	ContainerID   string
	ContainerName string
	ImageID       string
	ImageName     string
}

// ContainerMetaCache 容器元数据缓存
type ContainerMetaCache struct {
	mu       sync.RWMutex
	cache    map[string]*containerCacheEntry
	ttl      time.Duration
}

type containerCacheEntry struct {
	meta      *ContainerMeta
	expiresAt time.Time
}

// 匹配 Docker container ID 的正则（64 位十六进制字符串）
var (
	// cgroup v1: /docker/<64hex> 或 docker-<64hex>.scope
	cgroupV1DockerRe = regexp.MustCompile(`/docker/([a-f0-9]{64})`)
	cgroupV1ScopeRe  = regexp.MustCompile(`docker-([a-f0-9]{64})\.scope`)
	// cgroup v2: system.slice/docker-<64hex>.scope
	cgroupV2DockerRe = regexp.MustCompile(`docker-([a-f0-9]{64})\.scope`)
	// containerd: /cri-containerd-<64hex>.scope 或 /<id>
	containerdRe = regexp.MustCompile(`cri-containerd-([a-f0-9]{64})\.scope`)
)

// NewContainerMetaCache 创建容器元数据缓存
func NewContainerMetaCache(ttl time.Duration) *ContainerMetaCache {
	return &ContainerMetaCache{
		cache: make(map[string]*containerCacheEntry),
		ttl:   ttl,
	}
}

// GetContainerID 从 /proc/<tgid>/cgroup 提取容器 ID
func (c *ContainerMetaCache) GetContainerID(tgid uint32) string {
	f, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", tgid))
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if id := extractContainerID(line); id != "" {
			return id
		}
	}
	return ""
}

// extractContainerID 从 cgroup 行中提取容器 ID
func extractContainerID(line string) string {
	// Docker cgroup v1: "12:memory:/docker/abc123..."
	if matches := cgroupV1DockerRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	// Docker cgroup v1 scope: "0::/system.slice/docker-abc123.scope"
	if matches := cgroupV1ScopeRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	// Docker cgroup v2: "0::/system.slice/docker-abc123.scope"
	if matches := cgroupV2DockerRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	// containerd: "0::/system.slice/cri-containerd-abc123.scope"
	if matches := containerdRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// GetContainerMeta 获取容器元数据（带缓存）
// 当前仅返回 container_id，后续可扩展 Docker SDK 查询 name/image
func (c *ContainerMetaCache) GetContainerMeta(containerID string) *ContainerMeta {
	if containerID == "" {
		return nil
	}

	// 检查缓存
	c.mu.RLock()
	if entry, ok := c.cache[containerID]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.meta
	}
	c.mu.RUnlock()

	// 缓存未命中，尝试从 /proc/1/mountinfo 或 Docker socket 获取信息
	meta := &ContainerMeta{
		ContainerID: containerID,
	}

	// 尝试从 Docker API 获取容器名称和镜像（通过 unix socket）
	name, image := queryDockerInfo(containerID)
	meta.ContainerName = name
	meta.ImageName = image

	// 写入缓存
	c.mu.Lock()
	c.cache[containerID] = &containerCacheEntry{
		meta:      meta,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return meta
}

// queryDockerInfo 通过读取 Docker API unix socket 获取容器信息
// 返回 (container_name, image_name)
// 不引入 Docker SDK 依赖，直接通过 HTTP over unix socket 查询
func queryDockerInfo(containerID string) (string, string) {
	// 尝试连接 Docker socket
	socketPath := "/var/run/docker.sock"
	if _, err := os.Stat(socketPath); err != nil {
		return "", ""
	}

	// 使用 net 包直接连接 unix socket 发送 HTTP 请求
	// 简化实现：读取 /proc/pid 信息来获取容器名
	// 完整实现需要 Docker SDK，暂时返回短 ID
	shortID := containerID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	return shortID, ""
}

// CleanExpired 清理过期的缓存条目
func (c *ContainerMetaCache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, entry := range c.cache {
		if now.After(entry.expiresAt) {
			delete(c.cache, id)
		}
	}
}

// IsContainer 判断给定的 mntns_id 和 root_mntns_id 是否在容器内
func IsContainer(mntnsID, rootMntnsID uint64) bool {
	return mntnsID != rootMntnsID && rootMntnsID != 0
}

// GetShortContainerID 返回短容器 ID（前12位）
func GetShortContainerID(containerID string) string {
	if len(containerID) > 12 {
		return containerID[:12]
	}
	return containerID
}

// enrichContainerFields 向 record fields 中添加容器元数据
func enrichContainerFields(fields map[string]string, tgid uint32, cache *ContainerMetaCache) {
	if cache == nil {
		return
	}
	cid := cache.GetContainerID(tgid)
	if cid == "" {
		return
	}
	fields["container_id"] = cid
	fields["container_id_short"] = GetShortContainerID(cid)
	meta := cache.GetContainerMeta(cid)
	if meta != nil {
		if meta.ContainerName != "" {
			fields["container_name"] = meta.ContainerName
		}
		if meta.ImageName != "" {
			fields["image_name"] = meta.ImageName
		}
	}
}

// parseContainerIDFromEnv 尝试从进程环境变量获取容器 ID
// 某些运行时会设置 HOSTNAME 为容器短 ID
func parseContainerIDFromEnv(tgid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", tgid))
	if err != nil {
		return ""
	}
	// 环境变量以 NULL 分隔
	for _, kv := range strings.Split(string(data), "\x00") {
		if strings.HasPrefix(kv, "HOSTNAME=") {
			return strings.TrimPrefix(kv, "HOSTNAME=")
		}
	}
	return ""
}
