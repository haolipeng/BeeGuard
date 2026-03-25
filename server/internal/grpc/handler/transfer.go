package handler

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/db/repository"
	"github.com/haolipeng/BeeGuard/server/internal/geoip"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/mapper"
	"github.com/haolipeng/BeeGuard/server/internal/models/assets/host"
	"github.com/haolipeng/BeeGuard/server/internal/models/common"
	"github.com/haolipeng/BeeGuard/server/internal/models/system"
	"github.com/haolipeng/BeeGuard/server/proto"
	"shared/datatype"

	pb "google.golang.org/protobuf/proto"
)

// 数据包处理队列：Recv 只入队，由 worker 异步调用 handlePackagedData，避免阻塞收包
const (
	pkgChanCapacity = 20000 // 有界 channel 容量
	dispatcherCount = 4     // dispatcher worker 数量
)

// ===== 资产采集 (5050-5099) =====
const (
	dataTypeProcess       int32 = datatype.Process
	dataTypePort          int32 = datatype.Port
	dataTypeUser          int32 = datatype.User
	dataTypeService       int32 = datatype.Service
	dataTypeSoftware      int32 = datatype.Software
	dataTypeContainer     int32 = datatype.Container
	dataTypeEnvSuspicious int32 = datatype.EnvSuspicious
	dataTypeImage         int32 = datatype.Image
	dataTypeImagePackage  int32 = datatype.ImagePackage
	dataTypeWebService    int32 = datatype.WebService
	dataTypeDatabase      int32 = datatype.Database
	dataTypeKmod          int32 = datatype.Kmod
)

// ===== eBPF 实时事件 =====
const (
	dataTypeExecve        int32 = datatype.EventExecve
	dataTypeConnect       int32 = datatype.EventConnect
	dataTypeDNS           int32 = datatype.EventDNS
	dataTypeFileEvent     int32 = datatype.EventFile
	dataTypePerfEventLoss int32 = datatype.EventPerfLoss
)

// ===== 安全告警 (6001-6099) =====
const (
	dataTypeSSHBruteForce       int32 = datatype.AlertSSHBruteForce
	dataTypeFTPBruteForce       int32 = datatype.AlertFTPBruteForce
	dataTypeDangerousCommand    int32 = datatype.AlertDangerousCommand
	dataTypeReverseShell        int32 = datatype.AlertReverseShell
	dataTypeSSHAnomalyLogin     int32 = datatype.AlertSSHAnomalyLogin
	dataTypePrivilegeEscalation int32 = datatype.AlertPrivilegeEscalation
	dataTypeNIDS                int32 = datatype.AlertNIDS
	dataTypeMaliciousRequest    int32 = datatype.AlertMaliciousRequest
	dataTypeSensitiveFile       int32 = datatype.AlertSensitiveFile
)

// ===== 恶意文件扫描 (6060-6069) =====
const (
	dataTypeMalwareScanStatus int32 = datatype.ScannerScanStatus
	dataTypeMalwareFileDetect int32 = datatype.ScannerFileDetect
	dataTypeMalwareProcDetect int32 = datatype.ScannerProcDetect
)

// ===== 任务与基线 =====
const (
	dataTypeTaskResult         int32 = datatype.TaskResult
	dataTypeBaselineResult     int32 = datatype.BaselineCheck
	dataTypeBaselineTaskStatus int32 = datatype.BaselineTaskStatus
)

// ===== 容器安全告警 (7001-7099) =====
const (
	dataTypeContainerDangerousCommand int32 = datatype.AlertContainerDangerousCommand
	dataTypeContainerReverseShell     int32 = datatype.AlertContainerReverseShell
	dataTypeContainerSensitiveFile    int32 = datatype.AlertContainerSensitiveFile
)

// getDataTypeName 获取数据类型名称
func getDataTypeName(dt int32) string {
	switch dt {
	case dataTypeProcess:
		return "Process"
	case dataTypePort:
		return "Port"
	case dataTypeUser:
		return "User"
	case dataTypeService:
		return "Service"
	case dataTypeSoftware:
		return "Software"
	case dataTypeContainer:
		return "Container"
	case dataTypeEnvSuspicious:
		return "EnvSuspicious"
	case dataTypeImage:
		return "Image"
	case dataTypeImagePackage:
		return "ImagePackage"
	case dataTypeWebService:
		return "WebService"
	case dataTypeDatabase:
		return "Database"
	case dataTypeKmod:
		return "Kmod"
	case dataTypeTaskResult:
		return "TaskResult"
	case dataTypeSSHBruteForce:
		return "SSHBruteForce"
	case dataTypeFTPBruteForce:
		return "FTPBruteForce"
	case dataTypeDangerousCommand:
		return "DangerousCommand"
	case dataTypeReverseShell:
		return "ReverseShell"
	case dataTypeSSHAnomalyLogin:
		return "SSHAnomalyLogin"
	case dataTypeExecve:
		return "Execve"
	case dataTypeConnect:
		return "Connect"
	case dataTypeDNS:
		return "DNS"
	case dataTypeFileEvent:
		return "FileEvent"
	case dataTypePerfEventLoss:
		return "PerfEventLoss"
	case dataTypePrivilegeEscalation:
		return "PrivilegeEscalation"
	case dataTypeMaliciousRequest:
		return "MaliciousRequest"
	case dataTypeSensitiveFile:
		return "SensitiveFile"
	case dataTypeNIDS:
		return "NIDS"
	case dataTypeMalwareScanStatus:
		return "MalwareScanStatus"
	case dataTypeMalwareFileDetect:
		return "MalwareFileDetect"
	case dataTypeMalwareProcDetect:
		return "MalwareProcDetect"
	case dataTypeBaselineResult:
		return "BaselineResult"
	case dataTypeBaselineTaskStatus:
		return "BaselineTaskStatus"
	case dataTypeContainerDangerousCommand:
		return "ContainerDangerousCommand"
	case dataTypeContainerReverseShell:
		return "ContainerReverseShell"
	case dataTypeContainerSensitiveFile:
		return "ContainerSensitiveFile"
	default:
		return "Unknown"
	}
}

