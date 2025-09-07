# Fish shell completion for nat-manager
# Copy to ~/.config/fish/completions/nat-manager.fish

# Remove existing completions to avoid conflicts
complete -c nat-manager -e

# Global options
complete -c nat-manager -f -l config -d "Configuration file path" -r
complete -c nat-manager -f -l verbose -s v -d "Enable verbose output"
complete -c nat-manager -f -l config-path -d "Configuration directory path" -r
complete -c nat-manager -f -l help -s h -d "Show help message"
complete -c nat-manager -f -l version -d "Show version information"

# Main commands
complete -c nat-manager -f -n __fish_use_subcommand -a start -d "Start NAT service"
complete -c nat-manager -f -n __fish_use_subcommand -a stop -d "Stop NAT service"  
complete -c nat-manager -f -n __fish_use_subcommand -a status -d "Show NAT service status"
complete -c nat-manager -f -n __fish_use_subcommand -a interfaces -d "List available network interfaces"
complete -c nat-manager -f -n __fish_use_subcommand -a monitor -d "Monitor NAT traffic and connections"
complete -c nat-manager -f -n __fish_use_subcommand -a help -d "Show help information"
complete -c nat-manager -f -n __fish_use_subcommand -a version -d "Show version information"

# Helper functions
function __fish_nat_manager_using_command
    set -l cmd (commandline -opc)
    if [ (count $cmd) -gt 1 ]
        if [ $argv[1] = $cmd[2] ]
            return 0
        end
    end
    return 1
end

function __fish_nat_manager_external_interfaces
    # Try to get interfaces from nat-manager
    if command -sq nat-manager
        nat-manager interfaces 2>/dev/null | awk 'NR>2 && ($2=="Ethernet" || $2=="WiFi") && $4=="Up" {print $1 "\t" $2 " (" $3 ")"}'
    else
        # Fallback to common interfaces
        echo -e "en0\tEthernet (Primary)"
        echo -e "en1\tWiFi"
        echo -e "en2\tEthernet (Secondary)"
    end
end

function __fish_nat_manager_internal_interfaces
    echo -e "bridge100\tBridge Interface 100"
    echo -e "bridge101\tBridge Interface 101"
    echo -e "bridge102\tBridge Interface 102"
    echo -e "utun0\tUser Tunnel 0"
    echo -e "utun1\tUser Tunnel 1"
end

function __fish_nat_manager_networks
    echo -e "192.168.1\tPrivate Class C (192.168.1.0/24)"
    echo -e "192.168.100\tPrivate Class C (192.168.100.0/24)"
    echo -e "10.0.1\tPrivate Class A (10.0.1.0/24)"
    echo -e "172.16.1\tPrivate Class B (172.16.1.0/24)"
end

function __fish_nat_manager_dns_servers
    echo -e "8.8.8.8,8.8.4.4\tGoogle DNS"
    echo -e "1.1.1.1,1.0.0.1\tCloudflare DNS"
    echo -e "208.67.222.222,208.67.220.220\tOpenDNS"
    echo -e "9.9.9.9,149.112.112.112\tQuad9 DNS"
end

function __fish_nat_manager_interface_types
    echo -e "ethernet\tEthernet interfaces"
    echo -e "bridge\tBridge interfaces"
    echo -e "vpn\tVPN tunnel interfaces"
    echo -e "tunnel\tTunnel interfaces"
    echo -e "loopback\tLoopback interfaces"
end

function __fish_nat_manager_intervals
    echo -e "1s\t1 second"
    echo -e "2s\t2 seconds"
    echo -e "5s\t5 seconds"
    echo -e "10s\t10 seconds"
    echo -e "30s\t30 seconds"
    echo -e "1m\t1 minute"
    echo -e "2m\t2 minutes"
    echo -e "5m\t5 minutes"
end

function __fish_nat_manager_connection_limits
    echo -e "10\t10 connections"
    echo -e "20\t20 connections"
    echo -e "50\t50 connections"
    echo -e "100\t100 connections"
    echo -e "200\t200 connections"
end

# Start command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l external -s e -d "External interface" -a "(__fish_nat_manager_external_interfaces)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l internal -s i -d "Internal interface" -a "(__fish_nat_manager_internal_interfaces)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l network -s n -d "Internal network" -a "(__fish_nat_manager_networks)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l dhcp-start -d "DHCP range start" -a "192.168.100.100 10.0.1.100"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l dhcp-end -d "DHCP range end" -a "192.168.100.200 10.0.1.200"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l dns -d "DNS servers" -a "(__fish_nat_manager_dns_servers)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command start' -l help -s h -d "Show help for start command"

# Stop command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command stop' -l force -s f -d "Force stop even if some operations fail"
complete -c nat-manager -f -n '__fish_nat_manager_using_command stop' -l help -s h -d "Show help for stop command"

# Status command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command status' -l json -d "Output status in JSON format"
complete -c nat-manager -f -n '__fish_nat_manager_using_command status' -l help -s h -d "Show help for status command"

# Interfaces command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command interfaces' -l all -s a -d "Show all interfaces including inactive"
complete -c nat-manager -f -n '__fish_nat_manager_using_command interfaces' -l type -s t -d "Filter by interface type" -a "(__fish_nat_manager_interface_types)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command interfaces' -l help -s h -d "Show help for interfaces command"

# Monitor command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command monitor' -l interval -s i -d "Refresh interval" -a "(__fish_nat_manager_intervals)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command monitor' -l max -s m -d "Maximum connections to display" -a "(__fish_nat_manager_connection_limits)"
complete -c nat-manager -f -n '__fish_nat_manager_using_command monitor' -l devices -s d -d "Show connected devices"
complete -c nat-manager -f -n '__fish_nat_manager_using_command monitor' -l follow -s f -d "Continuous monitoring mode"
complete -c nat-manager -f -n '__fish_nat_manager_using_command monitor' -l help -s h -d "Show help for monitor command"

# Help command completions
complete -c nat-manager -f -n '__fish_nat_manager_using_command help' -a "start stop status interfaces monitor" -d "Get help for specific command"

# Alias completions if user has created aliases
if functions -q nat
    complete -c nat -w nat-manager
end

if functions -q natmgr  
    complete -c natmgr -w nat-manager
end