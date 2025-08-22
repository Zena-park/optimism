#!/bin/bash

# Phase 1 P2P Challenger Network - Automated Installation and Execution Script
# Optimism Sequencer System Improvement Project

set -e  # Stop script on error

# Color Definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Script Information Output
echo "=========================================="
echo "Phase 1 P2P Challenger Network Automation Script"
echo "Optimism Sequencer System Improvement Project"
echo "=========================================="
echo

# Execute System Tools Installation Script
log_info "Starting system status check and tools installation..."
if [ -f "$SCRIPT_DIR/install-tools.sh" ]; then
    log_success "Executing system tools installation script..."
    bash "$SCRIPT_DIR/install-tools.sh"
else
    log_warning "System tools installation script not found. Proceeding with manual installation..."
    # Manual installation logic (existing code)
fi

echo

# Basic Settings
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPTIMISM_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"
CHALLENGER_DIR="/opt/optimism/challenger"
LOG_DIR="/var/log/optimism/challenger"
KEYS_DIR="$CHALLENGER_DIR/keys"

# Network Selection (Devnet Detection and Selection)
select_network() {
    # Check if Devnet is running
    local devnet_running=false
    if kurtosis enclave list | grep -q "simple-devnet"; then
        devnet_running=true
        log_info "Local Devnet is running."
    fi

    echo "Select network:"
    echo "1) Sepolia (testnet - recommended)"
    echo "2) Mainnet (mainnet)"
    echo "3) Goerli (testnet)"
    if [ "$devnet_running" = true ]; then
        echo "4) Local Devnet (Kurtosis) - running"
    else
        echo "4) Local Devnet (Kurtosis) - create new"
    fi
    read -p "Selection (1-4): " network_choice

    case $network_choice in
        1)
            NETWORK="sepolia"
            L1_RPC="https://ethereum-sepolia-rpc.publicnode.com"
            L1_BEACON="https://ethereum-sepolia-beacon-api.publicnode.com"
            L2_RPC="https://sepolia.optimism.io"
            ROLLUP_RPC="https://sepolia.optimism.io"
            P2P_NETWORK_ID="optimism-challenger-sepolia"
            P2P_MAX_PEERS=50
            ;;
        2)
            NETWORK="mainnet"
            L1_RPC="https://ethereum-rpc.publicnode.com"
            L1_BEACON="https://ethereum-beacon-api.publicnode.com"
            L2_RPC="https://mainnet.optimism.io"
            ROLLUP_RPC="https://mainnet.optimism.io"
            P2P_NETWORK_ID="optimism-challenger-mainnet"
            P2P_MAX_PEERS=100
            ;;
        3)
            NETWORK="goerli"
            L1_RPC="https://ethereum-goerli-rpc.publicnode.com"
            L1_BEACON="https://ethereum-goerli-beacon-api.publicnode.com"
            L2_RPC="https://goerli.optimism.io"
            ROLLUP_RPC="https://goerli.optimism.io"
            P2P_NETWORK_ID="optimism-challenger-goerli"
            P2P_MAX_PEERS=50
            ;;
        4)
            NETWORK="devnet"
            configure_devnet
            ;;
        *)
            log_error "Invalid selection. Using Sepolia as default."
            NETWORK="sepolia"
            L1_RPC="https://ethereum-sepolia-rpc.publicnode.com"
            L1_BEACON="https://ethereum-sepolia-beacon-api.publicnode.com"
            L2_RPC="https://sepolia.optimism.io"
            ROLLUP_RPC="https://sepolia.optimism.io"
            P2P_NETWORK_ID="optimism-challenger-sepolia"
            P2P_MAX_PEERS=50
            ;;
    esac

    log_success "Selected network: $NETWORK"
}

# Devnet Configuration
configure_devnet() {
    log_info "Local Devnet Configuration"

    # If DEVNET_CHOICE is already set, process automatically
    if [ -n "$DEVNET_CHOICE" ]; then
        devnet_choice="$DEVNET_CHOICE"
        log_info "Auto-selection: Connect to existing devnet"
    else
        echo "Configure local Devnet settings."
        echo "1) Connect to existing devnet"
        echo "2) Auto-create new devnet (Kurtosis)"
        read -p "Selection (1-2): " devnet_choice
    fi

    case $devnet_choice in
        1)
            configure_existing_devnet
            ;;
        2)
            setup_new_devnet
            ;;
        *)
            log_warning "Invalid selection. Using existing devnet connection."
            configure_existing_devnet
            ;;
    esac
}

