package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/stats"
)

// DefaultStatsHandler 默认的流量统计处理器
var DefaultStatsHandler = StatsHandler{
	updateTime: time.Now(),
}

// StatsHandler gRPC 流量统计处理器
// 实现 google.golang.org/grpc/stats.Handler 接口
type StatsHandler struct {
	rxBytes    uint64    // 接收字节数（原子操作）
	txBytes    uint64    // 发送字节数（原子操作）
	updateTime time.Time // 上次更新时间
	mu         sync.Mutex
}

// Stats 流量统计结果
type Stats struct {
	RxSpeed float64 // 接收速度（字节/秒）
	TxSpeed float64 // 发送速度（字节/秒）
}

// GetStats 获取流量统计
// 返回自上次调用以来的平均收发速度
// 调用后会重置计数器
func (h *StatsHandler) GetStats() Stats {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	duration := now.Sub(h.updateTime).Seconds()

	var s Stats
	if duration > 0 {
		s.RxSpeed = float64(atomic.SwapUint64(&h.rxBytes, 0)) / duration
		s.TxSpeed = float64(atomic.SwapUint64(&h.txBytes, 0)) / duration
		h.updateTime = now
	}
	return s
}

// GetTotalBytes 获取累计字节数（不重置）
func (h *StatsHandler) GetTotalBytes() (rx, tx uint64) {
	return atomic.LoadUint64(&h.rxBytes), atomic.LoadUint64(&h.txBytes)
}

// TagRPC 实现 stats.Handler 接口
func (h *StatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC 处理 RPC 统计
// 累加接收和发送的字节数
func (h *StatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch v := s.(type) {
	case *stats.InPayload:
		atomic.AddUint64(&h.rxBytes, uint64(v.WireLength))
	case *stats.OutPayload:
		atomic.AddUint64(&h.txBytes, uint64(v.WireLength))
	}
}

// TagConn 实现 stats.Handler 接口
func (h *StatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn 实现 stats.Handler 接口
func (h *StatsHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	// no-op
}
