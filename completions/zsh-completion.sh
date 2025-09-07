#compdef nat-manager
# Zsh completion for nat-manager
# Copy to /usr/local/share/zsh/site-functions/_nat-manager

_nat_manager() {
    local context state state_descr line
    typeset -A opt_args

    # Global options
    local -a global_options=(
        '(-h --help)'{-h,--help}'[Show help message]'
        '(-v --verbose)'{-v,--verbose}'[Enable verbose output]'
        '--config[Configuration file path]:config file:_files'
        '--config-path[Configuration directory path]:config path:_directories'
        '--version[Show version information]'
    )

    # Main command specification
    _arguments -C \
        $global_options \
        '1: :_nat_manager_commands' \
        '*:: :->args' \
        && ret=0

    case "$state" in
        args)
            case $words[1] in
                start)
                    _arguments \
                        $global_options \
                        '(-e --external)'{-e,--external}'[External interface]:interface:_nat_manager_external_interfaces' \
                        '(-i --internal)'{-i,--internal}'[Internal interface]:interface:_nat_manager_internal_interfaces' \
                        '(-n --network)'{-n,--network}'[Internal network]:network:_nat_manager_networks' \
                        '--dhcp-start[DHCP range start]:ip address:_nat_manager_ip_addresses' \
                        '--dhcp-end[DHCP range end]:ip address:_nat_manager_ip_addresses' \
                        '--dns[DNS servers]:dns servers:_nat_manager_dns_servers' \
                        '(-h --help)'{-h,--help}'[Show help for start command]' \
                        && ret=0
                    ;;
                stop)
                    _arguments \
                        $global_options \
                        '(-f --force)'{-f,--force}'[Force stop even if some operations fail]' \
                        '(-h --help)'{-h,--help}'[Show help for stop command]' \
                        && ret=0
                    ;;
                status)
                    _arguments \
                        $global_options \
                        '--json[Output status in JSON format]' \
                        '(-h --help)'{-h,--help}'[Show help for status command]' \
                        && ret=0
                    ;;
                interfaces)
                    _arguments \
                        $global_options \
                        '(-a --all)'{-a,--all}'[Show all interfaces including inactive]' \
                        '(-t --type)'{-t,--type}'[Filter by interface type]:type:_nat_manager_interface_types' \
                        '(-h --help)'{-h,--help}'[Show help for interfaces command]' \
                        && ret=0
                    ;;
                monitor)
                    _arguments \
                        $global_options \
                        '(-i --interval)'{-i,--interval}'[Refresh interval]:interval:_nat_manager_intervals' \
                        '(-m --max)'{-m,--max}'[Maximum connections to display]:number:_nat_manager_connection_limits' \
                        '(-d --devices)'{-d,--devices}'[Show connected devices]' \
                        '(-f --follow)'{-f,--follow}'[Continuous monitoring mode]' \
                        '(-h --help)'{-h,--help}'[Show help for monitor command]' \
                        && ret=0
                    ;;
                help)
                    _arguments \
                        '1: :_nat_manager_commands' \
                        && ret=0
                    ;;
            esac
            ;;
    esac

    return ret
}

# Command completion
_nat_manager_commands() {
    local -a commands=(
        'start:Start NAT service'
        'stop:Stop NAT service'
        'status:Show NAT service status'
        'interfaces:List available network interfaces'
        'monitor:Monitor NAT traffic and connections'
        'help:Show help information'
        'version:Show version information'
    )
    _describe 'commands' commands
}

# External interface completion
_nat_manager_external_interfaces() {
    local -a interfaces
    
    # Try to get interfaces from nat-manager if available
    if command -v nat-manager >/dev/null 2>&1; then
        interfaces=(${(f)"$(nat-manager interfaces 2>/dev/null | awk 'NR>2 && ($2=="Ethernet" || $2=="WiFi") && $4=="Up" {print $1":"$2" ("$3")"}')"})
    fi
    
    # Fallback to common interface names
    if [[ ${#interfaces} -eq 0 ]]; then
        interfaces=(
            'en0:Ethernet (Primary)'
            'en1:WiFi'
            'en2:Ethernet (Secondary)'
        )
    fi
    
    _describe 'external interfaces' interfaces
}

# Internal interface completion
_nat_manager_internal_interfaces() {
    local -a interfaces=(
        'bridge100:Bridge Interface 100'
        'bridge101:Bridge Interface 101'
        'bridge102:Bridge Interface 102'
        'utun0:User Tunnel 0'
        'utun1:User Tunnel 1'
    )
    _describe 'internal interfaces' interfaces
}

# Network prefix completion
_nat_manager_networks() {
    local -a networks=(
        '192.168.1:Private Class C (192.168.1.0/24)'
        '192.168.100:Private Class C (192.168.100.0/24)'
        '10.0.1:Private Class A (10.0.1.0/24)'
        '172.16.1:Private Class B (172.16.1.0/24)'
    )
    _describe 'network prefixes' networks
}

# IP address completion
_nat_manager_ip_addresses() {
    local -a addresses=(
        '192.168.100.100'
        '192.168.100.200'
        '10.0.1.100'
        '10.0.1.200'
    )
    _describe 'ip addresses' addresses
}

# DNS servers completion
_nat_manager_dns_servers() {
    local -a dns_servers=(
        '8.8.8.8,8.8.4.4:Google DNS'
        '1.1.1.1,1.0.0.1:Cloudflare DNS'
        '208.67.222.222,208.67.220.220:OpenDNS'
        '9.9.9.9,149.112.112.112:Quad9 DNS'
    )
    _describe 'dns servers' dns_servers
}

# Interface type completion
_nat_manager_interface_types() {
    local -a types=(
        'ethernet:Ethernet interfaces'
        'bridge:Bridge interfaces'
        'vpn:VPN tunnel interfaces'
        'tunnel:Tunnel interfaces'
        'loopback:Loopback interfaces'
    )
    _describe 'interface types' types
}

# Time interval completion
_nat_manager_intervals() {
    local -a intervals=(
        '1s:1 second'
        '2s:2 seconds'
        '5s:5 seconds'
        '10s:10 seconds'
        '30s:30 seconds'
        '1m:1 minute'
        '2m:2 minutes'
        '5m:5 minutes'
    )
    _describe 'refresh intervals' intervals
}

# Connection limit completion
_nat_manager_connection_limits() {
    local -a limits=(
        '10:10 connections'
        '20:20 connections'
        '50:50 connections'
        '100:100 connections'
        '200:200 connections'
    )
    _describe 'connection limits' limits
}

# File path completion for configuration files
_nat_manager_config_files() {
    _alternative \
        'files:configuration files:_files -g "*.yaml *.yml *.json"' \
        'directories:configuration directories:_directories'
}

# Set completion function
_nat_manager "$@"