// AgentInfo 存储连接的 Agent 信息
type AgentInfo struct {
	AgentID     string
	Hostname    string
	IPv4        []string
	Version     string
	Product     string
	LastSeen    time.Time
	LastDBWrite time.Time // 用于心跳写入节流
	CommandCh   chan *proto.Command
}

const agentDBWriteInterval = 5 * time.Minute

func strPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TransferServer 实现 gRPC Transfer 服务
// AgentDataStats Agent 数据统计
type AgentDataStats struct {
	AgentID   string
	Hostname  string
	TotalPkgs int       // 总数据包数
	TotalRecs int       // 总记录数
	LastTime  time.Time // 最后更新时间
}

type TransferServer struct {
	proto.UnimplementedTransferServer

	mu            sync.RWMutex
	agents        map[string]*AgentInfo    // key: agentID
	pkgChan       chan *proto.PackagedData // 数据包队列，由 worker 消费
	pkgChanClosed atomic.Bool              // 标记 pkgChan 是否已关闭，防止关闭后 send panic
	dispatcher    *AssetDispatcher         // 资产批量写入调度器
	dispatcherWg  sync.WaitGroup           // 跟踪 dispatcher goroutine 完成
	AssetRepo     *repository.AssetRepository
	alertRepo     *repository.AlertRepository
	execveRepo    *repository.ExecveRepository
	connectRepo   *repository.ConnectRepository
	dnsRepo       *repository.DNSRepository
	fileEventRepo *repository.FileEventRepository
	baselineRepo  *repository.BaselineRepository
	agentInfoRepo *repository.AgentInfoRepository
	geoIPService  *geoip.Service
	// 数据统计相关
	dataStatsMu   sync.RWMutex               // 保护统计数据
	dataStats     map[string]*AgentDataStats // key: agentID
	statsTicker   *time.Ticker               // 定时输出统计
	statsStopChan chan bool                  // 停止统计定时器
}

// NewTransferServer 创建新的 TransferServer 实例
func NewTransferServer(geoIPService *geoip.Service) *TransferServer {
	s := &TransferServer{
		agents:        make(map[string]*AgentInfo),
		pkgChan:       make(chan *proto.PackagedData, pkgChanCapacity),
		AssetRepo:     repository.NewAssetRepository(),
		alertRepo:     repository.NewAlertRepository(),
		execveRepo:    repository.NewExecveRepository(),
		connectRepo:   repository.NewConnectRepository(),
		dnsRepo:       repository.NewDNSRepository(),
		fileEventRepo: repository.NewFileEventRepository(),
		baselineRepo:  repository.NewBaselineRepository(),
		agentInfoRepo: repository.NewAgentInfoRepository(),
		geoIPService:  geoIPService,
		dataStats:     make(map[string]*AgentDataStats),
		statsStopChan: make(chan bool),
	}

	// 服务启动时重置所有Agent连接状态
	if err := s.agentInfoRepo.ResetAllConnectionStatus(context.Background()); err != nil {
		log.Errorf("[Transfer] 启动时重置Agent连接状态失败: %v", err)
	}

	// 创建并启动资产批量写入调度器
	s.dispatcher = NewAssetDispatcher(s.AssetRepo)
	s.dispatcher.Start()

	// 启动数据包处理 dispatcher worker，Recv 只入队不阻塞
	for i := 0; i < dispatcherCount; i++ {
		s.dispatcherWg.Add(1)
		go s.pkgWorker(i)
	}

	// 启动数据统计定时输出（每30秒）
	s.statsTicker = time.NewTicker(30 * time.Second)
	go s.startStatsReporter()

	return s
}

// pkgWorker 从 pkgChan 取包并调用 handlePackagedData，与 Recv 解耦
func (s *TransferServer) pkgWorker(_ int) {
	defer s.dispatcherWg.Done()
	for pkg := range s.pkgChan {
		s.handlePackagedData(pkg)
	}
}

// updateDataStats 更新 Agent 数据统计
func (s *TransferServer) updateDataStats(agentID, hostname string, recordCount int) {
	s.dataStatsMu.Lock()
	defer s.dataStatsMu.Unlock()

	stats, exists := s.dataStats[agentID]
	if !exists {
		stats = &AgentDataStats{
			AgentID:  agentID,
			Hostname: hostname,
		}
		s.dataStats[agentID] = stats
	}

	stats.TotalPkgs++
	stats.TotalRecs += recordCount
	stats.Hostname = hostname // 更新 hostname（可能变化）
	stats.LastTime = time.Now()
}

// startStatsReporter 启动统计信息定时输出
func (s *TransferServer) startStatsReporter() {
	for {
		select {
		case <-s.statsTicker.C:
			s.reportStats()
		case <-s.statsStopChan:
			s.statsTicker.Stop()
			return
		}
	}
}

