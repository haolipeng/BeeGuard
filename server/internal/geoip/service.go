package geoip

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// QueryResult GeoIP 查询结果
type QueryResult struct {
	Country  string // 国家
	City     string // 城市
	Location string // 完整位置（城市, 国家）
	Error    string // 错误信息
}

// CacheEntry 缓存项
type CacheEntry struct {
	Result    *QueryResult
	Timestamp time.Time
}

// Service GeoIP 查询服务
type Service struct {
	reader       *geoip2.Reader
	cache        map[string]*CacheEntry
	cacheMutex   sync.RWMutex
	cacheTTL     time.Duration
	maxCacheSize int
	enabled      bool
}

// NewService 创建 GeoIP 服务
func NewService(enabled bool, dbPath string, cacheTTL int, maxCacheSize int) (*Service, error) {
	service := &Service{
		enabled:      enabled,
		cache:        make(map[string]*CacheEntry),
		cacheTTL:     time.Duration(cacheTTL) * time.Second,
		maxCacheSize: maxCacheSize,
	}

	if !enabled {
		log.Infof("[GeoIP] GeoIP service disabled")
		return service, nil
	}

	// 打开 GeoLite2 数据库（支持 Country 或 City 数据库）
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open GeoIP database: %w", err)
	}

	service.reader = reader
	log.Infof("[GeoIP] GeoIP service initialized: db=%s, cache_ttl=%ds, max_cache_size=%d",
		dbPath, cacheTTL, maxCacheSize)

	return service, nil
}

// Query 查询 IP 的地理位置信息
func (s *Service) Query(ipStr string) *QueryResult {
	if !s.enabled {
		return &QueryResult{
			Country:  "Unknown",
			City:     "Unknown",
			Location: "Unknown",
			Error:    "GeoIP service disabled",
		}
	}

	// 检查缓存
	if cached := s.getFromCache(ipStr); cached != nil {
		return cached
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		result := &QueryResult{
			Country:  "Unknown",
			City:     "Unknown",
			Location: "Unknown",
			Error:    "invalid IP address",
		}
		s.addToCache(ipStr, result)
		return result
	}

	// 查询 GeoIP 数据库（使用 Country 方法，兼容 City 和 Country 数据库）
	record, err := s.reader.Country(ip)
	if err != nil {
		log.Debugf("[GeoIP] Query failed for %s: %v", ipStr, err)
		result := &QueryResult{
			Country:  "Unknown",
			City:     "Unknown",
			Location: "Unknown",
			Error:    err.Error(),
		}
		s.addToCache(ipStr, result)
		return result
	}

	// 提取国家信息（优先中文，fallback 英文）
	country := record.Country.Names["zh-CN"]
	if country == "" {
		country = record.Country.Names["en"]
	}
	if country == "" {
		country = "Unknown"
	}

	// Country 数据库没有城市信息，设置为空
	city := ""
	location := country

	result := &QueryResult{
		Country:  country,
		City:     city,
		Location: location,
		Error:    "",
	}

	s.addToCache(ipStr, result)
	log.Debugf("[GeoIP] Query success: %s -> %s", ipStr, location)
	return result
}

// getFromCache 从缓存获取
func (s *Service) getFromCache(ip string) *QueryResult {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	entry, exists := s.cache[ip]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Since(entry.Timestamp) > s.cacheTTL {
		return nil
	}

	return entry.Result
}

// addToCache 添加到缓存
func (s *Service) addToCache(ip string, result *QueryResult) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// 简单 LRU：达到最大大小时清空缓存
	if len(s.cache) >= s.maxCacheSize {
		s.cache = make(map[string]*CacheEntry)
		log.Debugf("[GeoIP] Cache cleared (max size reached)")
	}

	s.cache[ip] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
	}
}

// Close 关闭 GeoIP 服务
func (s *Service) Close() error {
	if s.reader != nil {
		if err := s.reader.Close(); err != nil {
			return fmt.Errorf("close GeoIP reader: %w", err)
		}
		log.Infof("[GeoIP] GeoIP service closed")
	}
	return nil
}
