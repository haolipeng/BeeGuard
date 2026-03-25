package vuln

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry/remote"

	trivydb "github.com/aquasecurity/trivy-db/pkg/db"
	"github.com/aquasecurity/trivy-db/pkg/metadata"
	"github.com/aquasecurity/trivy-db/pkg/types"

	"github.com/haolipeng/BeeGuard/server/internal/config"
	"github.com/haolipeng/BeeGuard/server/internal/log"
)

// DBManager 管理 Trivy 漏洞数据库的生命周期（下载、初始化、更新、查询）
type DBManager struct {
	cfg   *config.VulnConfig
	dbDir string // 实际 DB 文件所在目录（cfg.DBDir/db）
	dbc   trivydb.Config
	metaC metadata.Client
	mu    sync.RWMutex
	ready bool
}

// NewDBManager 创建漏洞数据库管理器
func NewDBManager(cfg *config.VulnConfig) *DBManager {
	dbDir := filepath.Join(cfg.DBDir, "db")
	return &DBManager{
		cfg:   cfg,
		dbDir: dbDir,
		dbc:   trivydb.Config{},
		metaC: metadata.NewClient(dbDir),
	}
}

// Init 初始化漏洞数据库：如果本地 DB 不存在则下载，然后打开 BoltDB 连接
func (m *DBManager) Init(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dbPath := trivydb.Path(m.dbDir)

	// 检查 DB 文件是否存在，不存在则下载
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Infof("[VulnDB] 漏洞数据库不存在，开始下载: %s", m.cfg.DBRepository)
		if err := m.download(ctx); err != nil {
			return fmt.Errorf("下载漏洞数据库失败: %w", err)
		}
	}

	// 打开 BoltDB
	if err := trivydb.Init(m.dbDir); err != nil {
		return fmt.Errorf("打开漏洞数据库失败: %w", err)
	}

	m.ready = true
	log.Infof("[VulnDB] 漏洞数据库初始化成功: %s", dbPath)
	return nil
}

// Close 关闭漏洞数据库
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ready = false
	return trivydb.Close()
}

// IsReady 检查数据库是否就绪
func (m *DBManager) IsReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ready
}

// NeedsUpdate 检查漏洞数据库是否需要更新
func (m *DBManager) NeedsUpdate() bool {
	meta, err := m.metaC.Get()
	if err != nil {
		log.Warnf("[VulnDB] 读取元数据失败，需要更新: %v", err)
		return true
	}

	// 检查 schema 版本
	if meta.Version != trivydb.SchemaVersion {
		log.Infof("[VulnDB] Schema 版本不匹配 (本地=%d, 需要=%d)，需要更新",
			meta.Version, trivydb.SchemaVersion)
		return true
	}

	// 检查更新时间
	updateInterval := time.Duration(m.cfg.UpdateInterval) * time.Hour
	if time.Since(meta.UpdatedAt) > updateInterval {
		log.Infof("[VulnDB] 数据库已过期 (上次更新: %s)，需要更新",
			meta.UpdatedAt.Format(time.RFC3339))
		return true
	}

	return false
}

// Update 更新漏洞数据库：关闭当前连接 → 下载新 DB → 重新打开
func (m *DBManager) Update(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 关闭现有连接
	if m.ready {
		if err := trivydb.Close(); err != nil {
			log.Warnf("[VulnDB] 关闭旧数据库连接失败: %v", err)
		}
		m.ready = false
	}

	// 下载新数据库
	log.Infof("[VulnDB] 开始更新漏洞数据库...")
	if err := m.download(ctx); err != nil {
		return fmt.Errorf("下载漏洞数据库失败: %w", err)
	}

	// 重新打开
	if err := trivydb.Init(m.dbDir); err != nil {
		return fmt.Errorf("重新打开漏洞数据库失败: %w", err)
	}

	m.ready = true
	log.Infof("[VulnDB] 漏洞数据库更新成功")
	return nil
}

// GetAdvisories 查询指定平台和软件包的漏洞 Advisory 列表
// source 为平台标识，如 "debian 12", "ubuntu 22.04", "alpine 3.18"
func (m *DBManager) GetAdvisories(source, pkgName string) ([]types.Advisory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.ready {
		return nil, fmt.Errorf("漏洞数据库未就绪")
	}

	advisories, err := m.dbc.GetAdvisories(source, pkgName)
	if err != nil {
		return nil, fmt.Errorf("查询 advisory 失败 (source=%s, pkg=%s): %w",
			source, pkgName, err)
	}
	return advisories, nil
}

