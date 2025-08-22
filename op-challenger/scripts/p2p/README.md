# Phase 1 P2P Challenger Network

## 🚀 Quick Start Guide

### 🎯 Local Devnet Development Environment (Recommended)
```bash
# Step 1: Build Devnet Environment
cd op-challenger/scripts/p2p
./build-devnet.sh

# Step 2: Run P2P Challenger
./run-challenger-devnet.sh
```

### 📋 Multi-Network Support (Testnet/Mainnet)
```bash
# Support for Sepolia, Mainnet, Goerli and other networks
cd op-challenger/scripts/p2p
./run-challenger-multi.sh
```

### 📋 System Tools Installation
```bash
# Check system status
cd op-challenger/scripts/p2p
./check-system.sh

# Auto-install tools
./install-tools.sh
```

## 📁 Script List

### 🔧 System Management Scripts
- **`check-system.sh`**: System requirements check and diagnostics
- **`install-tools.sh`**: Auto-install Go, Docker, Mise, Kurtosis, Just

### 🚀 Devnet Environment Scripts
- **`build-devnet.sh`**: **Devnet Environment Setup** - Enhanced Devnet builder that replaces Optimism's simple-devnet (automatically resolves dependency issues)

### 🎯 Challenger Execution Scripts
- **`run-challenger-devnet.sh`**: **Local Devnet Only** - P2P challenger execution for development and testing purposes
- **`run-challenger-multi.sh`**: **Multi-Network Support** - Support for Sepolia, Mainnet, Goerli and other networks

### 📊 Management Scripts
- **`check-system.sh`**: System status check and diagnostics

## 🔄 Workflow

### Scenario 1: Local Devnet Development (Recommended)
```bash
# 1. Build Devnet Environment
./build-devnet.sh

# 2. Run P2P Challenger
./run-challenger-devnet.sh
```

### Scenario 2: Multi-Network Support
```bash
# Sepolia, Mainnet, Goerli and other networks
./run-challenger-multi.sh
```

### Scenario 3: System Tools Installation
```bash
# Install system tools
./install-tools.sh
```

## ✅ Resolved Issues

### 🔧 Automatically Resolved Technical Issues
- **GitHub API Rate Limit**: Bypassed by direct installation of `just`
- **Go-libp2p-mplex Compatibility**: Automatic patching during Docker build
- **apk/apt-get Issues**: Automatically fixed for Ubuntu images
- **Git Information Passing**: Automatic environment variable setup
- **Docker Build Cache**: Automatic cache management

### 🚀 Advantages of Current Environment
- **Complete Automation**: No manual configuration required
- **Automatic Dependency Resolution**: Automatic patching of compatibility issues
- **Development Environment Integration**: Forked repository code automatically reflected in Devnet
- **Fast Deployment**: Complete Devnet setup in 5-15 minutes

## ⚠️ Important Notes

### Devnet Status Check
```bash
# Check Devnet running status
kurtosis enclave inspect simple-devnet

# Check challenger status
docker ps | grep challenger

# Check challenger logs
docker logs op-challenger
```

### Development Environment Integration
- **Forked Repository**: `https://github.com/Zena-park/optimism`
- **Docker Build**: Local code automatically reflected in Devnet
- **Real-time Testing**: Code changes immediately testable in Devnet

## 📚 Additional Documentation
- [Devnet Guide](docs/devnet-guide.md): Detailed Devnet setup and operation guide
- [Installation Guide](docs/installation-guide.md): System requirements and installation methods
