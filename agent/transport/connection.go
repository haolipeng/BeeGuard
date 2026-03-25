package transport

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/haolipeng/BeeGuard/agent/config"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// conn 当前 gRPC 连接
	conn atomic.Value // *grpc.ClientConn

	// retries 重试计数
	retries int32

	// dialOptions gRPC 连接选项（无 TLS 加密）
	dialOptions = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 无 TLS 加密
		grpc.WithStatsHandler(&DefaultStatsHandler),              // 流量统计
		grpc.WithBlock(), // 阻塞直到连接建立
	}

	// ErrMaxRetryExceeded 超过最大重试次数错误
	ErrMaxRetryExceeded = errors.New("max retry count exceeded")

	// ErrNoConnection 无可用连接错误
	ErrNoConnection = errors.New("no connection available")

	// ErrServerNotConfigured 服务器地址未配置错误
	ErrServerNotConfigured = errors.New("server address is not configured")
)

const (
	// defaultDialTimeout 默认连接超时时间
	defaultDialTimeout = 15 * time.Second

	// maxRetryBeforeReconnect 连接正常时的最大重试次数
	// 超过此次数后强制重连，防止长时间使用同一连接
	maxRetryBeforeReconnect = 5
)

// GetConnection 获取 gRPC 连接
// 如果当前连接可用则直接返回，否则尝试建立新连接
// 连接失败时会根据配置进行重试
func GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	// 检查现有连接状态
	if c, ok := conn.Load().(*grpc.ClientConn); ok && c != nil {
		state := c.GetState()
		switch state {
		case connectivity.Ready:
			// 连接就绪，检查重试计数
			// 超过阈值时强制重连，防止长时间使用同一连接
			if atomic.AddInt32(&retries, 1) > maxRetryBeforeReconnect {
				zap.S().Debug("forcing reconnect after max retries on ready connection")
				c.Close()
			} else {
				return c, nil
			}

		case connectivity.Idle:
			// 连接空闲，检查重试计数
			// 超过阈值时强制重连
			if atomic.AddInt32(&retries, 1) > maxRetryBeforeReconnect {
				zap.S().Debug("forcing reconnect after max retries on idle connection")
				c.Close()
			} else {
				return c, nil
			}

		case connectivity.Connecting:
			// 正在连接中，直接关闭并重新建立
			zap.S().Debug("connection is connecting, closing to retry")
			c.Close()

		case connectivity.TransientFailure:
			// 临时失败，关闭连接并重新建立
			zap.S().Debug("connection in transient failure, closing")
			c.Close()

		case connectivity.Shutdown:
			// 连接已关闭
			zap.S().Debug("connection is shutdown")
		}
	}

	// 获取配置
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	if cfg.Server == "" {
		return nil, ErrServerNotConfigured
	}

	// 当前尝试次数超过最大重试次数
	currentRetries := atomic.LoadInt32(&retries)
	maxRetries := int32(cfg.RetryMaxCount)
	if maxRetries > 0 && currentRetries >= maxRetries {
		zap.S().Errorw("max retry count exceeded", "current_retries", currentRetries, "max_retries", maxRetries)
		return nil, fmt.Errorf("%w: %d/%d", ErrMaxRetryExceeded, currentRetries, maxRetries)
	}

	// 重试间隔等待
	retryInterval := time.Duration(cfg.RetryInterval) * time.Second
	if retryInterval > 0 && currentRetries > 0 {
		zap.S().Debugw("waiting before retry", "retry", currentRetries, "interval", retryInterval)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
		}
	}

	// 建立新连接
	connectTimeout := time.Duration(cfg.ConnectTimeout) * time.Second
	if connectTimeout <= 0 {
		connectTimeout = defaultDialTimeout
	}

	zap.S().Infow("dialing server", "server", cfg.Server, "retry", currentRetries, "timeout", connectTimeout)

	dialCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	newConn, err := grpc.DialContext(dialCtx, cfg.Server, dialOptions...)
	if err != nil {
		atomic.AddInt32(&retries, 1)
		zap.S().Errorw("failed to dial server", "server", cfg.Server, "retry", currentRetries+1, "error", err.Error())
		return nil, fmt.Errorf("failed to dial %s (retry %d/%d): %w",
			cfg.Server, currentRetries+1, maxRetries, err)
	}

	// 连接成功,重置重试计数
	conn.Store(newConn)
	atomic.StoreInt32(&retries, 0)
	zap.S().Infow("connected to server", "server", cfg.Server)

	return newConn, nil
}

// CloseConnection 关闭当前连接
func CloseConnection() error {
	c, ok := conn.Load().(*grpc.ClientConn)
	if ok && c != nil {
		conn.Store((*grpc.ClientConn)(nil))
		zap.S().Info("closing connection")
		return c.Close()
	}
	return nil
}

// GetConnectionState 获取连接状态
func GetConnectionState() (connectivity.State, error) {
	c, ok := conn.Load().(*grpc.ClientConn)
	if !ok || c == nil {
		return connectivity.Shutdown, ErrNoConnection
	}
	return c.GetState(), nil
}

// IsConnected 检查是否已连接
func IsConnected() bool {
	state, err := GetConnectionState()
	if err != nil {
		return false
	}
	return state == connectivity.Ready || state == connectivity.Idle
}

// GetRetries 获取当前重试计数
func GetRetries() int32 {
	return atomic.LoadInt32(&retries)
}

// ResetRetries 重置重试计数
func ResetRetries() {
	atomic.StoreInt32(&retries, 0)
}

// ForceReconnect 强制重新连接
// 关闭现有连接并建立新连接
func ForceReconnect(ctx context.Context) (*grpc.ClientConn, error) {
	zap.S().Info("forcing reconnection")
	if err := CloseConnection(); err != nil {
		zap.S().Warnw("error closing connection during force reconnect", "error", err.Error())
	}
	ResetRetries()
	return GetConnection(ctx)
}