# Configure Existing Devnet Connection
configure_existing_devnet() {
    log_info "Configure Existing Devnet Connection"

    # Use default values if auto-detected Devnet exists, otherwise user input
    local auto_detected=false
    if kurtosis enclave list | grep -q "simple-devnet"; then
        auto_detected=true
        log_info "Local Devnet detected. Using default settings."
    fi

    if [ "$auto_detected" = true ]; then
        # Local Devnet default settings
        L1_RPC="http://localhost:8545"
        L1_BEACON="http://localhost:5052"
        L2_RPC="http://localhost:9545"
        ROLLUP_RPC="http://localhost:9545"
        P2P_NETWORK_ID="optimism-challenger-devnet"
        P2P_MAX_PEERS=20
    else
        echo "Press Enter to use default values."

        # L1 RPC Configuration
        read -p "L1 RPC URL (default: http://localhost:8545): " l1_rpc_input
        L1_RPC="${l1_rpc_input:-http://localhost:8545}"

        # L1 Beacon Configuration (optional)
        read -p "L1 Beacon URL (default: http://localhost:5052, leave empty to not use): " l1_beacon_input
        L1_BEACON="${l1_beacon_input:-http://localhost:5052}"

        # L2 RPC Configuration
        read -p "L2 RPC URL (default: http://localhost:9545): " l2_rpc_input
        L2_RPC="${l2_rpc_input:-http://localhost:9545}"

        # Rollup RPC Configuration
        read -p "Rollup RPC URL (default: http://localhost:9545): " rollup_rpc_input
        ROLLUP_RPC="${rollup_rpc_input:-http://localhost:9545}"

        # P2P Network ID Configuration
        read -p "P2P Network ID (default: optimism-challenger-devnet): " p2p_network_id_input
        P2P_NETWORK_ID="${p2p_network_id_input:-optimism-challenger-devnet}"

        # P2P Max Peers Configuration
        read -p "P2P Max Peers (default: 20): " p2p_max_peers_input
        P2P_MAX_PEERS="${p2p_max_peers_input:-20}"
    fi

    # Check Devnet ports
    check_devnet_ports

    log_success "Existing Devnet connection configuration completed"
    log_info "L1 RPC: $L1_RPC"
    log_info "L2 RPC: $L2_RPC"
    log_info "Rollup RPC: $ROLLUP_RPC"
    log_info "P2P Network ID: $P2P_NETWORK_ID"
}