// reportStats 输出统计数据
func (s *TransferServer) reportStats() {
	s.dataStatsMu.RLock()
	defer s.dataStatsMu.RUnlock()

	if len(s.dataStats) == 0 {
		return
	}

	// 汇总统计信息
	totalAgents := len(s.dataStats)
	totalPkgs := 0
	totalRecs := 0
	var statsList []*AgentDataStats

	for _, stats := range s.dataStats {
		totalPkgs += stats.TotalPkgs
		totalRecs += stats.TotalRecs
		statsList = append(statsList, stats)
	}

	// 输出汇总日志，暂时注释掉
	// log.Infof("[Transfer] 数据统计汇总 (30秒): agents=%d total_pkgs=%d total_recs=%d",
	// 	totalAgents, totalPkgs, totalRecs)

	// 输出每个 Agent 的详细统计（如果 Agent 数量不多）
	if totalAgents <= 10 {
		for _, stats := range statsList {
			log.Infof("[Transfer]   agent_id=%s hostname=%s pkgs=%d recs=%d last_time=%s",
				stats.AgentID, stats.Hostname, stats.TotalPkgs, stats.TotalRecs,
				stats.LastTime.Format("15:04:05"))
		}
	} else {
		// Agent 数量多时，只输出前10个
		for i := 0; i < 10 && i < len(statsList); i++ {
			stats := statsList[i]
			log.Infof("[Transfer]   agent_id=%s hostname=%s pkgs=%d recs=%d",
				stats.AgentID, stats.Hostname, stats.TotalPkgs, stats.TotalRecs)
		}
		log.Infof("[Transfer]   ... 还有 %d 个 Agent 未显示", totalAgents-10)
	}

	// 清空统计数据（为下一个30秒周期做准备）
	for k := range s.dataStats {
		s.dataStats[k].TotalPkgs = 0
		s.dataStats[k].TotalRecs = 0
	}
}

// Stop 优雅关闭：先排空 dispatcher，再 flush writer 剩余数据
func (s *TransferServer) Stop() {
	// 停止统计定时器
	close(s.statsStopChan)

	// 标记 pkgChan 已关闭，防止 grpcServer.Stop() 后仍在运行的
	// Transfer handler 调用 enqueuePkg 向已关闭的 channel 发送数据导致 panic
	s.pkgChanClosed.Store(true)
	close(s.pkgChan)      // 通知 dispatcher worker 退出
	s.dispatcherWg.Wait() // 等待所有 dispatcher 处理完
	s.dispatcher.Stop()   // 停止 writer（drain + flush 剩余数据）
	log.Infof("[Transfer] TransferServer stopped")
}

// enqueuePkg 将数据包送入队列；队列满时丢弃并打日志，避免阻塞 Recv
func (s *TransferServer) enqueuePkg(pkg *proto.PackagedData, agentID string, isFirst bool) {
	if s.pkgChanClosed.Load() {
		return
	}
	// recover 兜底：pkgChanClosed.Load() 与 channel send 之间存在极小的竞态窗口，
	// Stop() 可能恰好在两者之间 close(pkgChan)，导致向已关闭 channel 发送而 panic
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("[Transfer] enqueuePkg recover: %v (shutdown in progress)", r)
		}
	}()
	select {
	case s.pkgChan <- pkg:
	default:
		if isFirst {
			log.Warnf("[Transfer] 数据包队列已满，丢弃首包 agent_id=%s", agentID)
		} else {
			log.Warnf("[Transfer] 数据包队列已满，丢弃 agent_id=%s records=%d", agentID, len(pkg.Records))
		}
	}
}

// Transfer 处理双向流通信
func (s *TransferServer) Transfer(stream proto.Transfer_TransferServer) error {
	// 1. 先接收第一个包，获取 AgentID
	pkg, err := stream.Recv()
	if err != nil {
		log.Errorf("[Transfer] 接收首包失败: %v", err)
		return err
	}

	agentID := pkg.AgentId
	if agentID == "" {
		log.Warnf("[Transfer] 首包缺少 agent_id")
		return io.EOF
	}

	// 2. 创建 channel 并注册 Agent
	commandCh := make(chan *proto.Command, 100)
	s.registerAgent(pkg, commandCh)
	log.Infof("[Transfer] Agent 连接 agent_id=%s hostname=%s version=%s",
		pkg.AgentId, pkg.Hostname, pkg.Version)

	// 首包入队由 worker 处理，不阻塞 Recv
	s.enqueuePkg(pkg, agentID, true)

	// 3. 启动发送 goroutine（此时 commandCh 已初始化）
	sendDone := make(chan struct{})
	go func() {
		defer close(sendDone)
		for {
			select {
			case cmd := <-commandCh:
				if cmd == nil {
					// nil 是哨兵值，表示连接已关闭
					return
				}
				if err := stream.Send(cmd); err != nil {
					log.Errorf("[Transfer] 发送命令失败 agent_id=%s: %v", agentID, err)
					return
				}
				log.Infof("[Transfer] 发送命令成功 agent_id=%s task=%v", agentID, cmd.Task)
			case <-stream.Context().Done():
				return
			}
		}
	}()

	// 自动下发插件配置
	s.sendOnConnectPlugins(agentID, commandCh)

	// 自动下发任务
	s.sendOnConnectTasks(agentID, commandCh)

	// 4. 接收循环
	for {
		pkg, err := stream.Recv()
		if err == io.EOF {
			log.Infof("[Transfer] 连接正常关闭 agent_id=%s", agentID)
			break
		}
		if err != nil {
			log.Errorf("[Transfer] 接收数据失败 agent_id=%s: %v", agentID, err)
			break
		}

		// 更新 Agent 状态
		s.updateAgent(pkg)

		// 入队由 worker 处理，Recv 尽快返回
		s.enqueuePkg(pkg, agentID, false)
	}

	// 5. 清理：发送 nil 哨兵值通知发送 goroutine 退出（不 close channel，避免 panic）
	s.unregisterAgent(agentID)

	select {
	case commandCh <- nil:
	default:
	}

	<-sendDone
	return nil
}

