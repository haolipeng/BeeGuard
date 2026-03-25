package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// DiskCache 磁盘缓存
type DiskCache struct {
	cacheDir string
	ttl      time.Duration
	mu       sync.RWMutex
	data     map[string]cacheEntry
}

type cacheEntry struct {
	AnalyzedAt time.Time `json:"analyzed_at"`
}

// NewDiskCache 创建磁盘缓存
func NewDiskCache(cacheDir string, ttl time.Duration) *DiskCache {
	c := &DiskCache{
		cacheDir: cacheDir,
		ttl:      ttl,
		data:     make(map[string]cacheEntry),
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Warnf("[AnalysisCache] 创建缓存目录失败: %v", err)
	}

	// 从磁盘加载已有缓存
	c.loadFromDisk()

	// 启动后台清理过期条目
	go c.cleanupLoop()

	return c
}

// cacheFilePath 缓存文件路径
func (c *DiskCache) cacheFilePath() string {
	return filepath.Join(c.cacheDir, "analysis_cache.json")
}

// loadFromDisk 从磁盘加载缓存
func (c *DiskCache) loadFromDisk() {
	data, err := os.ReadFile(c.cacheFilePath())
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("[AnalysisCache] 读取缓存文件失败: %v", err)
		}
		return
	}

	var loaded map[string]cacheEntry
	if err := json.Unmarshal(data, &loaded); err != nil {
		log.Warnf("[AnalysisCache] 解析缓存文件失败: %v", err)
		return
	}

	c.mu.Lock()
	// 只加载未过期的条目
	now := time.Now()
	for k, v := range loaded {
		if now.Sub(v.AnalyzedAt) < c.ttl {
			c.data[k] = v
		}
	}
	c.mu.Unlock()

	log.Infof("[AnalysisCache] 从磁盘加载 %d 条有效缓存", len(c.data))
}

// saveToDisk 保存缓存到磁盘
func (c *DiskCache) saveToDisk() {
	c.mu.RLock()
	data, err := json.MarshalIndent(c.data, "", "  ")
	c.mu.RUnlock()

	if err != nil {
		log.Warnf("[AnalysisCache] 序列化缓存失败: %v", err)
		return
	}

	if err := os.WriteFile(c.cacheFilePath(), data, 0644); err != nil {
		log.Warnf("[AnalysisCache] 写入缓存文件失败: %v", err)
	}
}

// IsAnalyzed 检查告警是否已被分析
func (c *DiskCache) IsAnalyzed(alertType string, alertID int64) bool {
	key := c.makeKey(alertType, alertID)

	c.mu.RLock()
	entry, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	// 检查是否过期
	if time.Since(entry.AnalyzedAt) > c.ttl {
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return false
	}

	return true
}

// MarkAnalyzed 标记告警为已分析
func (c *DiskCache) MarkAnalyzed(alertType string, alertID int64) {
	key := c.makeKey(alertType, alertID)

	c.mu.Lock()
	c.data[key] = cacheEntry{
		AnalyzedAt: time.Now(),
	}
	c.mu.Unlock()

	// 异步保存到磁盘
	go c.saveToDisk()
}

// MarkBatch 批量标记已分析
func (c *DiskCache) MarkBatch(alerts []AlertContext) {
	now := time.Now()

	c.mu.Lock()
	for _, a := range alerts {
		key := c.makeKey(a.AlertType, a.ID)
		c.data[key] = cacheEntry{
			AnalyzedAt: now,
		}
	}
	c.mu.Unlock()

	// 异步保存到磁盘
	go c.saveToDisk()
}

// FilterAnalyzed 过滤掉已分析的告警
func (c *DiskCache) FilterAnalyzed(alerts []AlertContext) []AlertContext {
	var result []AlertContext
	for _, a := range alerts {
		if !c.IsAnalyzed(a.AlertType, a.ID) {
			result = append(result, a)
		}
	}
	return result
}

// makeKey 生成缓存键
func (c *DiskCache) makeKey(alertType string, alertID int64) string {
	return fmt.Sprintf("%s:%d", alertType, alertID)
}

// cleanupLoop 定期清理过期条目
func (c *DiskCache) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期条目
func (c *DiskCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expired := 0
	for k, v := range c.data {
		if now.Sub(v.AnalyzedAt) > c.ttl {
			delete(c.data, k)
			expired++
		}
	}

	if expired > 0 {
		log.Infof("[AnalysisCache] 清理 %d 条过期缓存", expired)
		go c.saveToDisk()
	}
}

// Stats 缓存统计
func (c *DiskCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"total_entries": len(c.data),
		"ttl_hours":     c.ttl.Hours(),
		"cache_dir":     c.cacheDir,
	}
}