# Auto-create New Devnet
setup_new_devnet() {
    log_info "Auto-creating New Kurtosis Devnet"

    # Check Kurtosis
    if ! command -v kurtosis &> /dev/null; then
        log_error "Kurtosis is not installed. Please install Kurtosis first."
        log_info "Kurtosis installation: https://docs.kurtosis.com/install"
        log_info "Or auto-install via mise:"
        log_info "  curl -fsSL https://mise.run | sh"
        log_info "  eval \"\$(/Users/zena/.local/bin/mise activate zsh)\"  # activate mise"
        log_info "  mise trust  # trust config files (if needed)"
        log_info "  mise install kurtosis"
        exit 1
    fi

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        log_info "Docker installation: https://docs.docker.com/get-docker/"
        exit 1
    fi

    log_success "Kurtosis and Docker verification completed"

    # Check Kurtosis devnet directory
    if [ ! -d "$OPTIMISM_ROOT/kurtosis-devnet" ]; then
        log_error "Kurtosis devnet directory not found: $OPTIMISM_ROOT/kurtosis-devnet"
        log_info "Please check if Optimism repository is fully cloned."
        exit 1
    fi

    # Start Devnet
    log_info "Starting Kurtosis devnet..."
    cd "$OPTIMISM_ROOT/kurtosis-devnet"

    # Clean up existing devnet
    if kurtosis enclave list | grep -q "simple-devnet"; then
        log_info "Cleaning up existing devnet..."
        kurtosis rm simple-devnet || true
        sleep 2
    fi

        # Start new devnet (using our own Devnet builder)
    log_info "Starting new Devnet... (using our own builder)"
    log_info "This process takes 5-15 minutes..."
    log_info "Automatically resolving Go version compatibility issues..."

    # Execute our own Devnet builder
    log_info "Executing our own Devnet builder..."
    local build_script_path="$SCRIPT_DIR/build-devnet.sh"
    if [ -f "$build_script_path" ]; then
        if bash "$build_script_path"; then
        local build_exit_code=0
        log_success "Devnet build successful"
            else
            local build_exit_code=1
            log_error "Devnet build failed"
        fi
    else
        log_error "build-devnet.sh file not found"
        log_info "Current directory: $(pwd)"
        log_info "Please check build-devnet.sh location"
        local build_exit_code=1
    fi

        # Check build result
    log_info "=== Build Result Check ==="

    if [ $build_exit_code -eq 0 ]; then
        log_success "✅ Devnet build completed successfully"
        log_info "Our own Devnet builder successfully processed all steps."
    else
        log_error "❌ Devnet build failed"
        log_info "Check build logs to diagnose the issue:"
        log_info "cat /tmp/devnet-build.log"
        exit 1
    fi

    # devnet JSON 출력에서 포트 정보 추출
    if [ -f /tmp/devnet_output.json ]; then
        # JSON 파싱을 위해 jq 사용 (없으면 기본값 사용)
        if command -v jq &> /dev/null; then
            L1_EL_PORT=$(jq -r '.l1.nodes[0].el' /tmp/devnet_output.json | sed 's|http://localhost:||')
            L1_CL_PORT=$(jq -r '.l1.nodes[0].cl' /tmp/devnet_output.json | sed 's|http://localhost:||')
            L2_EL_PORT=$(jq -r '.l2[0].nodes[0].el' /tmp/devnet_output.json | sed 's|http://localhost:||')
            L2_CL_PORT=$(jq -r '.l2[0].nodes[0].cl' /tmp/devnet_output.json | sed 's|http://localhost:||')
        else
            # jq가 없으면 기본 포트 사용
            L1_EL_PORT="53620"
            L1_CL_PORT="53689"
            L2_EL_PORT="56781"
            L2_CL_PORT="57029"
        fi
    else
        # 기본 포트 사용
        L1_EL_PORT="53620"
        L1_CL_PORT="53689"
        L2_EL_PORT="56781"
        L2_CL_PORT="57029"
    fi

    # Wait for devnet startup and verification
    log_info "Waiting for Devnet to start... (maximum 5 minutes)"

    # Step 1: Check L1/L2 node RPC response
    for i in {1..60}; do
        if curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            "http://localhost:$L1_EL_PORT" > /dev/null 2>&1; then
            log_success "L1 node is ready (port: $L1_EL_PORT)"
            break
        fi
        if [ $i -eq 60 ]; then
            log_error "L1 node startup timeout"
            exit 1
        fi
        sleep 5
    done

    for i in {1..60}; do
        if curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            "http://localhost:$L2_EL_PORT" > /dev/null 2>&1; then
            log_success "L2 node is ready (port: $L2_EL_PORT)"
            break
        fi
        if [ $i -eq 60 ]; then
            log_error "L2 node startup timeout"
            exit 1
        fi
        sleep 5
    done

    # Step 2: Check Optimism service deployment
    log_info "Checking Optimism service deployment status..."
    sleep 10  # Wait for services to start

    # Check Docker containers
    local op_services=0
    local total_services=5  # op-node, op-batcher, op-proposer, op-faucet, geth

    if docker ps --format "{{.Names}}" | grep -q "op-node"; then
        log_success "✅ op-node service running"
        ((op_services++))
    else
        log_warning "❌ op-node service not running"
    fi

    if docker ps --format "{{.Names}}" | grep -q "op-batcher"; then
        log_success "✅ op-batcher service running"
        ((op_services++))
    else
        log_warning "❌ op-batcher service not running"
    fi

    if docker ps --format "{{.Names}}" | grep -q "op-proposer"; then
        log_success "✅ op-proposer service running"
        ((op_services++))
    else
        log_warning "❌ op-proposer service not running"
    fi

    if docker ps --format "{{.Names}}" | grep -q "op-faucet"; then
        log_success "✅ op-faucet service running"
        ((op_services++))
    else
        log_warning "❌ op-faucet service not running"
    fi

    if docker ps --format "{{.Names}}" | grep -q "geth"; then
        log_success "✅ geth service running"
        ((op_services++))
    else
        log_warning "❌ geth service not running"
    fi

    # Service deployment status summary
    if [ $op_services -eq $total_services ]; then
        log_success "🎉 All Optimism services deployed successfully! ($op_services/$total_services)"
    elif [ $op_services -gt 0 ]; then
        log_warning "⚠️  Only some Optimism services deployed ($op_services/$total_services)"
        log_info "Devnet will work but some functionality may be limited."
    else
        log_error "❌ No Optimism services deployed (0/$total_services)"
        log_info "Devnet infrastructure is running but Optimism services failed to start."
        log_info "This may be due to Docker build failure or Go version compatibility issues."
        exit 1
    fi

    # Apply settings
    L1_RPC="http://localhost:$L1_EL_PORT"
    L1_BEACON="http://localhost:$L1_CL_PORT"
    L2_RPC="http://localhost:$L2_EL_PORT"
    ROLLUP_RPC="http://localhost:$L2_CL_PORT"
    P2P_NETWORK_ID="optimism-challenger-devnet"
    P2P_MAX_PEERS=20

    log_success "Devnet creation completed!"
    log_info "L1 RPC: $L1_RPC"
    log_info "L2 RPC: $L2_RPC"
    log_info "Rollup RPC: $ROLLUP_RPC"
    log_info "P2P Network ID: $P2P_NETWORK_ID"

    log_success "Kurtosis devnet auto-creation completed!"
    log_info "L1 RPC: $L1_RPC"
    log_info "L1 Beacon: $L1_BEACON"
    log_info "L2 RPC: $L2_RPC"
    log_info "Rollup RPC: $ROLLUP_RPC"
    log_info "P2P Network ID: $P2P_NETWORK_ID"

    # Output devnet management information
    echo
    echo "=== Kurtosis Devnet Management Commands ==="
    echo "Check Devnet status: cd $OPTIMISM_ROOT/kurtosis-devnet && kurtosis enclave inspect simple-devnet"
    echo "Stop Devnet: cd $OPTIMISM_ROOT/kurtosis-devnet && kurtosis rm simple-devnet"
    echo "Restart Devnet: cd $OPTIMISM_ROOT/kurtosis-devnet && just simple-devnet"
    echo "Check Devnet logs: cd $OPTIMISM_ROOT/kurtosis-devnet && kurtosis enclave logs simple-devnet"
    echo "Complete cleanup: cd $OPTIMISM_ROOT/kurtosis-devnet && kurtosis clean"
    echo
}

