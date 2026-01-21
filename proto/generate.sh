#!/bin/bash
# 生成 protobuf 代码的脚本
# 需要先安装: protoc 和 protoc-gen-gogofaster
# 使用 plugins=grpc 参数同时生成消息和 gRPC 服务代码（与 Elkeid 一致）

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装，请先安装 Protocol Buffers 编译器"
    echo "Ubuntu/Debian: sudo apt-get install protobuf-compiler"
    echo "macOS: brew install protobuf"
    exit 1
fi

# 检查 protoc-gen-gogofaster 是否安装
if ! command -v protoc-gen-gogofaster &> /dev/null; then
    echo "错误: protoc-gen-gogofaster 未安装"
    echo "安装命令: go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest"
    exit 1
fi

# 生成 protobuf 消息代码 + gRPC 服务代码（单个文件）
protoc --gogofaster_out=plugins=grpc:. grpc.proto

echo "生成完成！"
