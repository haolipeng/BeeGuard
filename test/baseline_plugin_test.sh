#!/bin/bash
# 编译 baseline 插件
cd /home/work/goProject/src/company/agent/business_plugins/baseline
go build -o baseline main.go

# 准备插件目录
mkdir -p /tmp/plugin/baseline
cp baseline /tmp/plugin/baseline/baseline
chmod +x /tmp/plugin/baseline/baseline

# 运行测试
cd /home/work/goProject/src/company/agent/test
go run main.go
