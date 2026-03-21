#!/bin/bash
# 构建无 D-Bus/DPDK 依赖的静态 libpcap
# 生成的 libpcap.a 仅依赖 glibc，适合静态链接到 nids 插件

set -e

LIBPCAP_VERSION="1.10.1"
LIBPCAP_URL="https://www.tcpdump.org/release/libpcap-${LIBPCAP_VERSION}.tar.gz"
BUILD_DIR="${1:-$(dirname "$0")/../build/libpcap-static}"
BUILD_DIR=$(cd "$(dirname "$BUILD_DIR")" && pwd)/$(basename "$BUILD_DIR")
SRC_DIR="/tmp/libpcap-${LIBPCAP_VERSION}"

# 如果已构建则跳过
if [ -f "${BUILD_DIR}/lib/libpcap.a" ]; then
    echo "Static libpcap already built at ${BUILD_DIR}/lib/libpcap.a"
    exit 0
fi

echo "Building static libpcap ${LIBPCAP_VERSION}..."
mkdir -p "${BUILD_DIR}"

# 下载源码（如果不存在）
if [ ! -d "${SRC_DIR}" ]; then
    echo "Downloading libpcap ${LIBPCAP_VERSION}..."
    wget -q -O "/tmp/libpcap-${LIBPCAP_VERSION}.tar.gz" "${LIBPCAP_URL}"
    tar xzf "/tmp/libpcap-${LIBPCAP_VERSION}.tar.gz" -C /tmp
fi

# 配置：禁用所有不需要的功能，避免额外依赖
cd "${SRC_DIR}"
make clean 2>/dev/null || true
./configure \
    --disable-dbus \
    --disable-rdma \
    --disable-bluetooth \
    --disable-usb \
    --without-dpdk \
    --prefix="${BUILD_DIR}" \
    --quiet

# 编译
make -j"$(nproc)" 2>&1 | grep -v '^$'

# 安装头文件和静态库
make install 2>/dev/null

echo "Static libpcap built successfully:"
echo "  Library: ${BUILD_DIR}/lib/libpcap.a"
echo "  Headers: ${BUILD_DIR}/include/"
