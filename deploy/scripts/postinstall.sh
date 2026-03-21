#!/bin/bash
# postinstall.sh - 安装后启动服务

product_name="cloudsec-agent"
root_dir="/opt/cloudsec/agent"
agent_ctl="${root_dir}/bin/cloudsecctl"

error(){
    echo -e "\e[91m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[ERRO]\t$1\e[0m"
}
warn(){
    echo -e "\e[93m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[WARN]\t$1\e[0m"
}
info(){
    echo -e "\e[96m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[INFO]\t$1\e[0m"
}
succ(){
    echo -e "\e[92m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[SUCC]\t$1\e[0m"
}
expect(){
    eval "$1"
    rtc=$?
    if [ $rtc -ne 0 ]; then
        if [ -n "$2" ]; then
            eval "$2"
        fi
        error "when exec '$1', an unexpected error occurred, code: $rtc"
        exit $rtc
    fi
}

create_dirs() {
    info "creating runtime directories"
    mkdir -p "${root_dir}/data/agent"
    mkdir -p "${root_dir}/data/plugins/collector"
    mkdir -p "${root_dir}/data/plugins/baseline"
    mkdir -p "${root_dir}/data/plugins/detector"
    mkdir -p "${root_dir}/data/plugins/ebpf_base_detector"
    mkdir -p "${root_dir}/data/plugins/nids"
    mkdir -p "${root_dir}/data/plugins/scanner"
    mkdir -p "${root_dir}/logs/agent"
    mkdir -p "${root_dir}/logs/plugins/collector"
    mkdir -p "${root_dir}/logs/plugins/baseline"
    mkdir -p "${root_dir}/logs/plugins/detector"
    mkdir -p "${root_dir}/logs/plugins/ebpf_base_detector"
    mkdir -p "${root_dir}/logs/plugins/nids"
    mkdir -p "${root_dir}/logs/plugins/scanner"
}

# restore_config restores agent.yaml with default values when the file is
# missing. This can happen when a previous "rm -rf /opt/cloudsec/agent" removed
# the file without "dpkg --purge", causing dpkg to treat the deletion as
# intentional and skip restoring the conffile on reinstall.
restore_config() {
    local config_file="${root_dir}/agent.yaml"
    if [ -f "${config_file}" ]; then
        return
    fi
    warn "agent.yaml missing, restoring default config"
    cat > "${config_file}" <<'YAML'
server: "127.0.0.1:50051"
connect_timeout: 30
working_directory: "/opt/cloudsec/agent/data/agent"
plugins_directory: "/opt/cloudsec/agent/plugins"
log_directory: "/opt/cloudsec/agent/logs"
retry_max_count: 10
retry_interval: 5
YAML
    chmod 644 "${config_file}"
}

enable_service() {
    info "enabling agent service"
    expect "${agent_ctl} enable"
    succ "service enabled successfully"
}

set_env() {
    if [ -n "${SPECIFIED_SERVER}" ]; then
        ${agent_ctl} set --server="${SPECIFIED_SERVER}"
    fi
    if [ -n "${SPECIFIED_AGENT_ID}" ]; then
        ${agent_ctl} set --id="${SPECIFIED_AGENT_ID}"
    fi
}

reload_service() {
    ${agent_ctl} service-reload
}

start_agent() {
    ${agent_ctl} start
}

install() {
    create_dirs
    restore_config
    enable_service
    set_env
    reload_service
    start_agent
    succ "installation finished successfully"
}

upgrade() {
    create_dirs
    restore_config
    enable_service
    reload_service
    succ "upgrade finished successfully"
}

# DEB: $1=configure, $2="" (first install) or $2=<old-version> (upgrade)
# RPM: $1=1 (first install) or $1=2 (upgrade)
if [ "$1" = "configure" ] && [ -z "$2" ] || [ "$1" = "1" ]; then
    install
elif [ "$1" = "configure" ] && [ -n "$2" ] || [ "$1" = "2" ]; then
    upgrade
fi