// registerAgent 注册新 Agent
func (s *TransferServer) registerAgent(pkg *proto.PackagedData, commandCh chan *proto.Command) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents[pkg.AgentId] = &AgentInfo{
		AgentID:     pkg.AgentId,
		Hostname:    pkg.Hostname,
		IPv4:        pkg.Ipv4,
		Version:     pkg.Version,
		Product:     pkg.Product,
		LastSeen:    now,
		LastDBWrite: now,
		CommandCh:   commandCh,
	}

	// 异步写入数据库
	agentID := pkg.AgentId
	version := pkg.Version
	hostname := pkg.Hostname
	hostIPStr := strings.Join(pkg.Ipv4, ",")
	osType := pkg.OsType
	osVersion := pkg.OsVersion

	go func() {
		nowDB := common.DateTime{Time: now}
		agentInfo := &system.AgentInfo{
			AgentID:          agentID,
			AgentVersion:     strPtrIfNotEmpty(version),
			ConnectionStatus: system.ConnectionStatusConnected,
			HostName:         hostname,
			HostIP:           hostIPStr,
			OSType:           osType,
			OSVersion:        strPtrIfNotEmpty(osVersion),
			LastConnectedAt:  &nowDB,
			RegisteredAt:     nowDB,
			CreatedAt:        nowDB,
			UpdatedAt:        nowDB,
		}
		if err := s.agentInfoRepo.RegisterAgent(context.Background(), agentInfo); err != nil {
			log.Errorf("[Transfer] Agent注册写入DB失败 agent_id=%s: %v", agentID, err)
		}
	}()
}

// updateAgent 更新 Agent 状态
func (s *TransferServer) updateAgent(pkg *proto.PackagedData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if agent, ok := s.agents[pkg.AgentId]; ok {
		agent.Hostname = pkg.Hostname
		agent.IPv4 = pkg.Ipv4
		agent.Version = pkg.Version
		agent.LastSeen = time.Now()

		// 心跳节流：每 agentDBWriteInterval 更新一次 last_connected_at
		if time.Since(agent.LastDBWrite) >= agentDBWriteInterval {
			agent.LastDBWrite = time.Now()
			agentID := pkg.AgentId
			go func() {
				if err := s.agentInfoRepo.UpdateLastConnected(context.Background(), agentID); err != nil {
					log.Errorf("[Transfer] Agent心跳更新DB失败 agent_id=%s: %v", agentID, err)
				}
			}()
		}
	}
}

// unregisterAgent 注销 Agent
func (s *TransferServer) unregisterAgent(agentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.agents, agentID)
	log.Infof("[Transfer] Agent 断开 agent_id=%s", agentID)

	// 异步更新数据库连接状态
	go func() {
		if err := s.agentInfoRepo.DisconnectAgent(context.Background(), agentID); err != nil {
			log.Errorf("[Transfer] Agent断开更新DB失败 agent_id=%s: %v", agentID, err)
		}
	}()
}

// handlePackagedData 处理接收到的数据包
func (s *TransferServer) handlePackagedData(pkg *proto.PackagedData) {
	if len(pkg.Records) == 0 {
		return
	}

	// 更新统计数据（替代原来的每条日志输出）
	s.updateDataStats(pkg.AgentId, pkg.Hostname, len(pkg.Records))

	// 构建Agent上下文（用于字段映射）
	if len(pkg.Ipv4) == 0 {
		log.Warnf("[Transfer] agent_id=%s hostname=%s 缺少IPv4地址，跳过数据处理", pkg.AgentId, pkg.Hostname)
		return
	}
	agentCtx := &mapper.AgentContext{
		AgentID:      pkg.AgentId,
		HostName:     pkg.Hostname,
		HostIP:       pkg.Ipv4,
		AgentVersion: pkg.Version,
		MacAddr:      pkg.MacAddr,
		OsType:       pkg.OsType,
		OsVersion:    pkg.OsVersion,
	}

	// 更新主机资产信息
	ctx := context.Background()
	hostObj := mapper.MapHost(agentCtx)
	if err := s.AssetRepo.CreateOrUpdateHost(ctx, hostObj); err != nil {
		log.Errorf("[Transfer] 更新主机信息失败: %v", err)
	}

	// 遍历记录，port/process/user 发送到 dispatcher，其他类型立即处理
	for _, rec := range pkg.Records {
		dataTypeName := getDataTypeName(rec.DataType)
		log.Debugf("[Transfer]   -> data_type=%d(%s) timestamp=%d data_len=%d",
			rec.DataType, dataTypeName, rec.Timestamp, len(rec.Data))

		// 解析 Payload 数据
		if len(rec.Data) > 0 {
			payload := &proto.Payload{}
			if err := pb.Unmarshal(rec.Data, payload); err != nil {
				log.Errorf("[Transfer]      解析失败: %v", err)
				continue
			}

			fields := payload.Fields
			if fields == nil {
				continue
			}

			switch rec.DataType {
			case dataTypePort:
				if p := s.mapPort(fields, agentCtx); p != nil {
					s.dispatcher.SendPort(p)
				}
			case dataTypeProcess:
				if p := s.mapProcess(fields, agentCtx); p != nil {
					s.dispatcher.SendProcess(p)
				}
			case dataTypeUser:
				if a := s.mapAccount(fields, agentCtx); a != nil {
					s.dispatcher.SendAccount(a)
				}
			case dataTypeSoftware:
				if sw := s.mapSoftware(fields, agentCtx); sw != nil {
					s.dispatcher.SendSoftware(sw)
				}
			default:
				// 其他类型仍立即处理
				s.processPayload(ctx, rec.DataType, rec.Timestamp, payload, agentCtx)
			}
		}
	}
}

