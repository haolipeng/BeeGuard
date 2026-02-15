package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"nids/log"
)

// HTTPRequest 解析后的 HTTP 请求（供检测引擎使用）
type HTTPRequest struct {
	SrcIP      string
	DstIP      string
	SrcPort    uint16
	DstPort    uint16
	Method     string
	URI        string
	Headers    string // 所有 header 拼接为 "Key: Value\r\n" 格式
	Body       []byte
	RawPayload string // Method + URI + Headers + Body 拼接
}

// httpStream TCP 流，包装 tcpreader.ReaderStream
type httpStream struct {
	reader  *tcpreader.ReaderStream
	netFlow gopacket.Flow
	tcpFlow gopacket.Flow
}

// newHTTPStream 创建 HTTP 流
func newHTTPStream(netFlow, tcpFlow gopacket.Flow, factory *httpStreamFactory) *httpStream {
	r := tcpreader.NewReaderStream()
	return &httpStream{
		reader:  &r,
		netFlow: netFlow,
		tcpFlow: tcpFlow,
	}
}

// Reassembled 实现 tcpassembly.Stream 接口
func (s *httpStream) Reassembled(reassembly []tcpassembly.Reassembly) {
	s.reader.Reassembled(reassembly)
}

// ReassemblyComplete 实现 tcpassembly.Stream 接口
func (s *httpStream) ReassemblyComplete() {
	s.reader.ReassemblyComplete()
}

// parseHTTPRequests 从重组后的 TCP 流中解析 HTTP 请求
// 支持 HTTP/1.1 keep-alive（同一连接多个请求）
func parseHTTPRequests(reader io.Reader, netFlow, tcpFlow gopacket.Flow,
	det *Detector, logger *log.Logger, maxBodySize int) {

	bufReader := bufio.NewReader(reader)

	srcIP := netFlow.Src().String()
	dstIP := netFlow.Dst().String()
	srcPort := parsePort(tcpFlow.Src().String())
	dstPort := parsePort(tcpFlow.Dst().String())

	for {
		req, err := http.ReadRequest(bufReader)
		if err != nil {
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				// 非正常 HTTP 数据，可能是非 HTTP 流量或连接关闭
			}
			return
		}

		// 读取 Body（限制大小）
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(io.LimitReader(req.Body, int64(maxBodySize)))
			req.Body.Close()
		}

		// 拼接 Headers
		var headerBuf strings.Builder
		for key, values := range req.Header {
			for _, v := range values {
				headerBuf.WriteString(key)
				headerBuf.WriteString(": ")
				headerBuf.WriteString(v)
				headerBuf.WriteString("\r\n")
			}
		}
		headers := headerBuf.String()

		// URI 做 URL 解码以便规则匹配（真实攻击流量常使用 URL 编码绕过）
		// 保留原始 URI 和解码后的 URI，匹配时使用解码后的
		rawURI := req.RequestURI
		decodedURI := urlDecode(rawURI)

		// 构建 RawPayload（使用解码后的 URI）
		rawPayload := fmt.Sprintf("%s %s\r\n%s\r\n%s",
			req.Method, decodedURI, headers, string(body))

		httpReq := &HTTPRequest{
			SrcIP:      srcIP,
			DstIP:      dstIP,
			SrcPort:    srcPort,
			DstPort:    dstPort,
			Method:     req.Method,
			URI:        decodedURI,
			Headers:    headers,
			Body:       body,
			RawPayload: rawPayload,
		}

		// 检测
		det.Match(httpReq)
	}
}

// parsePort 将端口字符串转为 uint16
func parsePort(s string) uint16 {
	var port uint16
	fmt.Sscanf(s, "%d", &port)
	return port
}

// urlDecode 对 URI 进行 URL 解码
// 解码失败时返回原始字符串
func urlDecode(raw string) string {
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return decoded
}
