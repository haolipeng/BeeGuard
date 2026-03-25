package handler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
)

// WriterConfig 批量写入器配置
type WriterConfig struct {
	Name          string
	DataType      int32
	ChannelCap    int
	BatchSize     int
	FlushInterval time.Duration
	WriterCount   int
}

// AssetWriter 泛型批量写入器，通过 channel 聚合跨 packet 的记录并定期批量落库
type AssetWriter[T any] struct {
	config  WriterConfig
	ch      chan T
	flushFn func(ctx context.Context, batch []T) error
	keyFn   func(T) string // 去重键提取函数，避免同批次内 ON CONFLICT 冲突
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewAssetWriter 创建批量写入器（调用 Start 后才启动 goroutine）
// keyFn 用于提取去重键，同一批次内相同键的记录只保留最后一条
func NewAssetWriter[T any](config WriterConfig, flushFn func(ctx context.Context, batch []T) error, keyFn func(T) string) *AssetWriter[T] {
	return &AssetWriter[T]{
		config:  config,
		ch:      make(chan T, config.ChannelCap),
		flushFn: flushFn,
		keyFn:   keyFn,
		stopCh:  make(chan struct{}),
	}
}

// Start 启动 WriterCount 个后台写入 goroutine
func (w *AssetWriter[T]) Start() {
	for i := 0; i < w.config.WriterCount; i++ {
		w.wg.Add(1)
		go w.writerRun(i)
	}
	log.Infof("[AssetWriter] %s writer started: writers=%d, chanCap=%d, batchSize=%d, flushInterval=%s",
		w.config.Name, w.config.WriterCount, w.config.ChannelCap, w.config.BatchSize, w.config.FlushInterval)
}

// Stop 通知所有 writer 退出，drain 剩余数据后返回
func (w *AssetWriter[T]) Stop() {
	close(w.stopCh)
	w.wg.Wait()
	log.Infof("[AssetWriter] %s writer stopped", w.config.Name)
}

// Send 将单条记录发送到 writer channel；channel 满时阻塞（背压），已停止时丢弃
func (w *AssetWriter[T]) Send(item T) {
	select {
	case w.ch <- item:
	case <-w.stopCh:
	}
}

// writerRun 核心批量写入循环：按 BatchSize 或 FlushInterval 触发落库
func (w *AssetWriter[T]) writerRun(id int) {
	defer w.wg.Done()
	batch := make([]T, 0, w.config.BatchSize)
	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		// 去重：同一批次内相同键的记录只保留最后一条，避免 ON CONFLICT 冲突
		if w.keyFn != nil {
			seen := make(map[string]int, len(batch))
			deduped := make([]T, 0, len(batch))
			for _, item := range batch {
				key := w.keyFn(item)
				if idx, ok := seen[key]; ok {
					deduped[idx] = item // 后来的覆盖先来的
				} else {
					seen[key] = len(deduped)
					deduped = append(deduped, item)
				}
			}
			batch = deduped
		}
		// 按冲突键排序，保证并发事务以相同顺序加锁，避免死锁
		if w.keyFn != nil {
			sort.Slice(batch, func(i, j int) bool {
				return w.keyFn(batch[i]) < w.keyFn(batch[j])
			})
		}

		log.Debugf("[AssetWriter] %s writer#%d flushing batch: %d records", w.config.Name, id, len(batch))
		// 带死锁重试的写入
		const maxRetries = 3
		var err error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				// 随机退避，避免重试风暴
				jitter := time.Duration(rand.Int63n(int64(100*time.Millisecond))) + time.Duration(attempt)*200*time.Millisecond
				time.Sleep(jitter)
				log.Warnf("[AssetWriter] %s writer#%d deadlock detected, retry %d/%d", w.config.Name, id, attempt, maxRetries)
			}
			err = w.flushFn(context.Background(), batch)
			if err == nil || !isDeadlockError(err) {
				break
			}
		}
		if err != nil {
			log.Errorf("[AssetWriter] %s writer#%d flush error: %v", w.config.Name, id, err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case item, ok := <-w.ch:
			if !ok {
				flush()
				return
			}
			batch = append(batch, item)
			if len(batch) >= w.config.BatchSize {
				flush()
				ticker.Reset(w.config.FlushInterval)
			}
		case <-ticker.C:
			flush()
		case <-w.stopCh:
			// drain 剩余数据
			for {
				select {
				case item := <-w.ch:
					batch = append(batch, item)
					if len(batch) >= w.config.BatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

// isDeadlockError 检查是否为 PostgreSQL 死锁错误 (SQLSTATE 40P01)
func isDeadlockError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40P01"
	}
	return false
}

// AssetDispatcher 资产批量写入调度器，持有 per-type AssetWriter
type AssetDispatcher struct {
	portWriter     *AssetWriter[*host.Port]
	processWriter  *AssetWriter[*host.Process]
	accountWriter  *AssetWriter[*host.Account]
	softwareWriter *AssetWriter[*host.Software]
}

// NewAssetDispatcher 用 AssetRepository 的批量方法构造调度器
func NewAssetDispatcher(assetRepo *repository.AssetRepository) *AssetDispatcher {
	return &AssetDispatcher{
		processWriter: NewAssetWriter(WriterConfig{
			Name:          "Process",
			DataType:      dataTypeProcess,
			ChannelCap:    1000,
			BatchSize:     500,
			FlushInterval: 5 * time.Second,
			WriterCount:   2,
		}, assetRepo.BatchCreateOrUpdateProcesses, func(p *host.Process) string {
			return p.AgentID + "|" + p.Path
		}),

		portWriter: NewAssetWriter(WriterConfig{
			Name:          "Port",
			DataType:      dataTypePort,
			ChannelCap:    1000,
			BatchSize:     500,
			FlushInterval: 5 * time.Second,
			WriterCount:   2,
		}, assetRepo.BatchCreateOrUpdatePorts, func(p *host.Port) string {
			return fmt.Sprintf("%s|%d|%d", p.AgentID, p.Port, p.Protocol)
		}),

		accountWriter: NewAssetWriter(WriterConfig{
			Name:          "Account",
			DataType:      dataTypeUser,
			ChannelCap:    500,
			BatchSize:     200,
			FlushInterval: 5 * time.Second,
			WriterCount:   1,
		}, assetRepo.BatchCreateOrUpdateAccounts, func(a *host.Account) string {
			return a.AgentID + "|" + a.Name
		}),

		softwareWriter: NewAssetWriter(WriterConfig{
			Name:          "Software",
			DataType:      dataTypeSoftware,
			ChannelCap:    1000,
			BatchSize:     500,
			FlushInterval: 5 * time.Second,
			WriterCount:   2,
		}, assetRepo.BatchCreateOrUpdateSoftware, func(s *host.Software) string {
			return s.AgentID + "|" + s.Name + "|" + s.Type
		}),
	}
}

// Start 启动所有 writer
func (d *AssetDispatcher) Start() {
	d.processWriter.Start()
	d.portWriter.Start()
	d.accountWriter.Start()
	d.softwareWriter.Start()
}

// Stop 停止所有 writer（drain + flush 剩余数据）
func (d *AssetDispatcher) Stop() {
	d.processWriter.Stop()
	d.portWriter.Stop()
	d.accountWriter.Stop()
	d.softwareWriter.Stop()
}

// SendPort 发送端口记录到 writer
func (d *AssetDispatcher) SendPort(p *host.Port) { d.portWriter.Send(p) }

// SendProcess 发送进程记录到 writer
func (d *AssetDispatcher) SendProcess(p *host.Process) { d.processWriter.Send(p) }

// SendAccount 发送账号记录到 writer
func (d *AssetDispatcher) SendAccount(a *host.Account) { d.accountWriter.Send(a) }

// SendSoftware 发送软件包记录到 writer
func (d *AssetDispatcher) SendSoftware(s *host.Software) { d.softwareWriter.Send(s) }