# Check Devnet Ports
check_devnet_ports() {
    log_info "Checking Devnet port connections..."

    # Check L1 RPC port
    if [[ "$L1_RPC" == *"localhost"* ]] || [[ "$L1_RPC" == *"127.0.0.1"* ]]; then
        local l1_port=$(echo "$L1_RPC" | sed 's/.*://' | sed 's/\/.*//')
        if ! nc -z localhost "$l1_port" 2>/dev/null; then
            log_warning "Cannot connect to L1 RPC port $l1_port. Check if Kurtosis devnet is running."
            log_info "Start Devnet: cd $OPTIMISM_ROOT/kurtosis-devnet && just simple-devnet"
        else
            log_success "L1 RPC port $l1_port connection verified"
        fi
    fi

    # Check L2 RPC port
    if [[ "$L2_RPC" == *"localhost"* ]] || [[ "$L2_RPC" == *"127.0.0.1"* ]]; then
        local l2_port=$(echo "$L2_RPC" | sed 's/.*://' | sed 's/\/.*//')
        if ! nc -z localhost "$l2_port" 2>/dev/null; then
            log_warning "Cannot connect to L2 RPC port $l2_port. Check if Kurtosis devnet is running."
            log_info "Start Devnet: cd $OPTIMISM_ROOT/kurtosis-devnet && just simple-devnet"
        else
            log_success "L2 RPC port $l2_port connection verified"
        fi
    fi
}

# Check System Requirements
check_requirements() {
    log_info "Checking system requirements..."

    # Check OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
    else
        log_error "Unsupported OS: $OSTYPE"
        exit 1
    fi

    # Check Go version
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)

    if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
        log_error "Go 1.21+ required. Current version: $GO_VERSION"
        exit 1
    fi

    log_success "Go version check: $GO_VERSION"

    # Check Git
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed."
        exit 1
    fi

    # Check Make
    if ! command -v make &> /dev/null; then
        log_error "Make is not installed."
        exit 1
    fi

    # Check memory
    if [[ "$OS" == "linux" ]]; then
        MEMORY_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}')
        MEMORY_GB=$((MEMORY_KB / 1024 / 1024))
    else
        MEMORY_GB=$(sysctl -n hw.memsize | awk '{print $0/1024/1024/1024}')
    fi

    if [ "$MEMORY_GB" -lt 4 ]; then
        log_warning "Recommended memory: 4GB+, current: ${MEMORY_GB}GB"
    else
        log_success "Memory check: ${MEMORY_GB}GB"
    fi

    # Check disk space
    # Disk space check for macOS and Linux compatibility
