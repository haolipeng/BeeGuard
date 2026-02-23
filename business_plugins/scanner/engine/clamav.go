package engine

/*
#cgo LDFLAGS: -lclamav
#include <clamav.h>
#include <stdlib.h>
#include <string.h>

// 创建默认扫描选项
static struct cl_scan_options default_scan_options() {
    struct cl_scan_options opts;
    memset(&opts, 0, sizeof(opts));
    opts.general = CL_SCAN_GENERAL_ALLMATCHES | CL_SCAN_GENERAL_HEURISTICS;
    opts.parse = CL_SCAN_PARSE_ARCHIVE | CL_SCAN_PARSE_ELF | CL_SCAN_PARSE_PDF |
                 CL_SCAN_PARSE_SWF | CL_SCAN_PARSE_XMLDOCS | CL_SCAN_PARSE_MAIL |
                 CL_SCAN_PARSE_OLE2 | CL_SCAN_PARSE_HTML | CL_SCAN_PARSE_PE;
    opts.heuristic = 0;
    opts.mail = 0;
    opts.dev = 0;
    return opts;
}

// 封装 cl_scanfile 调用
static cl_error_t scan_file_wrapper(const char *filename, const char **virname,
                                     unsigned long int *scanned,
                                     const struct cl_engine *engine) {
    struct cl_scan_options opts = default_scan_options();
    return cl_scanfile(filename, virname, scanned, engine, &opts);
}
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// ClamAVEngine ClamAV 扫描引擎实现
type ClamAVEngine struct {
	engine      *C.struct_cl_engine
	mu          sync.RWMutex
	maxFileSize int64
	maxScanTime int
	initialized bool
}

// NewClamAVEngine 创建 ClamAV 引擎实例
func NewClamAVEngine(maxFileSize int64, maxScanTime int) *ClamAVEngine {
	if maxFileSize <= 0 {
		maxFileSize = DefaultMaxFileSize
	}
	if maxScanTime <= 0 {
		maxScanTime = DefaultMaxScanTime
	}
	return &ClamAVEngine{
		maxFileSize: maxFileSize,
		maxScanTime: maxScanTime,
	}
}

// Init 初始化 ClamAV 引擎
func (e *ClamAVEngine) Init() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 初始化 ClamAV 库
	ret := C.cl_init(C.CL_INIT_DEFAULT)
	if ret != C.CL_SUCCESS {
		return fmt.Errorf("cl_init failed: %s", C.GoString(C.cl_strerror(ret)))
	}

	// 创建引擎实例
	e.engine = C.cl_engine_new()
	if e.engine == nil {
		return fmt.Errorf("cl_engine_new failed")
	}

	// 设置引擎参数
	C.cl_engine_set_num(e.engine, C.CL_ENGINE_MAX_FILESIZE, C.longlong(e.maxFileSize))
	C.cl_engine_set_num(e.engine, C.CL_ENGINE_MAX_SCANSIZE, C.longlong(e.maxFileSize))
	C.cl_engine_set_num(e.engine, C.CL_ENGINE_MAX_SCANTIME, C.longlong(e.maxScanTime*1000))

	e.initialized = true
	return nil
}

// LoadDB 加载病毒数据库
func (e *ClamAVEngine) LoadDB(dbPath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return fmt.Errorf("engine not initialized")
	}

	cPath := C.CString(dbPath)
	defer C.free(unsafe.Pointer(cPath))

	var signo C.uint
	ret := C.cl_load(cPath, e.engine, &signo, C.CL_DB_STDOPT)
	if ret != C.CL_SUCCESS {
		return fmt.Errorf("cl_load failed for %s: %s", dbPath, C.GoString(C.cl_strerror(ret)))
	}

	// 编译引擎
	ret = C.cl_engine_compile(e.engine)
	if ret != C.CL_SUCCESS {
		return fmt.Errorf("cl_engine_compile failed: %s", C.GoString(C.cl_strerror(ret)))
	}

	return nil
}

// ScanFile 扫描单个文件
func (e *ClamAVEngine) ScanFile(path string) (*ScanResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.initialized {
		return nil, fmt.Errorf("engine not initialized")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var virusName *C.char
	var scanned C.ulong

	// 使用超时控制
	type scanRes struct {
		ret C.cl_error_t
	}
	resultCh := make(chan scanRes, 1)
	go func() {
		ret := C.scan_file_wrapper(cPath, &virusName, &scanned, e.engine)
		resultCh <- scanRes{ret: ret}
	}()

	select {
	case r := <-resultCh:
		if r.ret == C.CL_VIRUS {
			return &ScanResult{
				Infected:  true,
				VirusName: C.GoString(virusName),
			}, nil
		}
		if r.ret != C.CL_CLEAN {
			return nil, fmt.Errorf("cl_scanfile failed for %s: %s", path, C.GoString(C.cl_strerror(r.ret)))
		}
		return &ScanResult{Infected: false}, nil

	case <-time.After(time.Duration(e.maxScanTime) * time.Second):
		return nil, fmt.Errorf("scan timeout for %s after %ds", path, e.maxScanTime)
	}
}

// ReloadDB 重新加载病毒数据库（热更新）
func (e *ClamAVEngine) ReloadDB(dbPath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.initialized {
		return fmt.Errorf("engine not initialized")
	}

	// 创建新引擎
	newEngine := C.cl_engine_new()
	if newEngine == nil {
		return fmt.Errorf("cl_engine_new failed during reload")
	}

	// 设置引擎参数
	C.cl_engine_set_num(newEngine, C.CL_ENGINE_MAX_FILESIZE, C.longlong(e.maxFileSize))
	C.cl_engine_set_num(newEngine, C.CL_ENGINE_MAX_SCANSIZE, C.longlong(e.maxFileSize))
	C.cl_engine_set_num(newEngine, C.CL_ENGINE_MAX_SCANTIME, C.longlong(e.maxScanTime*1000))

	// 加载数据库到新引擎
	cPath := C.CString(dbPath)
	defer C.free(unsafe.Pointer(cPath))

	var signo C.uint
	ret := C.cl_load(cPath, newEngine, &signo, C.CL_DB_STDOPT)
	if ret != C.CL_SUCCESS {
		C.cl_engine_free(newEngine)
		return fmt.Errorf("cl_load failed during reload: %s", C.GoString(C.cl_strerror(ret)))
	}

	ret = C.cl_engine_compile(newEngine)
	if ret != C.CL_SUCCESS {
		C.cl_engine_free(newEngine)
		return fmt.Errorf("cl_engine_compile failed during reload: %s", C.GoString(C.cl_strerror(ret)))
	}

	// 替换旧引擎
	oldEngine := e.engine
	e.engine = newEngine

	// 释放旧引擎
	if oldEngine != nil {
		C.cl_engine_free(oldEngine)
	}

	return nil
}

// Close 关闭引擎，释放资源
func (e *ClamAVEngine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.engine != nil {
		C.cl_engine_free(e.engine)
		e.engine = nil
	}
	e.initialized = false
}
