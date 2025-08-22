#!/bin/bash

# 로컬 Devnet 전용 P2P 챌린저 실행 스크립트
# 개발 및 테스트 목적으로만 사용

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

# 스크립트 정보
echo "=========================================="
echo "로컬 Devnet P2P 챌린저 실행"
echo "개발 및 테스트 전용"
echo "=========================================="
echo

# 기본 설정
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPTIMISM_ROOT="$(dirname "$(dirname "$(dirname "$SCRIPT_DIR")")")"

# Devnet 상태 확인
check_devnet_status() {
    log_info "Devnet 상태 확인 중..."

    if ! kurtosis enclave ls | grep -q "simple-devnet"; then
        log_error "Devnet이 실행되지 않고 있습니다."
        log_info "먼저 Devnet을 시작해주세요:"
        log_info "  ./build-devnet.sh"
        exit 1
    fi

    log_success "Devnet이 실행 중입니다"
}

# op-challenger Docker 이미지 확인
check_challenger_image() {
    log_info "op-challenger Docker 이미지 확인 중..."

    if ! docker images | grep -q "op-challenger.*devnet"; then
        log_error "op-challenger Docker 이미지가 없습니다."
        log_info "먼저 Docker 이미지를 빌드해주세요:"
        log_info "  ./build-devnet.sh"
        exit 1
    fi

    log_success "op-challenger Docker 이미지가 준비되었습니다"
}

# Devnet 포트 정보 가져오기
get_devnet_ports() {
    log_info "Devnet 포트 정보 확인 중..."

    # kurtosis enclave inspect에서 포트 정보 추출
    local enclave_info=$(kurtosis enclave inspect simple-devnet 2>/dev/null)

            # L1 RPC 포트 (el-1-geth-lighthouse)
    L1_RPC_PORT=$(echo "$enclave_info" | grep "rpc: 8545/tcp" | head -1 | sed 's/.*-> //' | sed 's/.*://' | tr -d ' ')

    # L2 RPC 포트 (op-el-2151908-node0-op-geth)
    L2_RPC_PORT=$(echo "$enclave_info" | grep "rpc: 8545/tcp" | tail -1 | sed 's/.*-> //' | sed 's/.*://' | tr -d ' ')

    # Rollup RPC 포트 (op-cl-2151908-node0-op-node)
    ROLLUP_RPC_PORT=$(echo "$enclave_info" | grep "rpc: 8547/tcp" | sed 's/.*-> //' | sed 's/.*://' | tr -d ' ')

    # 기본값 설정
    L1_RPC_PORT="${L1_RPC_PORT:-53620}"
    L2_RPC_PORT="${L2_RPC_PORT:-56781}"
    ROLLUP_RPC_PORT="${ROLLUP_RPC_PORT:-57029}"

    log_success "포트 정보 확인 완료"
    log_info "L1 RPC: http://localhost:$L1_RPC_PORT"
    log_info "L2 RPC: http://localhost:$L2_RPC_PORT"
    log_info "Rollup RPC: http://localhost:$ROLLUP_RPC_PORT"
}

# op-challenger 실행
run_challenger() {
    log_info "op-challenger 상태 확인 중..."

    # 1. 컨테이너가 실행 중인지 확인
    if docker ps --format "{{.Names}}" | grep -q "op-challenger"; then
        log_success "op-challenger가 이미 실행 중입니다"
        return 0
    fi

    # 2. 컨테이너가 중지된 상태인지 확인
    if docker ps -a --format "{{.Names}}" | grep -q "op-challenger"; then
        log_info "중지된 op-challenger 컨테이너를 시작합니다..."
        if docker start op-challenger; then
            log_success "op-challenger가 성공적으로 시작되었습니다"
            return 0
        else
            log_error "op-challenger 시작 실패"
            exit 1
        fi
    fi

    # 3. 새로운 컨테이너 실행
    log_info "새로운 op-challenger 컨테이너를 실행합니다..."
    docker run -d \
        --name op-challenger \
        --network host \
        op-challenger:devnet \
        --l1-eth-rpc "http://localhost:$L1_RPC_PORT" \
        --l2-eth-rpc "http://localhost:$L2_RPC_PORT" \
        --rollup-rpc "http://localhost:$ROLLUP_RPC_PORT" \
        --p2p-enabled \
        --p2p-network-id "optimism-challenger-devnet" \
        --p2p-max-peers 20 \
        --log.level INFO

    if [ $? -eq 0 ]; then
        log_success "op-challenger가 성공적으로 시작되었습니다"
    else
        log_error "op-challenger 시작 실패"
        exit 1
    fi
}

# 서비스 상태 확인
verify_services() {
    log_info "서비스 상태 확인 중..."

    # 잠시 대기
    sleep 10

    # op-challenger 상태 확인
    if docker ps --format "{{.Names}}" | grep -q "op-challenger"; then
        log_success "✅ op-challenger 실행 중"
    else
        log_warning "❌ op-challenger 실행되지 않음"
    fi

    # RPC 연결 확인
    log_info "RPC 연결 확인 중..."

    # L1 RPC 확인
    if curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        "http://localhost:$L1_RPC_PORT" > /dev/null 2>&1; then
        log_success "✅ L1 RPC 연결 성공"
    else
        log_warning "⚠️  L1 RPC 연결 실패"
    fi

    # L2 RPC 확인
    if curl -s -X POST -H "Content-Type: application/json" \
        --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
        "http://localhost:$L2_RPC_PORT" > /dev/null 2>&1; then
        log_success "✅ L2 RPC 연결 성공"
    else
        log_warning "⚠️  L2 RPC 연결 실패"
    fi
}

# 완료 메시지
show_completion_message() {
    log_success "🎉 로컬 Devnet P2P 챌린저 실행 완료!"

    echo
    echo "=== 연결 정보 ==="
    echo "L1 RPC: http://localhost:$L1_RPC_PORT"
    echo "L2 RPC: http://localhost:$L2_RPC_PORT"
    echo "Rollup RPC: http://localhost:$ROLLUP_RPC_PORT"
    echo

    echo "=== 관리 명령어 ==="
    echo "챌린저 상태 확인: docker ps | grep challenger"
    echo "챌린저 로그 확인: docker logs op-challenger"
    echo "챌린저 중지: docker stop op-challenger"
    echo "챌린저 제거: docker rm op-challenger"
    echo

    echo "=== 다음 단계 ==="
    echo "P2P 챌린저 네트워크가 준비되었습니다!"
    echo "이제 P2P 개발을 진행할 수 있습니다."
    echo
}

# 메인 함수
main() {
    # 각 단계 실행
    check_devnet_status
    check_challenger_image
    get_devnet_ports
    run_challenger
    verify_services
    show_completion_message
}

# 스크립트 실행
main "$@"
