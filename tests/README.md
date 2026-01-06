# 测试目录

本目录包含项目的所有测试代码。

## 目录结构

```
tests/
├── e2e/              # 端到端测试（End-to-End Tests）
│   ├── baseline/     # Baseline 插件 E2E 测试
│   ├── collector/    # Collector 插件 E2E 测试
│   └── README.md     # E2E 测试说明
└── README.md         # 本文件
```

## 测试类型说明

### E2E 测试（End-to-End Tests）

E2E 测试位于 `tests/e2e/` 目录，用于测试完整的插件功能，包括：
- 插件编译
- 插件加载
- 任务发送和接收
- 数据验证

这些测试是独立的可执行程序，模拟完整的 agent 运行环境。

## 运行测试

### 使用 Makefile（推荐）

```bash
# 运行所有 E2E 测试
make test-e2e

# 运行特定插件的 E2E 测试
make test-e2e-baseline
make test-e2e-collector

# 运行所有测试（单元测试 + E2E 测试）
make test-all
```

### 手动运行

```bash
# Baseline 插件测试
cd tests/e2e/baseline
./test.sh

# Collector 插件测试
cd tests/e2e/collector
./test.sh
```

## 未来计划

- [ ] 添加单元测试（`*_test.go` 文件）
- [ ] 添加集成测试（使用 build tag）
- [ ] 添加测试覆盖率报告
- [ ] 集成到 CI/CD 流程

## 相关文档

- [E2E 测试详细说明](e2e/README.md)
- [测试最佳实践指南](../TESTING.md)

