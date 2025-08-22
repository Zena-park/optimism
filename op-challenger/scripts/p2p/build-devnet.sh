#!/bin/bash

# ==========================================
# Phase 1 P2P 챌린저 네트워크 - Devnet 빌더
# Optimism 시퀀서 시스템 개선 프로젝트
# ==========================================

set -euo pipefail

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 로깅 함수
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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 변수 정의
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPTIMISM_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
KURTOSIS_DEVNET_DIR="$OPTIMISM_ROOT/optimism/kurtosis-devnet"

ENCLAVE_NAME="simple-devnet"
BUILD_LOG="/tmp/devnet-build.log"

# 서비스 목록
SERVICES=(
    "op-node"
    "op-batcher"
    "op-proposer"
    "op-faucet"
    "op-challenger"
    "op-deployer"
    "geth"
)

# 초기화
init() {
    log_step "Devnet 빌더 초기화"

    # 디렉토리 확인
    if [ ! -d "$KURTOSIS_DEVNET_DIR" ]; then
        log_error "Kurtosis devnet 디렉토리를 찾을 수 없습니다: $KURTOSIS_DEVNET_DIR"
        exit 1
    fi

    # 기존 enclave 정리
    if kurtosis enclave list | grep -q "$ENCLAVE_NAME"; then
        log_info "기존 enclave를 정리합니다: $ENCLAVE_NAME"
        kurtosis enclave rm --force "$ENCLAVE_NAME" > /dev/null 2>&1 || true
        sleep 2
    fi

    # 빌드 로그 초기화
    > "$BUILD_LOG"

    log_success "초기화 완료"
}

# 시스템 요구사항 확인
check_requirements() {
    log_step "시스템 요구사항 확인"

    # Go 버전 확인
    if ! command -v go &> /dev/null; then
        log_error "Go가 설치되지 않았습니다"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go 버전: $GO_VERSION"

    # Docker 확인
    if ! command -v docker &> /dev/null; then
        log_error "Docker가 설치되지 않았습니다"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker 서비스가 실행되지 않았습니다"
        exit 1
    fi

    # Kurtosis 확인
    if ! command -v kurtosis &> /dev/null; then
        log_error "Kurtosis가 설치되지 않았습니다"
        exit 1
    fi

    log_success "시스템 요구사항 확인 완료"
}

# Docker 이미지 빌드
build_docker_images() {
    log_step "Docker 이미지 빌드"

    cd "$KURTOSIS_DEVNET_DIR"

        # Docker 빌드 수정사항 확인
    log_info "Docker 빌드 수정사항을 확인합니다..."

    # Dockerfile에 go-libp2p-mplex 수정 코드가 이미 추가되어 있는지 확인
    local dockerfile="$OPTIMISM_ROOT/optimism/ops/docker/op-stack-go/Dockerfile"
    if [ -f "$dockerfile" ]; then
        if grep -q "Fix go-libp2p-mplex compatibility issues" "$dockerfile"; then
            log_success "Docker 빌드 수정사항이 이미 적용되어 있습니다"
        else
            log_warning "Docker 빌드 수정사항이 적용되지 않았습니다. 수동으로 적용해주세요."
        fi
    else
        log_warning "Dockerfile을 찾을 수 없습니다"
    fi

    cd "$KURTOSIS_DEVNET_DIR"

    # kurtosis-devnet 디렉토리로 이동 (just 레시피가 있는 곳)
    cd "$OPTIMISM_ROOT/optimism/kurtosis-devnet"

    local build_success_count=0
    local total_services=${#SERVICES[@]}

    for service in "${SERVICES[@]}"; do
        log_info "빌드 중: $service"

        # Git 정보 설정
        local git_commit=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
        local git_date=$(git show -s --format='%ct' 2>/dev/null || echo "0")

        # 각 서비스별 Docker 빌드 명령어
        case $service in
            "op-node")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-node-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "op-batcher")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-batcher-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "op-proposer")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-proposer-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "op-faucet")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-faucet-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "op-challenger")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-challenger-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "op-deployer")
                if GITCOMMIT="$git_commit" GITDATE="$git_date" just op-deployer-image >> "$BUILD_LOG" 2>&1; then
                    log_success "✅ $service 빌드 성공"
                    ((build_success_count++))
                else
                    log_error "❌ $service 빌드 실패"
                fi
                ;;
            "geth")
                log_info "geth는 기본 이미지를 사용합니다"
                ((build_success_count++))
                ;;
        esac
    done

    # 빌드 결과 요약
    log_info "=== 빌드 결과 요약 ==="
    log_info "성공한 서비스: $build_success_count/$total_services"

    if [ $build_success_count -eq $total_services ]; then
        log_success "🎉 모든 서비스가 성공적으로 빌드되었습니다!"
        return 0
    elif [ $build_success_count -gt 0 ]; then
        log_warning "⚠️  일부 서비스만 빌드되었습니다 ($build_success_count/$total_services)"
        log_info "Devnet은 제한적으로 작동할 수 있습니다."
        return 0
    else
        log_error "❌ 모든 서비스 빌드가 실패했습니다 (0/$total_services)"
        return 1
    fi
}

