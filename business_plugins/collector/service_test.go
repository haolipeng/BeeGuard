package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestGetServiceRuntimeInfo 诊断测试：验证 getServiceRuntimeInfo 函数对真实服务的输出
func TestGetServiceRuntimeInfo(t *testing.T) {
	// 先确认 systemctl 可用
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Fatalf("systemctl not found: %v", err)
	}

	testServices := []string{
		"docker.service",
		"ssh.service",
		"cron.service",
		"nonexistent-xxxx.service",
	}

	for _, name := range testServices {
		status, runUser, version := getServiceRuntimeInfo(name)
		t.Logf("Service: %-30s Status=%-20s RunUser=%-10s Version=%s",
			name, status, runUser, version)

		if name != "nonexistent-xxxx.service" && status == "unknown" {
			t.Errorf("Service %s: expected non-unknown status, got %q", name, status)
		}
	}
}

// TestGetServiceRuntimeInfo_ParseOutput 单元测试：验证 systemctl show 输出的解析逻辑
func TestGetServiceRuntimeInfo_ParseOutput(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		wantStatus string
	}{
		{
			name:       "active running",
			output:     "ExecMainPID=2130\nUser=\nActiveState=active\nSubState=running\n",
			wantStatus: "active(running)",
		},
		{
			name:       "inactive dead",
			output:     "ExecMainPID=0\nUser=\nActiveState=inactive\nSubState=dead\n",
			wantStatus: "inactive(dead)",
		},
		{
			name:       "failed",
			output:     "ActiveState=failed\nSubState=failed\n",
			wantStatus: "failed",
		},
		{
			name:       "empty output",
			output:     "",
			wantStatus: "unknown",
		},
		{
			name:       "no ActiveState line",
			output:     "User=root\nVersion=1.0\n",
			wantStatus: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 getServiceRuntimeInfo 的解析逻辑
			lines := strings.Split(tt.output, "\n")
			var activeState, subState string
			for _, line := range lines {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "ActiveState":
					activeState = value
				case "SubState":
					subState = value
				}
			}

			var status string
			if activeState != "" {
				if subState != "" && subState != activeState {
					status = activeState + "(" + subState + ")"
				} else {
					status = activeState
				}
			} else {
				status = "unknown"
			}

			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

// TestServiceFileParsingBug 验证服务文件解析使用 SplitN(_, "=", 2)，值中可含等号
func TestServiceFileParsingBug(t *testing.T) {
	lines := []struct {
		line    string
		wantKey string
		wantVal string
		skipped bool
	}{
		{"Type=notify", "Type", "notify", false},
		{"Restart=always", "Restart", "always", false},
		{"User=root", "User", "root", false},
		// 值中含等号：用 SplitN(_, "=", 2) 可正确解析
		{"ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock", "ExecStart", "/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock", false},
		{"ExecReload=/bin/kill -s HUP $MAINPID", "ExecReload", "/bin/kill -s HUP $MAINPID", false},
		{"EnvironmentFile=-/etc/default/ssh", "EnvironmentFile", "-/etc/default/ssh", false},
		{"ExecStart=/usr/sbin/sshd -D $SSHD_OPTS", "ExecStart", "/usr/sbin/sshd -D $SSHD_OPTS", false},
		{"# Comment", "", "", true},
		{"", "", "", true},
		{"[Service]", "", "", true},
	}

	for _, tt := range lines {
		parts := strings.SplitN(tt.line, "=", 2)
		skipped := len(parts) != 2
		if skipped != tt.skipped {
			t.Errorf("Line %q: skipped=%v, want %v", tt.line, skipped, tt.skipped)
			continue
		}
		if !skipped {
			key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if key != tt.wantKey || val != tt.wantVal {
				t.Errorf("Line %q: key=%q val=%q, want key=%q val=%q", tt.line, key, val, tt.wantKey, tt.wantVal)
			}
		}
	}
}

// TestServiceExecutablePathNotEmpty 验证系统服务采集得到的可执行文件路径（Command 首词）非空
// 对 docker.service、ssh.service、cron.service 等常见服务解析单元文件中的 ExecStart，检查是否成功解析出可执行路径
func TestServiceExecutablePathNotEmpty(t *testing.T) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		t.Skip("systemctl not found, skip")
	}

	serviceNames := []string{"docker.service", "ssh.service", "sshd.service", "cron.service"}
	for _, name := range serviceNames {
		t.Run(name, func(t *testing.T) {
			// 获取单元文件路径
			cmd := exec.Command("systemctl", "show", name, "--property=FragmentPath", "--value")
			out, err := cmd.Output()
			if err != nil {
				t.Skipf("systemctl show %s failed: %v", name, err)
			}
			path := strings.TrimSpace(string(out))
			if path == "" || path == "(null)" {
				t.Skipf("no FragmentPath for %s", name)
			}
			f, err := os.Open(path)
			if err != nil {
				t.Skipf("open %s: %v", path, err)
			}
			defer f.Close()

			// 与 service.go 中一致的解析方式：SplitN(line, "=", 2)，值中可含等号
			var command string
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key == "ExecStart" {
					command = value
					break
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatalf("read file: %v", err)
			}

			// 可执行路径 = Command 的第一个词（与 getVersionFromBinary 一致，去掉前缀 -@+!）
			exePath := ""
			if command != "" {
				parts := strings.Fields(command)
				if len(parts) > 0 {
					exePath = strings.TrimLeft(parts[0], "-@+!")
				}
			}

			if exePath == "" {
				t.Errorf("service %s: executable path is empty (Command=%q). ExecStart may be skipped due to multiple '=' in line (see TestServiceFileParsingBug)",
					name, command)
			} else {
				t.Logf("service %s: executable path = %s", name, exePath)
			}
		})
	}
}
