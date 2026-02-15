package main

import (
	"context"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"

	"nids/log"
)

// PacketCapture 网络抓包器
type PacketCapture struct {
	handle    *pcap.Handle
	assembler *tcpassembly.Assembler
	pool      *tcpassembly.StreamPool
	cfg       *NIDSConfig
	logger    *log.Logger
}

// NewPacketCapture 创建抓包器
func NewPacketCapture(cfg *NIDSConfig, factory tcpassembly.StreamFactory, logger *log.Logger) (*PacketCapture, error) {
	handle, err := pcap.OpenLive(cfg.Interface, cfg.Snaplen, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}

	if cfg.BPFFilter != "" {
		if err := handle.SetBPFFilter(cfg.BPFFilter); err != nil {
			handle.Close()
			return nil, err
		}
	}

	pool := tcpassembly.NewStreamPool(factory)
	assembler := tcpassembly.NewAssembler(pool)

	// 配置重组参数
	maxPagesPerConn := cfg.TCPReassembly.MaxBufferSize / 4096
	if maxPagesPerConn < 1 {
		maxPagesPerConn = 1
	}
	assembler.MaxBufferedPagesPerConnection = maxPagesPerConn
	assembler.MaxBufferedPagesTotal = maxPagesPerConn * cfg.TCPReassembly.MaxStreams

	return &PacketCapture{
		handle:    handle,
		assembler: assembler,
		pool:      pool,
		cfg:       cfg,
		logger:    logger,
	}, nil
}

// Run 主抓包循环
func (pc *PacketCapture) Run(ctx context.Context) {
	packetSource := gopacket.NewPacketSource(pc.handle, pc.handle.LinkType())
	packets := packetSource.Packets()

	flushTicker := time.NewTicker(30 * time.Second)
	defer flushTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			pc.logger.Info("Packet capture stopping...")
			pc.assembler.FlushAll()
			return

		case packet, ok := <-packets:
			if !ok {
				pc.logger.Info("Packet source closed")
				return
			}
			if packet == nil {
				continue
			}

			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer == nil {
				continue
			}
			tcp, _ := tcpLayer.(*layers.TCP)

			netLayer := packet.NetworkLayer()
			if netLayer == nil {
				continue
			}

			pc.assembler.AssembleWithTimestamp(
				netLayer.NetworkFlow(),
				tcp,
				packet.Metadata().Timestamp,
			)

		case <-flushTicker.C:
			pc.assembler.FlushOlderThan(time.Now().Add(-pc.cfg.TCPReassembly.StreamTimeout))
		}
	}
}

// Close 关闭抓包器
func (pc *PacketCapture) Close() {
	pc.handle.Close()
}

// httpStreamFactory 实现 tcpassembly.StreamFactory 接口
type httpStreamFactory struct {
	detector     *Detector
	logger       *log.Logger
	maxBodySize  int
	activeStreams int64
	maxStreams    int
	monitorPorts map[uint16]bool
}

// NewHTTPStreamFactory 创建 HTTP 流工厂
func NewHTTPStreamFactory(detector *Detector, logger *log.Logger, cfg *NIDSConfig) *httpStreamFactory {
	ports := parseBPFPorts(cfg.BPFFilter)
	return &httpStreamFactory{
		detector:     detector,
		logger:       logger,
		maxBodySize:  cfg.TCPReassembly.MaxBufferSize,
		maxStreams:    cfg.TCPReassembly.MaxStreams,
		monitorPorts: ports,
	}
}

// New 实现 tcpassembly.StreamFactory 接口
func (f *httpStreamFactory) New(netFlow, tcpFlow gopacket.Flow) tcpassembly.Stream {
	// 检查并发流数量
	current := atomic.LoadInt64(&f.activeStreams)
	if int(current) >= f.maxStreams {
		return &discardStream{}
	}

	// 判断目标端口是否在监控范围内
	dstPort, _ := strconv.Atoi(tcpFlow.Dst().String())
	if !f.monitorPorts[uint16(dstPort)] {
		return &discardStream{}
	}

	atomic.AddInt64(&f.activeStreams, 1)

	s := newHTTPStream(netFlow, tcpFlow, f)
	go func() {
		defer atomic.AddInt64(&f.activeStreams, -1)
		parseHTTPRequests(s.reader, netFlow, tcpFlow, f.detector, f.logger, f.maxBodySize)
	}()

	return s
}

// parseBPFPorts 从 BPF 过滤器字符串中提取端口号
func parseBPFPorts(filter string) map[uint16]bool {
	ports := make(map[uint16]bool)
	tokens := strings.Fields(filter)
	for i, tok := range tokens {
		if tok == "port" && i+1 < len(tokens) {
			portStr := strings.TrimRight(tokens[i+1], ",;)")
			if p, err := strconv.Atoi(portStr); err == nil {
				ports[uint16(p)] = true
			}
		}
	}
	// 至少包含默认端口
	if len(ports) == 0 {
		ports[80] = true
		ports[8080] = true
	}
	return ports
}

// discardStream 丢弃流，不做任何处理
type discardStream struct{}

func (s *discardStream) Reassembled([]tcpassembly.Reassembly) {}
func (s *discardStream) ReassemblyComplete()                  {}