// processPayload 处理Payload数据：日志记录 + 字段映射 + 数据库写入
func (s *TransferServer) processPayload(ctx context.Context, dataType int32, timestamp int64, payload *proto.Payload, agentCtx *mapper.AgentContext) {
	fields := payload.Fields
	if fields == nil {
		return
	}

	switch dataType {
	// ===== 资产采集 =====
	case dataTypeService:
		s.processService(ctx, fields, agentCtx)
	case dataTypeContainer:
		s.processContainer(ctx, fields, agentCtx)
	case dataTypeEnvSuspicious:
		s.processEnvSuspicious(ctx, fields, agentCtx)
	case dataTypeImage:
		s.processImage(ctx, fields, agentCtx)
	case dataTypeImagePackage:
		s.processImagePackage(ctx, fields, agentCtx)
	case dataTypeWebService:
		s.processWebService(ctx, fields, agentCtx)
	case dataTypeDatabase:
		s.processDatabase(ctx, fields, agentCtx)
	case dataTypeKmod:
		s.processKmod(ctx, fields, agentCtx)

	// ===== eBPF 实时事件 =====
	case dataTypeExecve:
		s.processExecve(ctx, fields, agentCtx)
	case dataTypeConnect:
		s.processConnect(ctx, fields, agentCtx)
	case dataTypeDNS:
		s.processDNS(ctx, fields, agentCtx)
	case dataTypeFileEvent:
		s.processFileEvent(ctx, fields, agentCtx)
	case dataTypePerfEventLoss:
		log.Warnf("[Transfer]      [Perf事件丢失] agent=%s lost_count=%s interval=%ss",
			agentCtx.AgentID, fields["lost_count"], fields["report_interval"])

	// ===== 安全告警 =====
	case dataTypeSSHBruteForce, dataTypeFTPBruteForce:
		s.processBruteForce(ctx, fields, agentCtx)
	case dataTypeDangerousCommand:
		s.processDangerousCommand(ctx, fields, agentCtx)
	case dataTypeReverseShell:
		s.processReverseShell(ctx, fields, agentCtx, timestamp)
	case dataTypeSSHAnomalyLogin:
		s.processAnomalyLogin(ctx, fields, agentCtx)
	case dataTypePrivilegeEscalation:
		s.processPrivilegeEscalation(ctx, fields, agentCtx)
	case dataTypeMaliciousRequest:
		s.processMaliciousRequest(ctx, fields, agentCtx, timestamp)
	case dataTypeSensitiveFile:
		s.processSensitiveFile(ctx, fields, agentCtx)
	case dataTypeNIDS:
		s.processNIDS(ctx, fields, agentCtx, timestamp)

	// ===== 容器安全告警 =====
	case dataTypeContainerDangerousCommand:
		s.processContainerDangerousCommand(ctx, fields, agentCtx)
	case dataTypeContainerReverseShell:
		s.processContainerReverseShell(ctx, fields, agentCtx, timestamp)
	case dataTypeContainerSensitiveFile:
		s.processContainerSensitiveFile(ctx, fields, agentCtx)

	// ===== 恶意文件扫描 =====
	case dataTypeMalwareFileDetect, dataTypeMalwareProcDetect:
		s.processMalwareDetect(ctx, fields, agentCtx)

	// ===== 基线 =====
	case dataTypeBaselineResult:
		s.processBaselineResult(ctx, fields, agentCtx)

	// ===== 仅日志记录 =====
	case dataTypeTaskResult:
		log.Infof("[Transfer]      [任务结果] token=%s status=%s msg=%s",
			fields["token"], fields["status"], fields["msg"])
	case dataTypeMalwareScanStatus:
		log.Infof("[Transfer]      [恶意文件扫描状态] status=%s msg=%s",
			fields["status"], fields["msg"])
	case dataTypeBaselineTaskStatus:
		log.Infof("[Transfer]      [基线任务状态] status=%s token=%s msg=%s",
			fields["status"], fields["token"], fields["msg"])

	default:
		log.Warnf("[Transfer]      [未知类型] fields=%v", fields)
	}
}

// ===== 资产采集处理方法 =====

// mapPort 映射端口记录，返回 *host.Port 或 nil（不写库）
func (s *TransferServer) mapPort(fields map[string]string, agentCtx *mapper.AgentContext) *host.Port {
	log.Debugf("[Transfer]      [端口] protocol=%s local_addr=%s:%s state=%s pid=%s",
		fields["protocol"], fields["sip"], fields["sport"], fields["state"], fields["pid"])
	port := mapper.MapPort(fields, agentCtx)
	if port.Port > 0 {
		return port
	}
	return nil
}

// mapProcess 映射进程记录，返回 *host.Process 或 nil（不写库）
func (s *TransferServer) mapProcess(fields map[string]string, agentCtx *mapper.AgentContext) *host.Process {
	log.Debugf("[Transfer]      [进程] pid=%s ppid=%s name=%s path=%s cmdline=%s",
		fields["pid"], fields["ppid"], fields["comm"], fields["path"], truncateString(fields["cmdline"], 100))
	if fields["container_id"] != "" {
		log.Debugf("[Transfer]             container_id=%s container_name=%s",
			fields["container_id"], fields["container_name"])
	}
	process := mapper.MapProcess(fields, agentCtx)
	if process.Path != "" {
		return process
	}
	return nil
}

// mapAccount 映射账号记录，返回 *host.Account 或 nil（不写库）
func (s *TransferServer) mapAccount(fields map[string]string, agentCtx *mapper.AgentContext) *host.Account {
	log.Debugf("[Transfer]      [用户] username=%s uid=%s gid=%s home=%s shell=%s is_root=%s is_sudo=%s",
		fields["username"], fields["uid"], fields["gid"], fields["home"], fields["shell"],
		fields["is_root"], fields["is_sudo"])
	account := mapper.MapAccount(fields, agentCtx)
	if account.Name != "" {
		return account
	}
	return nil
}

