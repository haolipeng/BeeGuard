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

# 兜底：强制杀死残留进程（避免 systemctl stop 因 KillMode=control-group 杀死本脚本）
if [ -d "/proc/${AGENT_PID}" ]; then
    kill -9 ${AGENT_PID} 2>/dev/null || true
    sleep 2
fi

# 禁用开机自启（在 stop 之前 disable，防止 stop 后被 Restart=always 重启）
systemctl disable ${PRODUCT_NAME} 2>/dev/null || true

# 停止服务（此时脚本已不在 agent cgroup 中，因为是通过 systemd-run 启动的）
systemctl stop ${PRODUCT_NAME} 2>/dev/null || true

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

// startUninstallScript 生成并启动脱离 agent cgroup 的卸载脚本。
// 使用 systemd-run --scope 在独立的 cgroup scope 中运行脚本，
// 避免 systemctl stop 的 KillMode=control-group 杀死脚本自身。
func startUninstallScript() error {
	pid := os.Getpid()
	scriptPath := fmt.Sprintf("/tmp/cloudsec-uninstall-%d.sh", pid)
	scriptContent := fmt.Sprintf(uninstallScript, pid)

	// 写入临时脚本文件
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0700); err != nil {
		return fmt.Errorf("failed to write uninstall script: %w", err)
	}

	zap.S().Infow("uninstall script created", "path", scriptPath)

	// 使用 systemd-run --scope 启动脚本，使其运行在独立的 cgroup scope 中，
	// 而非继承 agent 的 /system.slice/cloudsec-agent.service cgroup。
	// 这样 systemctl stop cloudsec-agent 时 KillMode=control-group 不会杀死脚本。
	cmd := exec.Command("systemd-run", "--scope", "--quiet", "/bin/bash", scriptPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		zap.S().Warnw("systemd-run failed, falling back to direct exec", "error", err)
		// 回退：直接启动脚本（适用于无 systemd 的环境）
		cmd = exec.Command("/bin/bash", scriptPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		if err := cmd.Start(); err != nil {
			os.Remove(scriptPath)
			return fmt.Errorf("failed to start uninstall script: %w", err)
		}
	}

	zap.S().Infow("uninstall script started", "script_pid", cmd.Process.Pid)

	// 释放进程资源，不等待脚本完成
	cmd.Process.Release()
	return nil
}
