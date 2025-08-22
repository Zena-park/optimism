# Phase 1 P2P Challenger Network - Installation Guide

## 📋 Table of Contents

1. [🚀 Quick Installation (Recommended)](#-quick-installation-recommended)
2. [💻 System Requirements](#-system-requirements)
3. [🔧 Automated Installation Process](#-automated-installation-process)
4. [🎯 Complete Automation Workflow](#-complete-automation-workflow)
5. [📊 Installation Verification](#-installation-verification)
6. [🔧 Manual Installation (Advanced Users)](#-manual-installation-advanced-users)

---

## 🚀 Quick Installation (Recommended)

### Local Devnet Development Environment
```bash
# Step 1: Install system tools
cd op-challenger/scripts/p2p
./install-tools.sh

# Step 2: Build Devnet Environment
./build-devnet.sh

# Step 3: Run P2P Challenger
./run-challenger-devnet.sh
```

### System Tools Auto Installation
```bash
# Install system tools only
cd op-challenger/scripts/p2p
./install-tools.sh
```

**What the auto-installation script does:**
- ✅ Automatic system status verification
- ✅ Auto-installation of missing tools
- ✅ Go 1.24.6+ auto-installation
- ✅ Docker, Mise, Kurtosis, Just auto-installation
- ✅ Optimized installation order (considering dependencies)
- ✅ User confirmation before installation

---

## 💻 System Requirements

### Minimum Requirements
- **OS**: Linux (Ubuntu 20.04+ recommended), macOS 12+
- **Go**: 1.24.6+ (auto-installed)
- **Docker**: Latest version (auto-installed)
- **Git**: 2.0+
- **Memory**: 8GB+ RAM
- **Disk**: 50GB+ free space
- **Network**: Stable internet connection

### Recommended Requirements
- **Memory**: 16GB+ RAM
- **Disk**: 100GB+ SSD
- **Network**: 100Mbps+ bandwidth
- **CPU**: 8+ cores

### Devnet Requirements (Local Development Environment)
- **Docker**: Latest version (auto-installed)
- **Docker Compose**: Latest version (auto-installed)
- **Mise**: Tool version management (auto-installed)
- **Kurtosis**: Devnet management (auto-installed)
- **Just**: Build tool (auto-installed)
- **Ports**: Dynamic allocation (automatically managed by Kurtosis)
- **Memory**: 8GB+ RAM (devnet + P2P challenger)
- **Disk**: 50GB+ free space (including devnet data)

---

## 🔧 Automated Installation Process

### 1. System Requirements Verification
The auto-installation script verifies the following:
- OS compatibility
- Memory and disk space
- Existing tool installation status
- Network connection status

### 2. Tool Auto Installation
```bash
# Go 1.24.6+ installation
mise install go@1.24.6
mise use go@1.24.6

# Docker installation (macOS)
brew install --cask docker

# Mise installation
curl https://mise.run | sh

# Kurtosis installation
mise install kurtosis@latest

# Just installation
mise install just@latest
```

### 3. Devnet Auto Creation
```bash
# Use our Devnet builder
./build-devnet.sh
```

**Build Process:**
1. System requirements verification
2. Docker image build (all Optimism services)
3. Devnet deployment (automatic deployment via Kurtosis)
4. Status verification (all services normal operation check)

**Build Results:**
- ✅ op-node, op-batcher, op-proposer, op-faucet, op-challenger, op-deployer: Success
- ✅ geth: Base image used

### 4. P2P Challenger Auto Setup
```bash
# Local Devnet-only challenger execution
./run-challenger-devnet.sh
```

**Setup Process:**
1. Devnet status check
2. op-challenger Docker image verification
3. Automatic Devnet port information extraction
4. op-challenger container execution
5. Service status verification

---

## 🎯 Complete Automation Workflow

### One-Click Complete Setup
```bash
# Complete automation (system tools + devnet + challenger)
cd op-challenger/scripts/p2p
./install-tools.sh
./build-devnet.sh
./run-challenger-devnet.sh
```

### Step-by-Step Setup (Detailed Control)
```bash
# 1. System tools installation
./install-tools.sh

# 2. New Devnet creation
./build-devnet.sh

# 3. P2P challenger setup
./run-challenger-devnet.sh
```

### Existing Devnet Utilization
```bash
# 1. System tools installation
./install-tools.sh

# 2. Connect to existing Devnet
./run-challenger-devnet.sh
```

---

## 📊 Installation Verification

### System Tools Verification
```bash
# Check Go version
go version

# Check Docker version
docker --version

# Check Mise version
mise --version

# Check Kurtosis version
kurtosis version

# Check Just version
just --version
```

### Devnet Verification
```bash
# Check Devnet status
kurtosis enclave inspect simple-devnet

# Check service list
kurtosis enclave inspect simple-devnet --format json | jq '.services'

# Check specific service status
kurtosis enclave inspect simple-devnet --service op-node
kurtosis enclave inspect simple-devnet --service op-challenger
```

### P2P Challenger Verification
```bash
# Check challenger status
docker ps | grep challenger

# Check challenger logs
docker logs op-challenger

# Check challenger configuration
docker exec op-challenger op-challenger --help
```

### Network Connectivity Verification
```bash
# Check L1 RPC connectivity
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:53620

# Check L2 RPC connectivity
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:56781

# Check Rollup RPC connectivity
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"optimism_syncStatus","params":[],"id":1}' \
  http://localhost:57029
```

### Performance Verification
```bash
# Check system resources
htop
df -h
free -h

# Check Docker resources
docker stats

# Check network usage
netstat -i
```

---

## 🔧 Manual Installation (Advanced Users)

### Manual Go Installation
```bash
# Download Go 1.24.6
wget https://go.dev/dl/go1.24.6.linux-amd64.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf go1.24.6.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### Manual Docker Installation
```bash
# Install Docker (Ubuntu)
sudo apt-get update
sudo apt-get install docker.io docker-compose

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group
sudo usermod -aG docker $USER

# Verify installation
docker --version
```

### Manual Mise Installation
```bash
# Install Mise
curl https://mise.run | sh

# Add to PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
mise --version
```

### Manual Kurtosis Installation
```bash
# Install Kurtosis via Mise
mise install kurtosis@latest

# Verify installation
kurtosis version
```

### Manual Just Installation
```bash
# Install Just via Mise
mise install just@latest

# Verify installation
just --version
```

### Manual Devnet Setup
```bash
# Clone Optimism repository
git clone https://github.com/ethereum-optimism/optimism.git
cd optimism

# Build Docker images
just op-node-image
just op-batcher-image
just op-proposer-image
just op-challenger-image
just op-faucet-image
just op-deployer-image

# Run Devnet
kurtosis run --enclave simple-devnet kurtosis-devnet/simple.yaml
```

### Manual P2P Challenger Setup
```bash
# Get Devnet port information
kurtosis enclave inspect simple-devnet

# Run challenger with correct ports
docker run -d \
  --name op-challenger \
  --network host \
  op-challenger:devnet \
  --l1-eth-rpc "http://localhost:53620" \
  --l2-eth-rpc "http://localhost:56781" \
  --rollup-rpc "http://localhost:57029" \
  --p2p-enabled \
  --p2p-network-id "optimism-challenger-devnet" \
  --p2p-max-peers 20 \
  --log.level INFO
```

---

## 🔧 Troubleshooting

### Common Installation Issues

#### Go Installation Issues
```bash
# Check Go installation
which go
go version

# Reinstall Go if needed
mise uninstall go
mise install go@1.24.6
```

#### Docker Installation Issues
```bash
# Check Docker installation
which docker
docker --version

# Check Docker service
sudo systemctl status docker

# Restart Docker service
sudo systemctl restart docker
```

#### Mise Installation Issues
```bash
# Check Mise installation
which mise
mise --version

# Reinstall Mise if needed
curl https://mise.run | sh
```

#### Kurtosis Installation Issues
```bash
# Check Kurtosis installation
which kurtosis
kurtosis version

# Reinstall Kurtosis if needed
mise uninstall kurtosis
mise install kurtosis@latest
```

#### Just Installation Issues
```bash
# Check Just installation
which just
just --version

# Reinstall Just if needed
mise uninstall just
mise install just@latest
```

### Devnet Build Issues

#### Docker Build Failures
```bash
# Clean Docker cache
docker system prune -f
docker builder prune -f

# Rebuild from scratch
./build-devnet.sh
```

#### Port Conflicts
```bash
# Check port usage
lsof -i :8545
lsof -i :8547
lsof -i :9876

# Kill conflicting processes
sudo kill -9 <PID>
```

#### Resource Issues
```bash
# Check system resources
htop
df -h
free -h

# Increase Docker resources (Docker Desktop)
# Settings > Resources > Advanced
```

### P2P Challenger Issues

#### Challenger Not Starting
```bash
# Check if Devnet is running
kurtosis enclave ls

# Check challenger image
docker images | grep challenger

# Check challenger logs
docker logs op-challenger
```

#### Network Connectivity Issues
```bash
# Check RPC connectivity
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:53620

# Check network configuration
docker network ls
docker network inspect bridge
```

#### Configuration Issues
```bash
# Check challenger configuration
docker exec op-challenger op-challenger --help

# Check environment variables
docker exec op-challenger env | grep -E "(L1|L2|ROLLUP|P2P)"
```

---

## 📚 Additional Resources

- [Devnet Guide](devnet-guide.md): Detailed Devnet setup and operation guide
- [Optimism Documentation](https://community.optimism.io/): Official Optimism documentation
- [Kurtosis Documentation](https://docs.kurtosis.com/): Kurtosis documentation
- [Docker Documentation](https://docs.docker.com/): Docker documentation
