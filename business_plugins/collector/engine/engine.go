package engine

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"sync"
	"time"

	businessplugins "business_plugins/lib"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Records map[string]map[string]string

func (r Records) Find(key, value string) (string, map[string]string, bool) {
	for k, v := range r {
		if v[key] == value {
			return k, v, true
		}
	}
	return "", map[string]string{}, false
}

type Cache struct {
	// DataType-Key-Record
	m  map[int]Records
	mu *sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		m:  map[int]Records{},
		mu: &sync.RWMutex{},
	}
}

// don't modify returned map
func (c *Cache) Get(dt int, key string) (map[string]string, bool) {
	c.mu.RLock()
	res, ok := c.m[dt][key]
	c.mu.RUnlock()
	return res, ok
}

func (c *Cache) Put(dt int, key string, value map[string]string) {
	c.mu.Lock()
	c.m[dt][key] = value
	c.mu.Unlock()
}

func (c *Cache) clear(dt int) {
	c.mu.Lock()
	c.m[dt] = map[string]map[string]string{}
	c.mu.Unlock()
}

type Handler interface {
	Handle(c *businessplugins.Client, cache *Cache, seq string)
	Name() string
	DataType() int
}

type handler struct {
	l *zap.SugaredLogger
	Handler
	done     chan struct{}
	interval time.Duration
}

func (h *handler) Handle(c *businessplugins.Client, cache *Cache) {
	h.l.Info("handling")
	var t struct{}
	select {
	case t = <-h.done:
		f := fnv.New32()
		binary.Write(f, binary.LittleEndian, time.Now().UnixNano())
		seq := hex.EncodeToString(f.Sum(nil))
		h.l.Info("do work")
		cache.clear(h.DataType())
		h.Handler.Handle(c, cache, seq)
	default:
		h.l.Info("wait work")
		t = <-h.done
	}
	h.l.Info("work done")
	h.done <- t
	h.l.Info("handled")
}

type Engine struct {
	m     map[int]*handler // 数据类型到处理器的映射,比如5050对应ProcessHandler,通过AddHandler注册
	s     *cron.Cron
	c     *businessplugins.Client
	cache *Cache
}

func BeforeDawn() time.Duration {
	return -1
}

func (e *Engine) AddHandler(interval time.Duration, h Handler) {
	e.m[h.DataType()] = &handler{
		zap.S().With("name", h.Name()),
		h,
		make(chan struct{}, 1),
		interval,
	}
	e.m[h.DataType()].done <- struct{}{}
}

func (e *Engine) Run() {
	zap.S().Info("engine running")
	// 阶段1：立即并发执行所有handler一次
	for _, h := range e.m {
		go func(h *handler) {
			h.l.Info("init call")
			h.Handle(e.c, e.cache)
		}(h)
	}
	// 阶段2：注册周期性cron定时任务
	for _, h := range e.m {
		var spec string
		minutes := int(h.interval.Minutes())
		if h.interval == BeforeDawn() {
			spec = fmt.Sprintf("%d %d * * *", rand.Intn(60), rand.Intn(6))
		} else if minutes > 0 {
			spec = fmt.Sprintf("@every %dm", minutes)
		} else {
			panic("unknown interval")
		}
		func(h *handler) {
			e.s.AddFunc(spec, func() { h.Handle(e.c, e.cache) })
		}(h)
		h.l.Infof("add func to scheduler: %s", spec)
	}
	go func() {
		zap.S().Info("scheduler running")
		e.s.Run()
	}()
	// receive task until stop
	for {
		//接收服务端任务
		t, err := e.c.ReceiveTask()
		if err != nil {
			break
		}
		zap.S().Infof("received task %+v", t)
		if h, ok := e.m[int(t.DataType)]; ok {
			h.Handle(e.c, e.cache)
			// send result recored
			//发送数据记录结果
			e.c.SendRecord(
				&businessplugins.Record{
					DataType:  5100,
					Timestamp: time.Now().Unix(),
					Data: &businessplugins.Payload{
						Fields: map[string]string{
							"status": "succeed",
							"msg":    "",
							"token":  t.Token,
						},
					}})
		} else {
			// can't find handler
			e.c.SendRecord(
				&businessplugins.Record{
					DataType:  5100,
					Timestamp: time.Now().Unix(),
					Data: &businessplugins.Payload{
						Fields: map[string]string{
							"status": "failed",
							"msg":    "the data_type hasn't been implemented",
							"token":  t.Token,
						},
					}})
		}
	}
	zap.S().Warn("engine will stop")
	e.c.Close()
	e.s.Stop()
}

func New(c *businessplugins.Client, l cron.Logger) *Engine {
	return &Engine{
		map[int]*handler{},
		cron.New(cron.WithChain(cron.SkipIfStillRunning(l)), cron.WithLogger(l)),
		c,
		&Cache{
			m:  map[int]Records{},
			mu: &sync.RWMutex{},
		},
	}
}

// RunOnce 立即执行指定（或全部）Handler一次，用于测试
func (e *Engine) RunOnce(handlerNames []string) {
	zap.S().Info("engine running in once mode")
	for _, h := range e.m {
		// 如果指定了Handler名称，则只执行匹配的
		if len(handlerNames) > 0 && !contains(handlerNames, h.Handler.Name()) {
			continue
		}
		zap.S().Infof("running handler: %s (DataType: %d)", h.Handler.Name(), h.Handler.DataType())
		h.Handle(e.c, e.cache)
		zap.S().Infof("handler %s completed", h.Handler.Name())
	}
	zap.S().Info("all handlers completed, exiting")
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == item {
			return true
		}
	}
	return false
}
