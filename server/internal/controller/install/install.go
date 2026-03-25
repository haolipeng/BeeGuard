package install

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/haolipeng/BeeGuard/server/internal/config"

	"github.com/gin-gonic/gin"
)

//go:embed install.sh.tpl
var installScriptTpl string

var installTmpl = template.Must(template.New("install.sh").Parse(installScriptTpl))

// Controller 一键安装控制器
type Controller struct{}

// NewController 创建安装控制器实例
func NewController() *Controller {
	return &Controller{}
}

// installScriptData 安装脚本模板数据
type installScriptData struct {
	BaseURL  string
	GRPCAddr string
}

// GetInstallScript 返回一键安装脚本
// GET /install.sh
func (ctrl *Controller) GetInstallScript(c *gin.Context) {
	cfg := config.AppConfig.Install

	// 确定服务器地址（用于 agent 连接 gRPC）
	grpcAddr := cfg.ServerAddr
	if grpcAddr == "" {
		// 从请求 Host 头提取 IP，拼接默认 gRPC 端口
		host := c.Request.Host
		ip, _, err := net.SplitHostPort(host)
		if err != nil {
			// Host 头可能不包含端口
			ip = host
		}
		grpcAddr = fmt.Sprintf("%s:%d", ip, config.AppConfig.Server.Port)
	}

	// 确定下载基地址
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	var buf bytes.Buffer
	if err := installTmpl.Execute(&buf, installScriptData{
		BaseURL:  baseURL,
		GRPCAddr: grpcAddr,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("生成安装脚本失败: %v", err)})
		return
	}

	c.Header("Content-Type", "text/x-shellscript")
	c.String(http.StatusOK, buf.String())
}

// DownloadPackage 下载安装包
// GET /api1/agent/download?type=deb|rpm&arch=amd64|arm64
func (ctrl *Controller) DownloadPackage(c *gin.Context) {
	pkgType := c.Query("type")
	arch := c.Query("arch")

	// 校验参数
	if pkgType != "deb" && pkgType != "rpm" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数 type 必须为 deb 或 rpm"})
		return
	}
	if arch != "amd64" && arch != "arm64" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数 arch 必须为 amd64 或 arm64"})
		return
	}

	packageDir := config.AppConfig.Install.PackageDir
	filePath, err := findPackage(packageDir, pkgType, arch)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.File(filePath)
}

// ListPackages 列出可用的安装包
// GET /api1/agent/packages
func (ctrl *Controller) ListPackages(c *gin.Context) {
	packageDir := config.AppConfig.Install.PackageDir

	entries, err := os.ReadDir(packageDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("读取安装包目录失败: %v", err)})
		return
	}

	type PackageInfo struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	var packages []PackageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".deb") && !strings.HasSuffix(name, ".rpm") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		packages = append(packages, PackageInfo{
			Name: name,
			Size: info.Size(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"package_dir": packageDir,
		"packages":    packages,
	})
}

// findPackage 在指定目录中查找匹配的安装包
func findPackage(dir, pkgType, arch string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("读取安装包目录失败: %v", err)
	}

	suffix := "." + pkgType

	// rpm 包中使用 x86_64 而非 amd64
	archKeywords := []string{arch}
	if pkgType == "rpm" && arch == "amd64" {
		archKeywords = append(archKeywords, "x86_64")
	}
	if pkgType == "rpm" && arch == "arm64" {
		archKeywords = append(archKeywords, "aarch64")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		for _, kw := range archKeywords {
			if strings.Contains(name, kw) {
				return filepath.Join(dir, name), nil
			}
		}
	}

	return "", fmt.Errorf("未找到匹配的安装包: type=%s, arch=%s", pkgType, arch)
}
