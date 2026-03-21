#!/bin/bash
# preremove.sh - 卸载前停止/清理

product_name="cloudsec-agent"
root_dir="/opt/cloudsec/agent"
service_file="/etc/systemd/system/${product_name}.service"
agent_ctl="${root_dir}/bin/cloudsecctl"

error(){
    echo -e "\e[91m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[ERRO]\t$1\e[0m"
}
info(){
    echo -e "\e[96m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[INFO]\t$1\e[0m"
}
succ(){
    echo -e "\e[92m$(date "+%Y-%m-%d %H:%M:%S.%3N")\t[SUCC]\t$1\e[0m"
}

disable_service() {
    info "disabling agent service"
    ${agent_ctl} disable 2>/dev/null
    succ "service disabled"
}

stop_agent() {
    info "stopping agent"
    ${agent_ctl} stop 2>/dev/null
}

clean_dirs() {
    info "cleaning runtime files"
    rm -rf "${root_dir}/logs"
    rm -rf "${root_dir}/data"
    rm -f "${root_dir}/specified_env"
    rm -f "${root_dir}/plugin.sock"
    rm -f "${service_file}"
}

uninstall() {
    disable_service
    stop_agent
    clean_dirs
    succ "uninstall cleanup finished"
}

# DEB: $1=remove; RPM: $1=0
if [ "$1" = "remove" ] || [ "$1" = "0" ]; then
    uninstall
fi