if [[ "$OSTYPE" == "darwin"* ]]; then
    DISK_GB=$(df -g . | tail -1 | awk '{print $4}')
else
    DISK_GB=$(df -BG . | tail -1 | awk '{print $4}' | sed 's/G//')
fi
    if [ "$DISK_GB" -lt 20 ]; then
        log_warning "Recommended disk space: 20GB+, current: ${DISK_GB}GB"
    else
        log_success "Disk space check: ${DISK_GB}GB"
    fi
}

# Create Directories
create_directories() {
    log_info "Creating required directories..."

    # Create directories with sudo permissions
    sudo mkdir -p "$CHALLENGER_DIR"
    sudo mkdir -p "$LOG_DIR"
    sudo mkdir -p "$KEYS_DIR"

    # Change user permissions
    sudo chown -R "$USER:$USER" "$CHALLENGER_DIR"
    sudo chown -R "$USER:$USER" "$LOG_DIR"

    log_success "Directory creation completed"
}

# Clone and Build Optimism Repository
build_optimism() {
    log_info "Cloning and building Optimism repository..."

    # Check if already exists
    if [ -d "$OPTIMISM_ROOT/.git" ]; then
        log_info "Existing Optimism repository found, updating..."
        cd "$OPTIMISM_ROOT"
        git fetch origin
        git checkout main
        git pull origin main
    else
        log_info "Cloning Optimism repository..."
        cd "$(dirname "$OPTIMISM_ROOT")"
        git clone https://github.com/ethereum-optimism/optimism.git
        cd optimism
    fi

    # Checkout P2P challenger branch (if exists)
    if git branch -r | grep -q "feature/p2p-challenger"; then
        log_info "Checking out P2P challenger branch..."
        git checkout feature/p2p-challenger-attention-test || git checkout main
    fi

        # Build op-challenger (from current directory)
    log_info "Building op-challenger..."
    go build -o bin/op-challenger ./cmd/

    if [ ! -f "bin/op-challenger" ]; then
        log_error "op-challenger build failed"
        exit 1
    fi

    log_success "op-challenger build completed"

    # Build Cannon
    log_info "Building Cannon..."
    cd ../cannon
    make cannon

    if [ ! -f "bin/cannon" ]; then
        log_error "Cannon build failed"
        exit 1
    fi

    log_success "Cannon build completed"

    # Build op-program (including prestate files)
    log_info "Building op-program (including prestate files)..."
    cd ../op-program
    make op-program

    # Check prestate file
    if [ ! -f "bin/prestate-mt64Next.bin.gz" ]; then
        log_error "prestate file was not generated!"
        log_error "This is essential for P2P challenger operation."
        exit 1
    fi

    log_success "op-program build completed (including prestate files)"

    # Copy binaries to system path (optional)
    read -p "Copy binaries to system path? (y/N): " copy_binaries
    if [[ $copy_binaries =~ ^[Yy]$ ]]; then
        sudo cp ../cannon/bin/cannon /usr/local/bin/
        sudo cp bin/op-program /usr/local/bin/
        log_success "Binaries copied to system path"
    fi
}

