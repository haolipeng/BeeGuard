package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"scanner/engine"
	"scanner/log"
	"scanner/utils"
)

// Scanner 扫描器核心
type Scanner struct {
	engine engine.Engine
	filter *Filter
	logger *log.Logger
}

// New 创建扫描器
func New(eng engine.Engine, filter *Filter, logger *log.Logger) *Scanner {
	return &Scanner{
		engine: eng,
		filter: filter,
		logger: logger,
	}
}

// ScanFile 扫描单个文件
func (s *Scanner) ScanFile(path string) (*MalwareResult, error) {
	// 文件过滤
	if ok, reason := s.filter.ShouldScanWithMagic(path); !ok {
		s.logger.Info("Skipped file", "path", path, "reason", reason)
		return nil, nil
	}

	// 调用引擎扫描
	result, err := s.engine.ScanFile(path)
	if err != nil {
		return nil, fmt.Errorf("engine scan %s: %w", path, err)
	}

	if !result.Infected {
		return nil, nil
	}

	// 构建检出结果
	return s.buildResult(path, result.VirusName)
}

// ScanProcess 扫描进程对应的可执行文件
func (s *Scanner) ScanProcess(pid int) (*MalwareResult, error) {
	exePath := fmt.Sprintf("/proc/%d/exe", pid)

	// 读取符号链接获取实际路径
	realPath, err := os.Readlink(exePath)
	if err != nil {
		return nil, fmt.Errorf("readlink %s: %w", exePath, err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(realPath); err != nil {
		return nil, fmt.Errorf("stat %s: %w", realPath, err)
	}

	return s.ScanFile(realPath)
}

// ScanDirectory 扫描目录
// 返回检出的恶意文件列表
func (s *Scanner) ScanDirectory(dir string, maxDepth int, throttle time.Duration, resultCh chan<- *MalwareResult) error {
	baseDepth := pathDepth(dir)

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Warn("Walk error", "path", path, "error", err)
			return nil // 继续遍历
		}

		// 深度检查
		currentDepth := pathDepth(path) - baseDepth
		if d.IsDir() {
			if currentDepth >= maxDepth {
				return filepath.SkipDir
			}
			// 跳过白名单目录
			if s.filter.IsPathWhitelisted(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// 扫描文件
		result, err := s.ScanFile(path)
		if err != nil {
			s.logger.Warn("Scan error", "path", path, "error", err)
			return nil
		}

		if result != nil {
			resultCh <- result
		}

		// 节流
		if throttle > 0 {
			time.Sleep(throttle)
		}

		return nil
	})
}

// ScanAllProcesses 扫描所有正在运行的进程
func (s *Scanner) ScanAllProcesses(throttle time.Duration, resultCh chan<- *MalwareResult) error {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("read /proc: %w", err)
	}

	scanned := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查是否为进程目录（数字名称）
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// 获取进程可执行文件路径
		exePath := fmt.Sprintf("/proc/%d/exe", pid)
		realPath, err := os.Readlink(exePath)
		if err != nil {
			continue
		}

		// 去重：同一路径只扫描一次
		if scanned[realPath] {
			continue
		}
		scanned[realPath] = true

		result, err := s.ScanFile(realPath)
		if err != nil {
			s.logger.Warn("Process scan error", "pid", pid, "path", realPath, "error", err)
			continue
		}

		if result != nil {
			resultCh <- result
		}

		// 节流
		if throttle > 0 {
			time.Sleep(throttle)
		}
	}

	return nil
}

// buildResult 构建恶意软件检出结果
func (s *Scanner) buildResult(path, virusName string) (*MalwareResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	hash, err := utils.CalcFileHash(path)
	if err != nil {
		s.logger.Warn("Hash calculation failed", "path", path, "error", err)
		hash = &utils.FileHash{MD5: "", SHA256: ""}
	}

	threatType, malwareFamily := ParseVirusName(virusName)

	return &MalwareResult{
		ThreatType:    threatType,
		FileName:      filepath.Base(path),
		FilePath:      path,
		FileSize:      info.Size(),
		FileMD5:       hash.MD5,
		FileSHA256:    hash.SHA256,
		MalwareFamily: malwareFamily,
		ScanTime:      time.Now().Unix(),
	}, nil
}

// pathDepth 计算路径深度
func pathDepth(path string) int {
	clean := filepath.Clean(path)
	if clean == "/" {
		return 0
	}
	return len(strings.Split(clean, string(filepath.Separator))) - 1
}
