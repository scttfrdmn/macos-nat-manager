# Bash completion for nat-manager
# Source this file or copy to /usr/local/share/bash-completion/completions/nat-manager

_nat_manager_complete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="start stop status interfaces monitor help version"
    
    # Global options
    local global_opts="--config --verbose --help --version"

    # Get the current command being completed
    local cmd=""
    local i=1
    while [[ $i -lt $COMP_CWORD ]]; do
        case "${COMP_WORDS[i]}" in
            start|stop|status|interfaces|monitor)
                cmd="${COMP_WORDS[i]}"
                break
                ;;
        esac
        ((i++))
    done

    # If no command yet, complete main commands and global options
    if [[ -z "$cmd" ]]; then
        case "$cur" in
            -*)
                COMPREPLY=( $(compgen -W "$global_opts" -- "$cur") )
                return 0
                ;;
            *)
                COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
                return 0
                ;;
        esac
    fi

    # Command-specific completion
    case "$cmd" in
        start)
            case "$prev" in
                --external|-e)
                    # Complete with available interfaces
                    local interfaces
                    if command -v nat-manager >/dev/null 2>&1; then
                        interfaces=$(nat-manager interfaces 2>/dev/null | awk 'NR>2 {print $1}' | grep -E '^en|^bridge|^utun' || echo "en0 en1 bridge100")
                    else
                        interfaces="en0 en1 bridge100"
                    fi
                    COMPREPLY=( $(compgen -W "$interfaces" -- "$cur") )
                    return 0
                    ;;
                --internal|-i)
                    # Complete with bridge interfaces
                    COMPREPLY=( $(compgen -W "bridge100 bridge101 bridge102" -- "$cur") )
                    return 0
                    ;;
                --network|-n)
                    # Common network prefixes
                    COMPREPLY=( $(compgen -W "192.168.1 192.168.100 10.0.1 172.16.1" -- "$cur") )
                    return 0
                    ;;
                --dhcp-start)
                    COMPREPLY=( $(compgen -W "192.168.100.100 10.0.1.100" -- "$cur") )
                    return 0
                    ;;
                --dhcp-end)
                    COMPREPLY=( $(compgen -W "192.168.100.200 10.0.1.200" -- "$cur") )
                    return 0
                    ;;
                --dns)
                    COMPREPLY=( $(compgen -W "8.8.8.8,8.8.4.4 1.1.1.1,1.0.0.1 208.67.222.222,208.67.220.220" -- "$cur") )
                    return 0
                    ;;
                *)
                    case "$cur" in
                        -*)
                            local start_opts="--external --internal --network --dhcp-start --dhcp-end --dns --help"
                            COMPREPLY=( $(compgen -W "$start_opts" -- "$cur") )
                            return 0
                            ;;
                    esac
                    ;;
            esac
            ;;
        stop)
            case "$cur" in
                -*)
                    local stop_opts="--force --help"
                    COMPREPLY=( $(compgen -W "$stop_opts" -- "$cur") )
                    return 0
                    ;;
            esac
            ;;
        status)
            case "$cur" in
                -*)
                    local status_opts="--json --help"
                    COMPREPLY=( $(compgen -W "$status_opts" -- "$cur") )
                    return 0
                    ;;
            esac
            ;;
        interfaces)
            case "$prev" in
                --type|-t)
                    COMPREPLY=( $(compgen -W "ethernet bridge vpn tunnel loopback" -- "$cur") )
                    return 0
                    ;;
                *)
                    case "$cur" in
                        -*)
                            local interfaces_opts="--all --type --help"
                            COMPREPLY=( $(compgen -W "$interfaces_opts" -- "$cur") )
                            return 0
                            ;;
                    esac
                    ;;
            esac
            ;;
        monitor)
            case "$prev" in
                --interval|-i)
                    COMPREPLY=( $(compgen -W "1s 2s 5s 10s 30s 1m" -- "$cur") )
                    return 0
                    ;;
                --max|-m)
                    COMPREPLY=( $(compgen -W "10 20 50 100" -- "$cur") )
                    return 0
                    ;;
                *)
                    case "$cur" in
                        -*)
                            local monitor_opts="--interval --max --devices --follow --help"
                            COMPREPLY=( $(compgen -W "$monitor_opts" -- "$cur") )
                            return 0
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac

    return 0
}

# Register completion
complete -F _nat_manager_complete nat-manager

# Additional helpers for common tasks
_nat_manager_interfaces() {
    if command -v nat-manager >/dev/null 2>&1; then
        nat-manager interfaces 2>/dev/null | awk 'NR>2 && $4=="Up" {print $1}'
    fi
}

# Alias completions if the user has aliases
if alias nat >/dev/null 2>&1; then
    complete -F _nat_manager_complete nat
fi

if alias natmgr >/dev/null 2>&1; then
    complete -F _nat_manager_complete natmgr
fi