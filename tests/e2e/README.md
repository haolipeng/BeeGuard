# E2E 测试目录

本目录包含端到端（End-to-End）测试，用于测试完整的插件功能。

## 目录结构

```
tests/e2e/
├── baseline/          # Baseline 插件 E2E 测试
│   ├── main.go       # 测试主程序
│   ├── test.sh       # 自动化测试脚本
│   └── README.md     # 详细说明文档
└── collector/        # Collector 插件 E2E 测试
    ├── main.go       # 测试主程序
    ├── test.sh       # 自动化测试脚本
    └── README.md     # 详细说明文档（待创建）
```

## 快速开始

### 运行 Baseline 插件测试

```bash
cd tests/e2e/baseline
./test.sh
```

### 运行 Collector 插件测试

```bash
cd tests/e2e/collector
./test.sh
```

## 测试说明

这些 E2E 测试会：
1. 编译对应的插件
2. 将插件复制到 `/tmp/plugin/{插件名}/` 目录
3. 启动测试 agent
4. 加载插件并发送测试任务
5. 接收并打印插件返回的结果

## 注意事项

- 测试需要 root 权限（某些插件可能需要）
- 测试会创建临时文件和目录
- 测试程序运行一段时间后会自动退出，也可以按 Ctrl+C 提前退出

