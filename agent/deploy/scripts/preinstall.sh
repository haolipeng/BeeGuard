#!/bin/bash
# preinstall.sh - 安装前检查

error(){
    echo -e "\e[91m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[ERRO]\t$1\e[0m"
}

if command -v systemctl >/dev/null 2>&1; then
    exit 0
else
    error "systemctl is required but not found. Only systemd is supported."
    exit 1
fi