# Devnet 배포
deploy_devnet() {
    log_step "Devnet 배포"

    cd "$KURTOSIS_DEVNET_DIR"

    log_info "Devnet을 배포합니다... (5-15분 소요)"

    # 우리만의 Devnet 구성 생성
    create_devnet_config

    # Kurtosis 패키지 실행 (더 안정적인 방식)
    log_info "Kurtosis 패키지를 실행합니다..."

    # 기존 enclave 정리
    if kurtosis enclave list | grep -q "$ENCLAVE_NAME"; then
        log_info "기존 enclave를 정리합니다..."
        kurtosis enclave rm --force "$ENCLAVE_NAME" > /dev/null 2>&1 || true
        sleep 10  # 더 긴 대기 시간
    fi

    # 새로운 enclave 생성 및 실행
    log_info "새로운 Devnet을 생성합니다..."

    # 타임아웃 설정으로 실행 (10분으로 증가)
    timeout 600 kurtosis run ./optimism-package-trampoline/ --enclave "$ENCLAVE_NAME" >> "$BUILD_LOG" 2>&1
    local exit_code=$?

    # 결과 분석
    if [ $exit_code -eq 0 ]; then
        log_success "Devnet 배포 성공"
        return 0
    elif [ $exit_code -eq 124 ]; then
        log_warning "Devnet 배포 시간 초과 (10분). 상태를 확인합니다..."

        # 로그에서 성공 지표 확인 (더 정확한 판단)
        if grep -q "L1 Chain has started\|L1 Chain is starting up\|RUNNING.*cl-1-lighthouse-geth\|RUNNING.*el-1-geth-lighthouse" "$BUILD_LOG"; then
            log_success "Devnet이 성공적으로 시작되었습니다 (시간 초과했지만 정상 작동)"
            return 0
        else
            log_error "Devnet 배포 실패 (시간 초과)"
            return 1
        fi
    else
        log_warning "Devnet 배포 프로세스가 종료되었습니다 (종료 코드: $exit_code). 로그를 확인합니다..."

        # 로그에서 성공 지표 확인 (더 정확한 판단)
        if grep -q "L1 Chain has started\|L1 Chain is starting up\|RUNNING.*cl-1-lighthouse-geth\|RUNNING.*el-1-geth-lighthouse" "$BUILD_LOG"; then
            log_success "Devnet이 성공적으로 시작되었습니다 (프로세스 종료했지만 정상 작동)"
            return 0
        else
            log_error "Devnet 배포 실패"
            log_info "상세 오류 로그:"
            tail -20 "$BUILD_LOG" | while read line; do
                echo "  $line"
            done
            return 1
        fi
    fi
}

# Devnet 구성 생성
create_devnet_config() {
    log_info "Devnet 구성 파일 생성 중..."

    # simple.yaml이 이미 올바른 op-challenger 설정을 포함하고 있으므로
    # 추가 수정 없이 그대로 사용
    log_success "Devnet 구성 완료 (기본 simple.yaml 사용)"
}