// GetVulnerability 根据 CVE ID 获取漏洞详情
func (m *DBManager) GetVulnerability(cveID string) (types.Vulnerability, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.ready {
		return types.Vulnerability{}, fmt.Errorf("漏洞数据库未就绪")
	}

	vuln, err := m.dbc.GetVulnerability(cveID)
	if err != nil {
		return types.Vulnerability{}, fmt.Errorf("查询漏洞详情失败 (cve=%s): %w",
			cveID, err)
	}
	return vuln, nil
}

// GetMetadata 获取数据库元数据
func (m *DBManager) GetMetadata() (metadata.Metadata, error) {
	return m.metaC.Get()
}

// download 从 OCI 仓库下载 trivy-db 数据库文件
func (m *DBManager) download(ctx context.Context) error {
	ref := m.cfg.DBRepository
	repoRef, tag := parseReference(ref)

	repo, err := remote.NewRepository(repoRef)
	if err != nil {
		return fmt.Errorf("创建 OCI 仓库客户端失败: %w", err)
	}
	repo.PlainHTTP = false

	// 解析 tag 获取 manifest descriptor
	manifestDesc, err := repo.Resolve(ctx, tag)
	if err != nil {
		return fmt.Errorf("解析 tag 失败 (ref=%s): %w", ref, err)
	}

	log.Infof("[VulnDB] 已解析 manifest: mediaType=%s, digest=%s",
		manifestDesc.MediaType, manifestDesc.Digest)

	// 获取 manifest 内容
	rc, err := repo.Fetch(ctx, manifestDesc)
	if err != nil {
		return fmt.Errorf("获取 manifest 失败: %w", err)
	}

	manifestBytes, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return fmt.Errorf("读取 manifest 失败: %w", err)
	}

	// 解析 manifest 获取第一个 layer（即 trivy.db tar.gz）
	layerDesc, err := parseFirstLayer(manifestBytes)
	if err != nil {
		return fmt.Errorf("解析 manifest layers 失败: %w", err)
	}

	log.Infof("[VulnDB] 开始下载 DB layer: mediaType=%s, size=%d",
		layerDesc.MediaType, layerDesc.Size)

	// 下载 layer
	layerRC, err := repo.Fetch(ctx, layerDesc)
	if err != nil {
		return fmt.Errorf("下载 layer 失败: %w", err)
	}
	defer layerRC.Close()

	// 解压 tar.gz 到本地目录
	if err := m.extractDB(layerRC); err != nil {
		return fmt.Errorf("解压数据库失败: %w", err)
	}

	log.Infof("[VulnDB] 漏洞数据库下载完成: %s", m.dbDir)
	return nil
}

// extractDB 从 tar.gz 流中提取 trivy.db 和 metadata.json
func (m *DBManager) extractDB(r io.Reader) error {
	if err := os.MkdirAll(m.dbDir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip reader 创建失败: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	extracted := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar 读取失败: %w", err)
		}

		// 只提取 trivy.db 和 metadata.json
		name := filepath.Base(header.Name)
		if name != "trivy.db" && name != "metadata.json" {
			continue
		}

		targetPath := filepath.Join(m.dbDir, name)
		f, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("创建文件失败 (%s): %w", targetPath, err)
		}

		if _, err := io.Copy(f, tr); err != nil {
			f.Close()
			return fmt.Errorf("写入文件失败 (%s): %w", targetPath, err)
		}
		f.Close()

		extracted++
		log.Infof("[VulnDB] 已提取: %s (%d bytes)", name, header.Size)
	}

	if extracted == 0 {
		return fmt.Errorf("tar.gz 中未找到 trivy.db 或 metadata.json")
	}
	return nil
}

// parseReference 解析 OCI 引用字符串，分离仓库路径和标签
// 输入: "ghcr.io/aquasecurity/trivy-db:2"
// 输出: ("ghcr.io/aquasecurity/trivy-db", "2")
func parseReference(ref string) (string, string) {
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		after := ref[idx+1:]
		if !strings.Contains(after, "/") {
			return ref[:idx], after
		}
	}
	return ref, "latest"
}

// parseFirstLayer 从 OCI manifest JSON 中解析第一个 layer 的 descriptor
func parseFirstLayer(data []byte) (ocispec.Descriptor, error) {
	var manifest ocispec.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("JSON 解析失败: %w", err)
	}
	if len(manifest.Layers) == 0 {
		return ocispec.Descriptor{}, fmt.Errorf("manifest 中无 layers")
	}
	return manifest.Layers[0], nil
}