# Create Environment Configuration File
create_env_file() {
    log_info "Creating environment configuration file..."

    cat > "$CHALLENGER_DIR/.env" << EOF
# Optimism P2P Challenger Environment Configuration
# Network: $NETWORK

# 네트워크 설정
export OP_CHALLENGER_NETWORK="$NETWORK"
export OP_CHALLENGER_L1_ETH_RPC="$L1_RPC"
export OP_CHALLENGER_L1_BEACON="$L1_BEACON"
export OP_CHALLENGER_L2_ETH_RPC="$L2_RPC"
export OP_CHALLENGER_ROLLUP_RPC="$ROLLUP_RPC"
export OP_CHALLENGER_DATADIR="$CHALLENGER_DIR/data"

# P2P 설정
export OP_CHALLENGER_P2P_ENABLED=true
export OP_CHALLENGER_P2P_LISTEN_ADDR="/ip4/0.0.0.0/tcp/9876"
export OP_CHALLENGER_P2P_NETWORK_ID="$P2P_NETWORK_ID"
export OP_CHALLENGER_P2P_MAX_PEERS=$P2P_MAX_PEERS
export OP_CHALLENGER_P2P_DISCOVERY_ENABLED=true
export OP_CHALLENGER_P2P_RATE_LIMIT=1000
export OP_CHALLENGER_P2P_CONNECTION_LIMIT=100
export OP_CHALLENGER_P2P_PRIVATE_KEY="$KEYS_DIR/p2p-key.pem"

# 바이너리 경로
export OP_CHALLENGER_CANNON_BIN="$OPTIMISM_ROOT/cannon/bin/cannon"
export OP_CHALLENGER_CANNON_SERVER="$OPTIMISM_ROOT/op-program/bin/op-program"
export OP_CHALLENGER_CANNON_PRESTATE="$OPTIMISM_ROOT/op-program/bin/prestate-mt64Next.bin.gz"

# 모니터링
export OP_CHALLENGER_METRICS_ENABLED=true
export OP_CHALLENGER_METRICS_ADDR="0.0.0.0"
export OP_CHALLENGER_METRICS_PORT=7300

# 로깅
export OP_CHALLENGER_LOG_LEVEL="info"
export OP_CHALLENGER_LOG_FORMAT="json"

# Go 런타임 최적화
export GOMAXPROCS=4
export GOMEMLIMIT=6GB
export GOGC=100
EOF

    log_success "Environment configuration file created: $CHALLENGER_DIR/.env"
}

# Create systemd Service File
create_systemd_service() {
    log_info "Creating systemd service file..."

    sudo tee /etc/systemd/system/op-challenger.service > /dev/null << EOF
[Unit]
Description=Optimism Challenger with P2P Networking
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
Group=$USER
WorkingDirectory=$CHALLENGER_DIR
ExecStart=$OPTIMISM_ROOT/op-challenger/bin/op-challenger
EnvironmentFile=$CHALLENGER_DIR/.env
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# 리소스 제한
LimitNOFILE=65536
LimitNPROC=4096

# 보안 설정
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$CHALLENGER_DIR $LOG_DIR

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    sudo systemctl daemon-reload

    log_success "systemd service file created"
}

# Configure Firewall
configure_firewall() {
    log_info "Configuring firewall..."

    # Open P2P port
    if command -v ufw &> /dev/null; then
        sudo ufw allow 9876/tcp
        sudo ufw allow from 127.0.0.1 to any port 7300
        log_success "UFW firewall configuration completed"
    elif command -v iptables &> /dev/null; then
        sudo iptables -A INPUT -p tcp --dport 9876 -j ACCEPT
        sudo iptables -A INPUT -p tcp --dport 7300 -s 127.0.0.1 -j ACCEPT
        log_success "iptables firewall configuration completed"
    else
        log_warning "Firewall tool not found. Please open ports manually:"
        echo "  - TCP 9876 (P2P communication)"
        echo "  - TCP 7300 (metrics, local only)"
    fi
}

# Configure Log Rotation
configure_log_rotation() {
    log_info "Configuring log rotation..."

    sudo tee /etc/logrotate.d/op-challenger > /dev/null << EOF
$LOG_DIR/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 $USER $USER
    postrotate
        systemctl reload op-challenger
    endscript
}
EOF

    log_success "Log rotation configuration completed"
}

# Run Tests
run_tests() {
    log_info "Running basic tests..."

    # Check P2P flags
    if $OPTIMISM_ROOT/op-challenger/bin/op-challenger --help | grep -q "p2p-enabled"; then
        log_success "P2P flags verification completed"
    else
        log_error "P2P flags not found!"
        exit 1
    fi

    # Check binary existence
    if [ -f "$OPTIMISM_ROOT/cannon/bin/cannon" ]; then
        log_success "Cannon binary verification completed"
    else
        log_error "Cannon binary not found!"
        exit 1
    fi

    if [ -f "$OPTIMISM_ROOT/op-program/bin/op-program" ]; then
        log_success "op-program binary verification completed"
    else
        log_error "op-program binary not found!"
        exit 1
    fi

    if [ -f "$OPTIMISM_ROOT/op-program/bin/prestate-mt64Next.bin.gz" ]; then
        log_success "prestate file verification completed"
    else
        log_error "prestate file not found!"
        exit 1
    fi
}

