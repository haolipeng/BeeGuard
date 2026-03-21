package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"scanner/engine"
	"scanner/log"
)

// Updater 病毒数据库更新器
type Updater struct {
	engine engine.Engine
	dbPath string
	logger *log.Logger
}

// New 创建更新器
func New(eng engine.Engine, dbPath string, logger *log.Logger) *Updater {
	return &Updater{
		engine: eng,
		dbPath: dbPath,
		logger: logger,
	}
}

// Update 下载并更新病毒数据库
func (u *Updater) Update(url, expectedSHA256 string) error {
	// 下载到临时文件
	tmpPath := filepath.Join(filepath.Dir(u.dbPath), ".db_update.tmp")
	if err := u.download(url, tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", err)
	}

	// SHA256 校验
	if expectedSHA256 != "" {
		actualSHA256, err := calcFileSHA256(tmpPath)
		if err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("sha256 calculation failed: %w", err)
		}
		if actualSHA256 != expectedSHA256 {
			os.Remove(tmpPath)
			return fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedSHA256, actualSHA256)
		}
		u.logger.Info("SHA256 verification passed")
	}

	// 原子替换：先备份旧文件，再重命名新文件
	backupPath := u.dbPath + ".bak"
	if _, err := os.Stat(u.dbPath); err == nil {
		os.Rename(u.dbPath, backupPath)
	}

	if err := os.Rename(tmpPath, u.dbPath); err != nil {
		// 回滚
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, u.dbPath)
		}
		return fmt.Errorf("rename failed: %w", err)
	}

	// 热更新引擎
	if err := u.engine.ReloadDB(u.dbPath); err != nil {
		// 回滚
		u.logger.Error("Engine reload failed, rolling back", "error", err)
		if _, err2 := os.Stat(backupPath); err2 == nil {
			os.Rename(backupPath, u.dbPath)
			u.engine.ReloadDB(u.dbPath)
		}
		return fmt.Errorf("engine reload failed: %w", err)
	}

	// 清理备份
	os.Remove(backupPath)
	u.logger.Info("Database updated and reloaded successfully")

	return nil
}

// download 下载文件
func (u *Updater) download(url, destPath string) error {
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file failed: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	u.logger.Info("Database downloaded", "size", written, "path", destPath)
	return nil
}

// calcFileSHA256 计算文件 SHA256
func calcFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
