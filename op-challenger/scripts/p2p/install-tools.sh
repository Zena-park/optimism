#!/bin/bash

# Phase 1 P2P 챌린저 네트워크 - 시스템 도구 설치 스크립트

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수
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

# Installation functions
install_go() {
    log_info "Installing Go 1.23.10..."
    if command -v mise &> /dev/null; then
        mise install go@1.23.10
        mise use go@1.23.10
        log_success "Go 1.23.10 installation completed"
    else
        log_error "Mise is not installed. Please install Mise first."
        return 1
    fi
}

install_docker() {
    log_info "Installing Docker..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew install --cask docker
            log_success "Docker installation completed (Homebrew)"
        else
            log_error "Homebrew is not installed. Download from https://docs.docker.com/get-docker/"
            return 1
        fi
    else
        log_error "Unsupported OS. Download from https://docs.docker.com/get-docker/"
        return 1
    fi
}

install_mise() {
    log_info "Installing Mise..."
    curl -fsSL https://mise.run | sh
    log_success "Mise installation completed"
    log_warning "Open a new terminal or run:"
    echo "   eval \"\$(/Users/zena/.local/bin/mise activate zsh)\""
}

install_kurtosis() {
    log_info "Installing Kurtosis..."
    if command -v mise &> /dev/null; then
        mise install kurtosis
        log_success "Kurtosis installation completed"
    else
        log_error "Mise is not installed. Please install Mise first."
        return 1
    fi
}

install_just() {
    log_info "Installing Just..."
    if command -v mise &> /dev/null; then
        mise install just
        log_success "Just installation completed"
    else
        log_error "Mise is not installed. Please install Mise first."
        return 1
    fi
}

# Main function
main() {
    echo "=========================================="
    echo "Phase 1 P2P Challenger Network - System Tools Installation"
    echo "=========================================="
    echo

    # List of tools to install
    TOOLS_TO_INSTALL=()

            # Go version check (1.23.10+ required)
    log_info "Checking Go version..."
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
        GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
        GO_PATCH=$(echo $GO_VERSION | cut -d. -f3)

        if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 23 ]) || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -eq 23 ] && [ "$GO_PATCH" -lt 10 ]); then
            log_warning "Go version update required: current $GO_VERSION, need 1.23.10+"
            log_warning "Note: Docker images use Go 1.23.8, which may cause compatibility issues"
            TOOLS_TO_INSTALL+=("go")
        else
            log_success "Go version check: $GO_VERSION (project requirement: 1.23.10+)"
        fi
    else
        log_error "Go is not installed"
        TOOLS_TO_INSTALL+=("go")
    fi

        # Docker check (24.0+ required)
    log_info "Checking Docker..."
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
        DOCKER_MAJOR=$(echo $DOCKER_VERSION | cut -d. -f1)
        DOCKER_MINOR=$(echo $DOCKER_VERSION | cut -d. -f2)

        if [ "$DOCKER_MAJOR" -lt 24 ]; then
            log_warning "Docker version update required: current $DOCKER_VERSION, need 24.0+"
            TOOLS_TO_INSTALL+=("docker")
        else
            log_success "Docker installed: $DOCKER_VERSION (container orchestration supported)"
        fi

        # Docker service status check
        if docker system info &> /dev/null; then
            log_success "Docker service is running"
        else
            log_warning "Docker service is not running"
            echo "   Solution: Start Docker Desktop"
        fi
    else
        log_error "Docker is not installed"
        TOOLS_TO_INSTALL+=("docker")
    fi

    # Mise check
    log_info "Checking Mise..."
    if command -v mise &> /dev/null; then
        MISE_VERSION=$(mise --version | head -1)
        log_success "Mise installed: $MISE_VERSION"
    else
        log_error "Mise is not installed"
        TOOLS_TO_INSTALL+=("mise")
    fi

    # Kurtosis check
    log_info "Checking Kurtosis..."
    if command -v kurtosis &> /dev/null; then
        KURTOSIS_VERSION=$(kurtosis version 2>/dev/null | grep "CLI Version" | awk '{print $3}')
        if [ -n "$KURTOSIS_VERSION" ]; then
            log_success "Kurtosis installed: CLI Version $KURTOSIS_VERSION"
        else
            log_success "Kurtosis installed"
        fi
    else
        log_error "Kurtosis is not installed"
        TOOLS_TO_INSTALL+=("kurtosis")
    fi

    # Just check
    log_info "Checking Just..."
    if command -v just &> /dev/null; then
        JUST_VERSION=$(just --version)
        log_success "Just installed: $JUST_VERSION"
    else
        log_error "Just is not installed"
        TOOLS_TO_INSTALL+=("just")
    fi

    echo

    # Check if tools need to be installed
    if [ ${#TOOLS_TO_INSTALL[@]} -eq 0 ]; then
        log_success "All required tools are already installed!"
        echo
        echo "Next steps:"
        echo "1. cd kurtosis-devnet"
        echo "2. just simple-devnet"
        return 0
    fi

    # Show tools to be installed
    echo "Tools to be installed:"
    for tool in "${TOOLS_TO_INSTALL[@]}"; do
        echo "  - $tool"
    done
    echo

    # User confirmation
    read -p "Do you want to install these tools automatically? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Installation cancelled."
        return 0
    fi

    # Install Mise first (dependency for other tools)
    if [[ " ${TOOLS_TO_INSTALL[@]} " =~ " mise " ]]; then
        install_mise
        # Update environment variables after Mise installation
        export PATH="$HOME/.local/bin:$PATH"
        if [ -f "$HOME/.zshrc" ]; then
            echo 'eval "$(/Users/zena/.local/bin/mise activate zsh)"' >> "$HOME/.zshrc"
        fi
    fi

    # Install remaining tools
    for tool in "${TOOLS_TO_INSTALL[@]}"; do
        case $tool in
            "go")
                install_go
                ;;
            "docker")
                install_docker
                ;;
            "kurtosis")
                install_kurtosis
                ;;
            "just")
                install_just
                ;;
        esac
    done

    echo
    log_success "Installation completed!"
    echo
    echo "Next steps:"
    echo "1. Open a new terminal or run:"
    echo "   source ~/.zshrc"
    echo "2. cd kurtosis-devnet"
    echo "3. just simple-devnet"
    echo
    echo "Note: If you encounter Go version compatibility issues:"
    echo "   AUTOFIX=true just simple-devnet"
    echo
    echo "Or to check system status again:"
    echo "   ./check-system.sh"
}

# Script execution
main "$@"
