package watcher

import (
	"os"
	"sync"

	"github.com/nxadm/tail"
	"go.uber.org/zap"
)

// LineHandler 日志行处理函数
type LineHandler func(line string)

// Watcher 日志文件监控器
type Watcher struct {
	paths   []string
	tails   []*tail.Tail
	handler LineHandler
	wg      sync.WaitGroup
	done    chan struct{}
}

// New 创建新的日志监控器
func New(paths []string, handler LineHandler) *Watcher {
	return &Watcher{
		paths:   paths,
		handler: handler,
		done:    make(chan struct{}),
	}
}

// Start 启动监控
func (w *Watcher) Start() error {
	for _, path := range w.paths {
		// 检查文件是否存在
		if _, err := os.Stat(path); os.IsNotExist(err) {
			zap.S().Warnf("log file not found, skipping: %s", path)
			continue
		}

		t, err := tail.TailFile(path, tail.Config{
			Follow:    true,           // 持续跟踪文件
			ReOpen:    true,           // 支持日志轮转后重新打开
			MustExist: false,          // 文件不存在时不报错
			Poll:      true,           // 使用轮询模式(兼容性更好)
			Location:  &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}, // 从文件末尾开始
		})
		if err != nil {
			zap.S().Errorf("failed to tail file %s: %v", path, err)
			continue
		}

		w.tails = append(w.tails, t)
		zap.S().Infof("watching log file: %s", path)

		// 启动监控协程
		w.wg.Add(1)
		go w.watch(t, path)
	}

	if len(w.tails) == 0 {
		zap.S().Warn("no log files to watch")
	}

	return nil
}

// watch 监控单个文件
func (w *Watcher) watch(t *tail.Tail, path string) {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return
		case line, ok := <-t.Lines:
			if !ok {
				zap.S().Warnf("tail channel closed for %s", path)
				return
			}
			if line.Err != nil {
				zap.S().Errorf("tail error for %s: %v", path, line.Err)
				continue
			}
			if line.Text != "" {
				w.handler(line.Text)
			}
		}
	}
}

// Stop 停止监控
func (w *Watcher) Stop() {
	close(w.done)

	for _, t := range w.tails {
		t.Stop()
		t.Cleanup()
	}

	w.wg.Wait()
	zap.S().Info("watcher stopped")
}
