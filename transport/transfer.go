package transport

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"gitlab.myinterest.top/security/agent/agent"
	"gitlab.myinterest.top/security/agent/buffer"
	"gitlab.myinterest.top/security/agent/host"
	"gitlab.myinterest.top/security/agent/plugin"
	"gitlab.myinterest.top/security/agent/proto"
)

var (
	txCnt      = uint64(0)
	rxCnt      = uint64(0)
	updateTime = time.Now()
)

// GetState 获取传输统计信息（发送和接收的 TPS）
func GetState(now time.Time) (txTPS, rxTPS float64) {
	instant := now.Sub(updateTime).Seconds()
	if instant != 0 {
		txTPS = float64(atomic.SwapUint64(&txCnt, 0)) / float64(instant)
		rxTPS = float64(atomic.SwapUint64(&rxCnt, 0)) / float64(instant)
	}
	updateTime = now
	return
}

// StartTransfer 启动传输守护进程
func StartTransfer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	slog.Info("transfer daemon startup")
	retries := 0
	subWg := &sync.WaitGroup{}
	defer subWg.Wait()

	for {
		conn, err := GetConnection(ctx)
		if err != nil {
			if retries > 5 {
				slog.Error("transfer will shutdown because of no available connections", slog.String("error", err.Error()))
				return
			}
			slog.Warn("wait to get next connection", slog.Int("retry", retries), slog.String("error", err.Error()))
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				retries++
				continue
			}
		}

		slog.Info("get connection successfully")
		retries = 0
		subCtx, cancel := context.WithCancel(ctx)
		client, err := proto.NewTransferClient(conn).Transfer(subCtx)
		if err != nil {
			slog.Error("failed to create transfer client", slog.String("error", err.Error()))
			cancel()
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		subWg.Add(2)
		go handleSend(subCtx, subWg, client)
		go func() {
			handleReceive(subCtx, subWg, client)
			cancel()
		}()
		subWg.Wait()
		cancel()

		slog.Info("transfer has been canceled, wait next try to transfer")
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

// handleSend 处理数据发送
func handleSend(ctx context.Context, wg *sync.WaitGroup, client proto.Transfer_TransferClient) {
	defer wg.Done()
	defer slog.Info("send handler will exit")
	defer client.CloseSend()

	slog.Info("send handler running")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			recs := buffer.ReadEncodedRecords()
			if len(recs) == 0 {
				continue
			}

			// 获取主机信息
			hostname := ""
			if name, ok := host.Name.Load().(string); ok {
				hostname = name
			}
			ipv4List := []string{}
			if ipv4, ok := host.IPv4.Load().([]string); ok {
				ipv4List = ipv4
			}

			// 构建 PackagedData
			pkg := &proto.PackagedData{
				Records:  recs,
				AgentId:  agent.ID,
				Ipv4:     ipv4List,
				Hostname: hostname,
				Version:  agent.Version,
				Product:  agent.Product,
			}

			err := client.Send(pkg)
			if err != nil {
				slog.Error("failed to send data", slog.String("error", err.Error()))
				return
			}

			atomic.AddUint64(&txCnt, uint64(len(recs)))
			for _, rec := range recs {
				buffer.PutEncodedRecord(rec)
			}
		}
	}
}

// handleReceive 处理命令接收
func handleReceive(ctx context.Context, wg *sync.WaitGroup, client proto.Transfer_TransferClient) {
	defer wg.Done()
	defer slog.Info("receive handler will exit")

	slog.Info("receive handler running")
	for {
		cmd, err := client.Recv()
		if err != nil {
			slog.Error("failed to receive command", slog.String("error", err.Error()))
			return
		}

		slog.Info("received command")
		atomic.AddUint64(&rxCnt, 1)

		// 处理任务命令
		if cmd.Task != nil {
			// Agent 自身的任务
			if cmd.Task.ObjectName == agent.Product {
				// 当前无具体 Agent 任务，只处理关闭命令
				if cmd.Task.DataType == 1060 {
					slog.Info("will shutdown agent")
					agent.Cancel()
					slog.Info("shutdown agent successfully")
					return
				}
			} else {
				// 转发给对应插件的任务
				plg, ok := plugin.Get(cmd.Task.ObjectName)
				if ok {
					err := plg.SendTask(*cmd.Task)
					if err != nil {
						plg.Error("send task to plugin failed: " + err.Error())
					}
				} else {
					slog.Error("can't find plugin", slog.String("plugin", cmd.Task.ObjectName))
				}
			}
			continue
		}

		// 处理配置命令
		agent.SetRunning()
		cfgs := make(map[string]*proto.Config)
		for _, config := range cmd.Configs {
			cfgs[config.Name] = config
		}

		// 同步插件配置（排除 Agent 自身）
		delete(cfgs, agent.Product)
		err = plugin.Sync(cfgs)
		if err != nil {
			slog.Error("failed to sync plugin configs", slog.String("error", err.Error()))
		}
	}
}