# Start Service
start_service() {
    log_info "Starting P2P challenger service..."

    # Enable and start service
    sudo systemctl enable op-challenger
    sudo systemctl start op-challenger

    # Check status
    sleep 3
    if sudo systemctl is-active --quiet op-challenger; then
        log_success "P2P challenger service started successfully!"
    else
        log_error "Service startup failed"
        sudo systemctl status op-challenger
        exit 1
    fi
}

# Check Status
check_status() {
    log_info "Checking service status..."

    echo
    echo "=== Service Status ==="
    sudo systemctl status op-challenger --no-pager -l

    echo
    echo "=== Recent Logs ==="
    sudo journalctl -u op-challenger --no-pager -l -n 20

    echo
    echo "=== Check Metrics Endpoint ==="
    if curl -s http://localhost:7300/metrics > /dev/null 2>&1; then
        log_success "Metrics server is working normally"
        echo "Metrics URL: http://localhost:7300/metrics"
    else
        log_warning "Cannot access metrics server"
    fi

    echo
    echo "=== P2P Network Status ==="
    if curl -s http://localhost:7300/debug/peers > /dev/null 2>&1; then
        log_success "P2P debug endpoint is working normally"
        echo "Peer information: http://localhost:7300/debug/peers"
    else
        log_warning "Cannot access P2P debug endpoint"
    fi
}

# Show Usage
show_usage() {
    echo
    echo "=== P2P Challenger Management Commands ==="
    echo "Start service:   sudo systemctl start op-challenger"
    echo "Stop service:    sudo systemctl stop op-challenger"
    echo "Restart service: sudo systemctl restart op-challenger"
    echo "Check status:    sudo systemctl status op-challenger"
    echo "Check logs:      sudo journalctl -u op-challenger -f"
    echo "Check metrics:   curl http://localhost:7300/metrics"
    echo "Check peers:     curl http://localhost:7300/debug/peers"
    echo
    echo "=== Configuration File Locations ==="
    echo "Environment:     $CHALLENGER_DIR/.env"
    echo "Service file:    /etc/systemd/system/op-challenger.service"
    echo "Log directory:   $LOG_DIR"
    echo "Key directory:   $KEYS_DIR"
    echo
    echo "=== Network Information ==="
    echo "Network:         $NETWORK"
    echo "P2P Port:        9876"
    echo "Metrics Port:    7300"
    echo "P2P Network ID:  $P2P_NETWORK_ID"
}

# Main Execution Function
main() {
    echo "Starting Phase 1 P2P Challenger Network automated installation..."
    echo

    # Select network
    select_network

    # Auto-execute when Devnet is selected
    if [ "$NETWORK" = "devnet" ]; then
        log_info "Running in Devnet mode..."
        log_success "Starting Kurtosis Devnet..."

        # Move to kurtosis-devnet directory
        cd "$OPTIMISM_ROOT/kurtosis-devnet"

        # Start Devnet
        log_info "Starting Simple devnet..."
        just simple-devnet

        log_success "Devnet started successfully!"
        echo
        echo "Devnet Information:"
        echo "- L1 RPC: $L1_RPC"
        echo "- L2 RPC: $L2_RPC"
        echo "- P2P Network ID: $P2P_NETWORK_ID"
        echo
        echo "Next Steps:"
        echo "1. Set up P2P challenger network"
        echo "2. Connect with other challengers"
        echo "3. Monitor network status"
        return 0
    fi

    # Check system requirements
    check_requirements

    # Create directories
    create_directories

    # Build Optimism
    build_optimism

    # Create environment configuration file
    create_env_file

    # Create systemd service file
    create_systemd_service

    # Configure firewall
    configure_firewall

    # Configure log rotation
    configure_log_rotation

    # Run tests
    run_tests

    # Start service
    start_service

    # Check status
    check_status

    # Show usage
    show_usage

    echo
    log_success "Phase 1 P2P Challenger Network installation completed! 🎉"
    echo
    echo "Next Steps:"
    echo "1. Connect with other nodes to form P2P network"
    echo "2. Monitor metrics to check network status"
    echo "3. Check logs regularly to detect issues early"
    echo
    echo "If problems occur, check logs:"
    echo "  sudo journalctl -u op-challenger -f"
}

# Execute Script
main "$@"
