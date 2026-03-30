package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/haolipeng/BeeGuard/server/internal/analysis"
	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/db"
	"github.com/haolipeng/BeeGuard/server/internal/geoip"
	"github.com/haolipeng/BeeGuard/server/internal/grpc/handler"
	"github.com/haolipeng/BeeGuard/server/internal/log"
	"github.com/haolipeng/BeeGuard/server/internal/models"
	"github.com/haolipeng/BeeGuard/server/internal/mysql"
	"github.com/haolipeng/BeeGuard/server/internal/router"
	"github.com/haolipeng/BeeGuard/server/internal/vuln"
	"github.com/haolipeng/BeeGuard/server/proto"
	"github.com/haolipeng/BeeGuard/server/web_console"

	"github.com/gin-gonic/gin"
)

var (
	configFile = flag.String("config", "./conf/server.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 设置全局配置变量（供其他包使用）
	config.AppConfig = cfg

	// 初始化日志
	if err := log.Init(&cfg.Log); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Infof("配置加载成功: grpc_port=%d, http_port=%d, log_level=%s",
		cfg.Server.Port, cfg.Server.HttpPort, cfg.Log.Level)

	// 初始化 GeoIP 服务
	geoIPService, err := geoip.NewService(
		cfg.GeoIP.Enabled,
		cfg.GeoIP.DBPath,
		cfg.GeoIP.CacheTTL,
		cfg.GeoIP.MaxCacheSize,
	)
	if err != nil {
		log.Fatalf("Failed to initialize GeoIP service: %v", err)
	}
	defer geoIPService.Close()

	// 检查是否跳过数据库初始化
	if os.Getenv("SKIP_DB_INIT") != "true" {
		// 初始化 GORM 数据库连接
		if err := db.Init(&cfg.Database); err != nil {
			log.Fatalf("数据库初始化失败: %v", err)
		}
		defer db.Close()

		// 设置 mysql.DB 共享连接（供远程新增的 controller 使用）
		mysql.SetDB(db.GetDB())

		// 自动迁移数据库表并初始化菜单
		if err := models.AutoMigrate(); err != nil {
			log.Fatalf("数据库迁移失败: %v", err)
		}
	} else {
		log.Infof("跳过数据库初始化（SKIP_DB_INIT=true）")
	}

	// 创建 TCP 监听
	grpcAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	// 初始化漏洞数据库管理器和调度器
	var vulnScheduler *vuln.Scheduler
	if cfg.Vuln.Enabled {
		vulnDBManager := vuln.NewDBManager(&cfg.Vuln)
		if err := vulnDBManager.Init(context.Background()); err != nil {
			// 漏洞模块初始化失败不阻塞服务启动，降级为警告
			log.Warnf("[VulnDB] 漏洞数据库初始化失败（服务继续运行）: %v", err)
		} else {
			defer vulnDBManager.Close()

			// 启动漏洞匹配调度器
			vulnScheduler = vuln.NewScheduler(&cfg.Vuln, vulnDBManager)
			vulnScheduler.Start()
		}
	}

	// 初始化AI分析模块
	if cfg.Analysis.Enabled {
		if err := analysis.Init(analysis.Config{
			OllamaURL:        cfg.Analysis.OllamaURL,
			OllamaModel:      cfg.Analysis.OllamaModel,
			CacheDir:         cfg.Analysis.CacheDir,
			ReportDir:        cfg.Analysis.ReportDir,
			ScheduleInterval: time.Duration(cfg.Analysis.ScheduleMinutes) * time.Minute,
			AutoStart:        true,
		}); err != nil {
			log.Warnf("[Analysis] AI分析模块初始化失败（服务继续运行）: %v", err)
		}
		defer analysis.Stop()
	}

	// 创建 gRPC Server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.Server.MaxRecvMsgSize*1024*1024),
		grpc.MaxSendMsgSize(cfg.Server.MaxSendMsgSize*1024*1024),
	)

	// 注册 Transfer 服务
	transferServer := handler.NewTransferServer(geoIPService)
	proto.RegisterTransferServer(grpcServer, transferServer)

	// 设置 HTTP 路由（包含业务路由和 gRPC 管理路由）
	httpRouter := router.SetupRouter(transferServer)

	// 前端静态文件服务 (Go embed)
	staticFS, err := fs.Sub(web_console.StaticFS, "dist")
	if err != nil {
		log.Warnf("[HTTP] 前端静态文件加载失败: %v", err)
	} else {
		httpRouter.StaticFS("/ui", http.FS(staticFS))
		// SPA fallback：非 API 路径返回 index.html
		httpRouter.NoRoute(func(c *gin.Context) {
			if !strings.HasPrefix(c.Request.URL.Path, "/api1/") &&
				!strings.HasPrefix(c.Request.URL.Path, "/health") &&
				!strings.HasPrefix(c.Request.URL.Path, "/install.sh") {
				c.FileFromFS("index.html", http.FS(staticFS))
				return
			}
			c.JSON(404, gin.H{"error": "not found"})
		})
	}

	// 启动 HTTP API 服务（使用配置文件中的端口）
	httpAddr := fmt.Sprintf(":%d", cfg.Server.HttpPort)
	go func() {
		log.Infof("[HTTP] HTTP API Server 启动，监听端口 %s", httpAddr)
		if err := httpRouter.Run(httpAddr); err != nil {
			log.Errorf("HTTP API Server 启动失败: %v", err)
		}
	}()

	// 优雅关闭
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	shutdownDone := make(chan struct{})

	go func() {
		<-sigCh
		log.Infof("收到关闭信号，正在优雅关闭...")

		// 停止漏洞匹配调度器
		if vulnScheduler != nil {
			vulnScheduler.Stop()
		}

		// 停止AI分析模块
		analysis.Stop()

		// GracefulStop 会等待所有活跃的流式 RPC handler 完成
		gracefulDone := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(gracefulDone)
		}()

		select {
		case <-gracefulDone:
			log.Infof("gRPC Server 优雅关闭完成")
		case <-time.After(5 * time.Second):
			log.Warnf("gRPC 优雅关闭超时(5s)，强制关闭连接")
			grpcServer.Stop() //强制关闭连接
		}

		transferServer.Stop() // drain dispatcher + flush writers
		close(shutdownDone)
	}()

	// 启动 gRPC 服务
	log.Infof("gRPC Server 启动，监听端口 %s", grpcAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC Server 运行失败: %v", err)
	}

	// 等待 shutdown goroutine 完成（transferServer flush 完毕），再执行 defer 链
	<-shutdownDone
	log.Infof("gRPC Server 已关闭")
}