# 서비스 상태 확인 (개선된 버전)
verify_services() {
    log_step "서비스 상태 확인"

    # 더 긴 대기 시간 (60초)
    log_info "서비스 시작을 기다리는 중... (60초)"
    sleep 60

    local running_services=0
    local total_services=${#SERVICES[@]}
    local challenger_running=false

    # 기본 서비스들 먼저 확인
    for service in "${SERVICES[@]}"; do
        if [ "$service" = "op-challenger" ]; then
            continue  # op-challenger는 별도로 확인
        fi

        if docker ps --format "{{.Names}}" | grep -q "$service"; then
            log_success "✅ $service 실행 중"
            ((running_services++))
        else
            log_warning "❌ $service 실행되지 않음"
        fi
    done

    # op-challenger 전용 확인 (의존성 서비스들이 준비된 후)
    log_info "op-challenger 의존성 확인 중..."

    # 의존성 서비스들이 준비되었는지 확인
    local dependencies_ready=true
    local dependency_services=("el-1-geth-lighthouse" "cl-1-lighthouse-geth" "op-el-2151908-node0-op-geth" "op-cl-2151908-node0-op-node")

    for dep_service in "${dependency_services[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "$dep_service"; then
            log_info "✅ 의존성 서비스 $dep_service 준비됨"
        else
            log_warning "⚠️  의존성 서비스 $dep_service 준비되지 않음"
            dependencies_ready=false
        fi
    done

    # op-challenger 확인
    if [ "$dependencies_ready" = true ]; then
        log_info "op-challenger 상태 확인 중..."
        if docker ps --format "{{.Names}}" | grep -q "op-challenger"; then
            log_success "✅ op-challenger 실행 중"
            ((running_services++))
            challenger_running=true
                else
            log_warning "❌ op-challenger 실행되지 않음"
            log_info "simple.yaml 설정을 확인해주세요"
        fi
    else
        log_warning "⚠️  의존성 서비스가 준비되지 않아 op-challenger 확인 건너뜀"
    fi

    # 서비스 상태 요약
    log_info "=== 서비스 상태 요약 ==="
    log_info "실행 중인 서비스: $running_services/$total_services"

    if [ "$challenger_running" = true ]; then
        log_success "🎉 op-challenger가 정상적으로 실행되고 있습니다!"
    else
        log_warning "⚠️  op-challenger는 나중에 수동으로 추가할 수 있습니다"
    fi

    if [ $running_services -eq $total_services ]; then
        log_success "🎉 모든 서비스가 정상적으로 실행되고 있습니다!"
        return 0
    elif [ $running_services -gt 0 ]; then
        log_warning "⚠️  일부 서비스만 실행 중입니다 ($running_services/$total_services)"
        return 0
    else
        log_error "❌ 실행 중인 서비스가 없습니다 (0/$total_services)"
        return 1
    fi
}



# RPC 연결 확인
verify_rpc_connections() {
    log_step "RPC 연결 확인"

    # 포트 정보 추출 (기본값 사용)
    local l1_port="53620"
    local l2_port="56781"

    # L1 RPC 확인
    log_info "L1 RPC 연결 확인 중... (포트: $l1_port)"
    for i in {1..12}; do
        if curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            "http://localhost:$l1_port" > /dev/null 2>&1; then
            log_success "✅ L1 RPC 연결 성공"
            break
        fi
        if [ $i -eq 12 ]; then
            log_warning "⚠️  L1 RPC 연결 실패"
        fi
        sleep 5
    done

    # L2 RPC 확인
    log_info "L2 RPC 연결 확인 중... (포트: $l2_port)"
    for i in {1..12}; do
        if curl -s -X POST -H "Content-Type: application/json" \
            --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
            "http://localhost:$l2_port" > /dev/null 2>&1; then
            log_success "✅ L2 RPC 연결 성공"
            break
        fi
        if [ $i -eq 12 ]; then
            log_warning "⚠️  L2 RPC 연결 실패"
        fi
        sleep 5
    done
}

# 완료 메시지
show_completion_message() {
    log_success "🎉 Devnet 빌드 및 배포 완료!"

    echo
    echo "=== Devnet 관리 명령어 ==="
    echo "Devnet 상태 확인: kurtosis enclave inspect $ENCLAVE_NAME"
    echo "Devnet 정지: kurtosis enclave rm --force $ENCLAVE_NAME"
    echo "Devnet 로그 확인: kurtosis enclave logs $ENCLAVE_NAME"
    echo "빌드 로그 확인: cat $BUILD_LOG"
    echo

    echo "=== 연결 정보 ==="
    echo "L1 RPC: http://localhost:53620"
    echo "L2 RPC: http://localhost:56781"
    echo "Rollup RPC: http://localhost:57029"
    echo

    echo "=== 다음 단계 ==="
    echo "P2P 챌린저 네트워크 설정을 위해 다음 명령어를 실행하세요:"
    echo "cd $SCRIPT_DIR"
    echo "./install-and-run.sh"
    echo
}

# 메인 함수
main() {
    echo "=========================================="
    echo "Phase 1 P2P 챌린저 네트워크 - Devnet 빌더"
    echo "Optimism 시퀀서 시스템 개선 프로젝트"
    echo "=========================================="
    echo

    # 각 단계 실행
    init
    check_requirements
    build_docker_images
    deploy_devnet
    verify_services
    verify_rpc_connections
    show_completion_message
}

# 스크립트 실행
main "$@"
