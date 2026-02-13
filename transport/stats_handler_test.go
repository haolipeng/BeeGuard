package transport

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/stats"
)

func TestStatsHandler_HandleRPC(t *testing.T) {
	h := &StatsHandler{
		updateTime: time.Now(),
	}

	// 模拟接收数据
	h.HandleRPC(context.Background(), &stats.InPayload{
		WireLength: 100,
	})
	h.HandleRPC(context.Background(), &stats.InPayload{
		WireLength: 200,
	})

	// 模拟发送数据
	h.HandleRPC(context.Background(), &stats.OutPayload{
		WireLength: 50,
	})
	h.HandleRPC(context.Background(), &stats.OutPayload{
		WireLength: 150,
	})

	rx, tx := h.GetTotalBytes()
	if rx != 300 {
		t.Errorf("expected rx=300, got %d", rx)
	}
	if tx != 200 {
		t.Errorf("expected tx=200, got %d", tx)
	}
}

func TestStatsHandler_GetStats(t *testing.T) {
	h := &StatsHandler{
		updateTime: time.Now().Add(-time.Second), // 1秒前
	}

	// 模拟数据
	h.HandleRPC(context.Background(), &stats.InPayload{WireLength: 1000})
	h.HandleRPC(context.Background(), &stats.OutPayload{WireLength: 500})

	// 获取统计
	s := h.GetStats()

	// 验证速度（大约 1000 bytes/s 和 500 bytes/s）
	if s.RxSpeed < 900 || s.RxSpeed > 1100 {
		t.Errorf("expected RxSpeed ~1000, got %f", s.RxSpeed)
	}
	if s.TxSpeed < 450 || s.TxSpeed > 550 {
		t.Errorf("expected TxSpeed ~500, got %f", s.TxSpeed)
	}

	// 验证计数器已重置
	rx, tx := h.GetTotalBytes()
	if rx != 0 || tx != 0 {
		t.Errorf("expected counters reset, got rx=%d, tx=%d", rx, tx)
	}
}

func TestStatsHandler_TagRPC(t *testing.T) {
	h := &StatsHandler{}
	ctx := context.Background()

	newCtx := h.TagRPC(ctx, &stats.RPCTagInfo{})
	if newCtx != ctx {
		t.Error("TagRPC should return the same context")
	}
}

func TestStatsHandler_TagConn(t *testing.T) {
	h := &StatsHandler{}
	ctx := context.Background()

	newCtx := h.TagConn(ctx, &stats.ConnTagInfo{})
	if newCtx != ctx {
		t.Error("TagConn should return the same context")
	}
}

func TestDefaultStatsHandler(t *testing.T) {
	// 验证默认处理器已初始化
	if DefaultStatsHandler.updateTime.IsZero() {
		t.Error("DefaultStatsHandler.updateTime should not be zero")
	}
}
