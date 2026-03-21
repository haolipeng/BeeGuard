#!/bin/bash
# 从 hids.bpf.dev.c 生成简洁注释版 hids.bpf.c
# 删除所有 // 注释（保留 SPDX）

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$REPO_ROOT/business_plugins/ebpf_base_detector/ebpf/bpf/hids.bpf.dev.c"
DST="$REPO_ROOT/business_plugins/ebpf_base_detector/ebpf/bpf/hids.bpf.c"

if [ ! -f "$SRC" ]; then
    echo "skip: $SRC not found"
    exit 0
fi

sed -e '1{/SPDX/b}' -e '/^[[:space:]]*\/\//d' -e 's/[[:space:]]*\/\/.*$//' "$SRC" > "$DST"
echo "generated: $DST"
