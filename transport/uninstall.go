package transport

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"go.uber.org/zap"
)

// uninstallScript 卸载脚本模板
// 参数: agent PID
const uninstallScript = `#!/bin/bash
# cloudsec-agent 远程卸载脚本（由 agent 进程生成）
set -e

AGENT_PID=%d
PRODUCT_NAME="cloudsec-agent"
INSTALL_DIR="/opt/cloudsec/agent"
TIMEOUT=60

# 等待 agent 进程退出
elapsed=0
while [ -d "/proc/${AGENT_PID}" ] && [ ${elapsed} -lt ${TIMEOUT} ]; do
    sleep 1
    elapsed=$((elapsed + 1))
done

# 兜底：确保进程停止
systemctl stop ${PRODUCT_NAME} 2>/dev/null || true

# 禁用开机自启
systemctl disable ${PRODUCT_NAME} 2>/dev/null || true

# 根据包管理器类型执行卸载
if command -v dpkg &>/dev/null && dpkg -l ${PRODUCT_NAME} &>/dev/null; then
    dpkg --purge ${PRODUCT_NAME} 2>/dev/null || true
elif command -v rpm &>/dev/null && rpm -q ${PRODUCT_NAME} &>/dev/null; then
    rpm -e ${PRODUCT_NAME} 2>/dev/null || true
fi

# 清理残留文件
rm -rf "${INSTALL_DIR}"

# 删除脚本自身
rm -f "$0"
`

// startUninstallScript 生成并启动脱离 agent 进程组的卸载脚本
func startUninstallScript() error {
	pid := os.Getpid()
	scriptPath := fmt.Sprintf("/tmp/cloudsec-uninstall-%d.sh", pid)
	scriptContent := fmt.Sprintf(uninstallScript, pid)

	// 写入临时脚本文件
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0700); err != nil {
		return fmt.Errorf("failed to write uninstall script: %w", err)
	}

	zap.S().Infow("uninstall script created", "path", scriptPath)

	// 以新会话启动脚本，脱离 agent 进程组
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		// 启动失败时清理脚本
		os.Remove(scriptPath)
		return fmt.Errorf("failed to start uninstall script: %w", err)
	}

	zap.S().Infow("uninstall script started", "script_pid", cmd.Process.Pid)

	// 释放进程资源，不等待脚本完成
	cmd.Process.Release()
	return nil
}