func (s *TransferServer) processService(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [服务] name=%s type=%s command=%s restart=%s",
		fields["name"], fields["type"], truncateString(fields["command"], 80), fields["restart"])
	service := mapper.MapSystemService(fields, agentCtx)
	if service.Name != "" {
		if err := s.AssetRepo.CreateOrUpdateSystemService(ctx, service); err != nil {
			log.Errorf("[Transfer]      系统服务写入失败: %v", err)
		}
	}
}

// mapSoftware 映射软件包记录，返回 *host.Software 或 nil（不写库）
func (s *TransferServer) mapSoftware(fields map[string]string, agentCtx *mapper.AgentContext) *host.Software {
	log.Debugf("[Transfer]      [软件] name=%s version=%s type=%s",
		fields["name"], fields["sversion"], fields["type"])
	software := mapper.MapSoftware(fields, agentCtx)
	if software.Name != "" && software.Type != "" {
		return software
	}
	return nil
}

func (s *TransferServer) processContainer(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [容器] id=%s name=%s image=%s state=%s",
		fields["id"], fields["name"], fields["image_name"], fields["state"])
	container := mapper.MapContainer(fields, agentCtx)
	if container.ContainerID != "" {
		if err := s.AssetRepo.CreateOrUpdateContainer(ctx, container); err != nil {
			log.Errorf("[Transfer]      容器写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processEnvSuspicious(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [可疑环境变量] var_name=%s var_value=%s reasons=%s",
		fields["var_name"], truncateString(fields["var_value"], 100), fields["suspicious_reasons"])
	if fields["var_name"] != "" && fields["suspicious_reasons"] != "" {
		env := mapper.MapEnvSuspicious(fields, agentCtx)
		if err := s.AssetRepo.CreateOrUpdateEnvSuspicious(ctx, env); err != nil {
			log.Errorf("[Transfer]      可疑环境变量写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processImage(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [镜像] image_id=%s image_name=%s version=%s size=%s runtime=%s",
		fields["image_id"], fields["image_name"], fields["image_version"], fields["image_size"], fields["runtime"])
	image := mapper.MapImage(fields, agentCtx)
	if image.ImageID != "" {
		if err := s.AssetRepo.CreateOrUpdateImage(ctx, image); err != nil {
			log.Errorf("[Transfer]      镜像写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processImagePackage(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [镜像软件包] image_id=%s package_name=%s version=%s type=%s",
		fields["image_id"], fields["package_name"], fields["package_version"], fields["package_type"])
	imgPkg := mapper.MapImagePackage(fields, agentCtx)
	if imgPkg.ImageID != "" && imgPkg.PackageName != "" {
		if err := s.AssetRepo.CreateOrUpdateImagePackage(ctx, imgPkg); err != nil {
			log.Errorf("[Transfer]      镜像软件包写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processWebService(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [Web服务] app_name=%s server_type=%s version=%s run_user=%s path=%s",
		fields["app_name"], fields["server_type"], fields["version"], fields["run_user"], fields["path"])
	webService := mapper.MapWebService(fields, agentCtx)
	if webService.Name != "" {
		if err := s.AssetRepo.CreateOrUpdateWebService(ctx, webService); err != nil {
			log.Errorf("[Transfer]      Web服务写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processDatabase(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [数据库] db_type=%s db_version=%s port=%s run_user=%s",
		fields["db_type"], fields["db_version"], fields["port"], fields["run_user"])
	database := mapper.MapDatabase(fields, agentCtx)
	if database.DbType != "" {
		if err := s.AssetRepo.CreateOrUpdateDatabase(ctx, database); err != nil {
			log.Errorf("[Transfer]      数据库服务写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processKmod(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [内核模块] name=%s size=%s used_by=%s",
		fields["name"], fields["size"], fields["used_by"])
	kmod := mapper.MapKmod(fields, agentCtx)
	if kmod.Name != "" {
		if err := s.AssetRepo.CreateOrUpdateKmod(ctx, kmod); err != nil {
			log.Errorf("[Transfer]      内核模块写入失败: %v", err)
		}
	}
}

// ===== eBPF 实时事件处理方法 =====

func (s *TransferServer) processExecve(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [Execve事件] pid=%s tgid=%s ppid=%s uid=%s comm=%s exe_path=%s args=%s",
		fields["pid"], fields["tgid"], fields["ppid"], fields["uid"],
		fields["comm"], fields["exe_path"], truncateString(fields["args"], 80))

	execve := mapper.MapExecve(fields, agentCtx)
	if execve.PID > 0 && execve.Comm != "" {
		if err := s.execveRepo.Create(ctx, execve); err != nil {
			log.Errorf("[Transfer]      execve事件写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processConnect(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [Connect事件] pid=%s comm=%s exe_path=%s protocol=%s remote_ip=%s remote_port=%s",
		fields["pid"], fields["comm"], fields["exe_path"],
		fields["protocol"], fields["remote_ip"], fields["remote_port"])

	connect := mapper.MapConnect(fields, agentCtx)
	if connect.PID > 0 && connect.RemoteIP != "" {
		if err := s.connectRepo.Create(ctx, connect); err != nil {
			log.Errorf("[Transfer]      connect事件写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processDNS(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [DNS事件] pid=%s comm=%s exe_path=%s domain=%s query_type=%s",
		fields["pid"], fields["comm"], fields["exe_path"],
		fields["domain"], fields["query_type"])

	dns := mapper.MapDNS(fields, agentCtx)
	if dns.PID > 0 && dns.Domain != "" {
		if err := s.dnsRepo.Create(ctx, dns); err != nil {
			log.Errorf("[Transfer]      DNS事件写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processFileEvent(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Debugf("[Transfer]      [文件事件] pid=%s comm=%s action=%s new_path=%s old_path=%s",
		fields["pid"], fields["comm"], fields["action"],
		fields["new_path"], fields["old_path"])

	fileEvent := mapper.MapFileEvent(fields, agentCtx)
	if fileEvent.PID > 0 && fileEvent.NewPath != "" {
		if err := s.fileEventRepo.Create(ctx, fileEvent); err != nil {
			log.Errorf("[Transfer]      文件事件写入失败: %v", err)
		}
	}
}

// ===== 安全告警处理方法 =====

func (s *TransferServer) processBruteForce(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [暴力破解告警] attack_type=%s source_ip=%s username=%s attempt_count=%s result=%s",
		fields["service"], fields["source_ip"], fields["target_user"], fields["count"], fields["result"])

	alert := mapper.MapBruteForceAlert(fields, agentCtx, s.geoIPService)
	if alert.SourceIP != "" && alert.Username != "" {
		if err := s.alertRepo.CreateBruteForceAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      暴力破解告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processDangerousCommand(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [高危命令告警] command_type=%s command=%s user=%s",
		fields["command_type"], truncateString(fields["command"], 100), fields["user"])

	alert := mapper.MapDangerousCommandAlert(fields, agentCtx)
	if alert.Command != "" {
		if err := s.alertRepo.CreateDangerousCommandAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      高危命令告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processAnomalyLogin(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [异常登录告警] source_ip=%s login_user=%s level=%s",
		fields["source_ip"], fields["target_user"], fields["level"])

	alert := mapper.MapAbnormalLoginAlert(fields, agentCtx, s.geoIPService)
	if alert.SourceIP != "" && alert.LoginUser != "" {
		if err := s.alertRepo.CreateAbnormalLoginAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      异常登录告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processPrivilegeEscalation(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [本地提权告警] escalated_user=%s parent_process=%s exe_path=%s",
		fields["escalated_user"], fields["parent_process"], fields["exe_path"])

	alert := mapper.MapPrivilegeEscalationAlert(fields, agentCtx)
	if alert.EscalatedUser != "" {
		if err := s.alertRepo.CreatePrivilegeEscalationAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      本地提权告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processSensitiveFile(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [敏感文件告警] rule_name=%s severity=%s action=%s file_path=%s operator_user=%s",
		fields["rule_name"], fields["severity"], fields["action"],
		fields["new_path"], fields["operator_user"])

	alert := mapper.MapFileIntegrityAlert(fields, agentCtx)
	if alert.FilePath != "" {
		if err := s.alertRepo.CreateFileIntegrityAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      敏感文件告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processReverseShell(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext, timestamp int64) {
	log.Warnf("[Transfer]      [反弹Shell告警] comm=%s exe_path=%s remote_ip=%s remote_port=%s args=%s",
		fields["comm"], fields["exe_path"], fields["remote_ip"], fields["remote_port"], truncateString(fields["args"], 100))

	alert := mapper.MapReverseShellAlert(fields, agentCtx, timestamp)
	if alert.CommandLine != "" {
		if err := s.alertRepo.CreateReverseShellAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      反弹Shell告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processMaliciousRequest(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext, timestamp int64) {
	log.Warnf("[Transfer]      [恶意请求检测] event_type=%s rule_name=%s threat_type=%s matched_value=%s domain=%s remote_ip=%s",
		fields["event_type"], fields["rule_name"], fields["threat_type"],
		fields["matched_value"], fields["domain"], fields["remote_ip"])

	alert := mapper.MapMaliciousRequestAlert(fields, agentCtx, timestamp)
	if alert.MaliciousDomain != "" || alert.MaliciousIP != nil {
		if err := s.alertRepo.CreateOrUpdateMaliciousRequestAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      恶意请求告警写入失败: %v", err)
		}
	} else {
		log.Warnf("[Transfer]      恶意请求告警缺少domain和IP，跳过写入")
	}
}

func (s *TransferServer) processNIDS(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext, timestamp int64) {
	log.Warnf("[Transfer]      [NIDS告警] vulnerability=%s src_ip=%s:%s -> dst_ip=%s:%s attack_count=%s",
		fields["vulnerability_name"], fields["src_ip"], fields["src_port"],
		fields["dst_ip"], fields["dst_port"], fields["attack_count"])

	alert := mapper.MapNetworkAttackAlert(fields, agentCtx, timestamp, s.geoIPService)
	if alert.AttackerIP != "" && alert.VulnerabilityName != "" {
		if err := s.alertRepo.CreateNetworkAttackAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      NIDS告警写入失败: %v", err)
		}
	} else {
		log.Warnf("[Transfer]      NIDS告警缺少必填字段，跳过写入")
	}
}

// ===== 恶意文件扫描处理方法 =====

func (s *TransferServer) processContainerDangerousCommand(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [容器高危命令告警] container_id=%s rule_id=%s command=%s uid=%s",
		fields["container_id"], fields["rule_id"], truncateString(fields["command"], 100), fields["uid"])

	alert := mapper.MapContainerDangerousCommandAlert(fields, agentCtx)
	if alert.ContainerID != "" && alert.Command != "" {
		if err := s.alertRepo.CreateContainerDangerousCommandAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      容器高危命令告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processContainerReverseShell(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext, timestamp int64) {
	log.Warnf("[Transfer]      [容器反弹Shell告警] container_id=%s comm=%s exe_path=%s remote_ip=%s remote_port=%s args=%s",
		fields["container_id"], fields["comm"], fields["exe_path"],
		fields["remote_ip"], fields["remote_port"], truncateString(fields["args"], 100))

	alert := mapper.MapContainerReverseShellAlert(fields, agentCtx, timestamp)
	if alert.ContainerID != "" && alert.Comm != "" {
		if err := s.alertRepo.CreateContainerReverseShellAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      容器反弹Shell告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processContainerSensitiveFile(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [容器核心文件告警] container_id=%s rule_name=%s severity=%s action=%s file_path=%s",
		fields["container_id"], fields["rule_name"], fields["severity"],
		fields["action"], fields["new_path"])

	alert := mapper.MapContainerSensitiveFileAlert(fields, agentCtx)
	if alert.ContainerID != "" && alert.FilePath != "" {
		if err := s.alertRepo.CreateContainerSensitiveFileAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      容器核心文件告警写入失败: %v", err)
		}
	}
}

func (s *TransferServer) processMalwareDetect(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Warnf("[Transfer]      [恶意文件检出] threat_type=%s file_path=%s malware_family=%s engine=%s",
		fields["threat_type"], fields["file_path"], fields["malware_family"], fields["detection_engine"])

	alert := mapper.MapMalwareScanAlert(fields, agentCtx)
	if alert.FilePath != "" && alert.ThreatType != "" {
		if err := s.alertRepo.CreateMalwareScanAlert(ctx, alert); err != nil {
			log.Errorf("[Transfer]      恶意文件告警写入失败: %v", err)
		}
	}
}

// ===== 基线处理方法 =====

func (s *TransferServer) processBaselineResult(ctx context.Context, fields map[string]string, agentCtx *mapper.AgentContext) {
	log.Infof("[Transfer]      [基线检查结果] data_len=%d token=%s baseline_id=%s", len(fields["data"]), fields["token"], fields["baseline_id"])
	result, details := mapper.MapBaselineResult(fields, agentCtx)
	if result != nil {
		if err := s.baselineRepo.CreateCheckResult(ctx, result); err != nil {
			log.Errorf("[Transfer]      基线结果写入失败: %v", err)
		} else if len(details) > 0 {
			for _, d := range details {
				d.ResultID = result.ID
			}
			if err := s.baselineRepo.BatchCreateCheckDetails(ctx, details); err != nil {
				log.Errorf("[Transfer]      基线明细写入失败: %v", err)
			}
		}
	}
}

// truncateString 按字符截断字符串，避免日志过长
func truncateString(s string, maxLen int) string {
	if utf8.ValidString(s) {
		runes := []rune(s)
		if len(runes) <= maxLen {
			return s
		}
		return string(runes[:maxLen]) + "..."
	}
	// 非 UTF-8：按字节截断，但避免截断多字节序列的中间位置
	if len(s) <= maxLen {
		return s
	}
	// 从 maxLen 位置向前找到合法的 UTF-8 边界
	truncated := s[:maxLen]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "..."
}

// sendOnConnectPlugins Agent上线时自动下发插件配置
func (s *TransferServer) sendOnConnectPlugins(agentID string, commandCh chan *proto.Command) {
	plugins := config.AppConfig.Plugins
	if len(plugins) == 0 {
		return
	}

	configs := make([]*proto.Config, 0, len(plugins))
	for _, p := range plugins {
		configs = append(configs, &proto.Config{
			Name:    p.Name,
			Type:    "binary",
			Version: p.Version,
		})
	}

	cmd := &proto.Command{Configs: configs}
	select {
	case commandCh <- cmd:
		log.Infof("[Transfer] 下发插件配置成功 agent_id=%s plugins=%d", agentID, len(configs))
	default:
		log.Warnf("[Transfer] 下发插件配置失败(队列满) agent_id=%s", agentID)
	}
}

// sendOnConnectTasks Agent上线时自动下发任务
func (s *TransferServer) sendOnConnectTasks(agentID string, commandCh chan *proto.Command) {
	tasks := config.AppConfig.Tasks
	if len(tasks) == 0 {
		return
	}

	// 延迟下发，等待插件加载完成
	go func() {
		time.Sleep(5 * time.Second)

		for i, t := range tasks {
			cmd := &proto.Command{
				Task: &proto.Task{
					DataType:   t.DataType,
					ObjectName: t.ObjectName,
					Data:       t.Data,
					Token:      fmt.Sprintf("auto-%d-%d", time.Now().UnixNano(), i),
				},
			}
			select {
			case commandCh <- cmd:
				log.Infof("[Transfer] 自动下发任务成功 agent_id=%s data_type=%d object_name=%s", agentID, t.DataType, t.ObjectName)
			default:
				log.Warnf("[Transfer] 自动下发任务失败(队列满) agent_id=%s data_type=%d", agentID, t.DataType)
			}
		}
	}()
}

// SendResult 发送命令的结果
type SendResult int

const (
	SendResultSuccess       SendResult = iota // 发送成功
	SendResultAgentNotFound                   // Agent 不存在
	SendResultQueueFull                       // 命令队列已满
)

// SendCommand 向指定 Agent 发送命令
func (s *TransferServer) SendCommand(agentID string, cmd *proto.Command) bool {
	return s.SendCommandWithError(agentID, cmd) == SendResultSuccess
}

// SendCommandWithError 向指定 Agent 发送命令，返回详细结果
// 持有 RLock 直到发送完成，防止与 unregisterAgent 竞态导致向已注销 Agent 发送
func (s *TransferServer) SendCommandWithError(agentID string, cmd *proto.Command) SendResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, ok := s.agents[agentID]
	if !ok {
		return SendResultAgentNotFound
	}

	select {
	case agent.CommandCh <- cmd:
		return SendResultSuccess
	default:
		log.Warnf("[Transfer] 命令队列已满 agent_id=%s", agentID)
		return SendResultQueueFull
	}
}

// GetAgents 获取所有连接的 Agent 列表
func (s *TransferServer) GetAgents() []*AgentInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*AgentInfo, 0, len(s.agents))
	for _, agent := range s.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetAgent 获取指定 Agent 信息
func (s *TransferServer) GetAgent(agentID string) (*AgentInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agent, ok := s.agents[agentID]
	return agent, ok
}
