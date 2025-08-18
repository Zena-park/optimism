# Optimism op-program 기술 가이드

## 📋 목차

1. [개요](#-개요)
2. [op-program이란?](#-op-program이란)
3. [아키텍처 및 구조](#-아키텍처-및-구조)
4. [핵심 기능](#-핵심-기능)
5. [Fault Proof 시스템에서의 역할](#-fault-proof-시스템에서의-역할)
6. [실행 모드](#-실행-모드)
7. [챌린저와의 통합](#-챌린저와의-통합)
8. [실제 사용 예제](#-실제-사용-예제)
9. [기술적 세부사항](#-기술적-세부사항)
10. [문제 해결](#-문제-해결)

---

## 🎯 개요

**op-program**은 Optimism의 **Fault Proof 시스템**에서 핵심적인 역할을 하는 컴포넌트로, L1 입력 데이터만을 사용하여 L2 출력을 검증하는 프로그램입니다. 챌린저가 dispute game에서 잘못된 상태 루트를 발견했을 때, 이를 증명하기 위해 사용됩니다.

### 🔑 핵심 개념

- **Fault Proof Program**: L1 데이터로부터 L2 상태를 재계산하는 검증 프로그램
- **Pre-image Oracle**: VM 실행 시 필요한 데이터를 제공하는 서버
- **Deterministic Execution**: 동일한 입력에 대해 항상 동일한 실행 경로 보장
- **On-chain Verification**: 온체인에서 검증 가능한 실행 증명 생성

---

## 🔍 op-program이란?

### 📖 정의

**op-program**은 Optimism L2 상태 전이(state transition)를 재실행하여 올바른 상태를 계산하는 프로그램입니다. 이 프로그램은 결정적(deterministic) 방식으로 실행되어 두 번의 동일한 입력 실행이 같은 출력과 실행 경로를 생성하도록 설계되었습니다.

### 🎯 주요 목적

1. **L2 상태 검증**: L1 데이터만으로 L2 블록 상태를 재계산
2. **Dispute 해결**: 잘못된 L2 output에 대한 반박 증명 제공
3. **온체인 검증**: VM에서 실행 가능한 결정적 프로그램 제공

### 📊 Optimism Fault Proof 생태계에서의 위치

```
Optimism Fault Proof System:
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   op-challenger │───▶│  op-program  │───▶│ Cannon/Asterisc │
│   (dispute 감지) │    │ (상태 재계산)  │    │  (VM 실행 증명)  │
└─────────────────┘    └──────────────┘    └─────────────────┘
        │                       │                     │
        ▼                       ▼                     ▼
   잘못된 output        L1→L2 상태 전이        실행 단계별 증명
   발견 및 신고              재실행                온체인 제출
```

---

## 🏗️ 아키텍처 및 구조

### 📁 소스코드 구조

```
op-program/
├── client/              ← 클라이언트 실행 로직
│   ├── cmd/main.go     ← 클라이언트 엔트리포인트
│   ├── l1/             ← L1 데이터 읽기 (블록, 트랜잭션)
│   ├── l2/             ← L2 상태 재계산 엔진
│   ├── mpt/            ← Merkle Patricia Trie 처리
│   └── program.go      ← 메인 프로그램 로직
├── host/               ← 호스트 서버 (pre-image oracle)
│   ├── cmd/main.go     ← 호스트 엔트리포인트
│   ├── config/         ← 설정 관리
│   ├── kvstore/        ← 키-값 저장소
│   └── prefetcher/     ← 데이터 프리페칭
├── verify/             ← 네트워크별 검증 로직
│   ├── mainnet/        ← 메인넷 검증
│   ├── sepolia/        ← 세폴리아 테스트넷 검증
│   └── verify.go       ← 공통 검증 로직
└── prestates/          ← 절대 사전 상태 관리
```

### 🔄 핵심 컴포넌트

#### 1. **Client (클라이언트)**
- **역할**: Cannon VM 내에서 실행되는 프로그램
- **기능**: L1 데이터를 읽어서 L2 상태를 재계산
- **실행 환경**: MIPS VM (Cannon) 내부

#### 2. **Host (호스트)**
- **역할**: Pre-image oracle 서버로 동작
- **기능**: Client가 필요한 데이터를 제공
- **실행 환경**: 호스트 시스템에서 직접 실행

#### 3. **Verify (검증)**
- **역할**: 네트워크별 검증 로직
- **기능**: 메인넷/테스트넷 설정에 따른 검증 수행

---

## ⚙️ 핵심 기능

### 1. **L2 상태 전이 재실행**

op-program은 L1 블록 데이터만을 사용하여 L2 블록을 순서대로 재처리합니다:

```go
// L2 상태 전이 재실행 과정
func (p *Program) RunL2StateTransition() error {
    // 1. L1 블록 데이터 읽기
    l1Block := p.FetchL1BlockData()
    
    // 2. L2 블록 파라미터 계산
    l2Params := p.DeriveL2Parameters(l1Block)
    
    // 3. L2 상태 전이 실행
    newState := p.ExecuteL2StateTransition(l2Params)
    
    // 4. 상태 루트 계산
    stateRoot := p.ComputeStateRoot(newState)
    
    return nil
}
```

### 2. **Pre-image Oracle 서버**

op-program은 Cannon/Asterisc VM이 실행될 때 필요한 데이터를 제공합니다:

```go
// Pre-image 타입 정의
type PreimageType uint32

const (
    LocalPreimage   PreimageType = 1  // 로컬 데이터
    Keccak256       PreimageType = 2  // Keccak256 해시
    SHA256          PreimageType = 3  // SHA256 해시  
    Blob            PreimageType = 4  // Blob 데이터
    Precompile      PreimageType = 5  // 사전 컴파일된 데이터
)

// Pre-image Oracle 인터페이스
type PreimageOracle interface {
    GetPreimage(key [32]byte) ([]byte, error)
    Hint(hint []byte) error
}
```

### 3. **결정적 실행 보장**

동일한 입력에 대해 항상 동일한 출력과 실행 경로를 생성합니다:

- **입력 고정**: L1 블록 데이터, 설정 파라미터
- **실행 순서**: 결정적 알고리즘 사용
- **출력 일치**: 상태 루트, 실행 경로 완전 일치

### 4. **데이터 소스 관리**

#### L1 데이터 읽기
```go
type L1Oracle interface {
    GetBlockByNumber(number uint64) (*types.Block, error)
    GetBlockByHash(hash common.Hash) (*types.Block, error)
    GetTransaction(hash common.Hash) (*types.Transaction, error)
    GetReceipts(blockHash common.Hash) ([]*types.Receipt, error)
}
```

#### L2 상태 계산
```go
type L2Engine interface {
    ProcessBlock(block *types.Block) (*types.Header, error)
    UpdateState(header *types.Header) error
    ComputeStateRoot() common.Hash
}
```

---

## 🎮 Fault Proof 시스템에서의 역할

### 🔄 Dispute Game 플로우

```
1. 챌린저가 잘못된 L2 output 발견
        ↓
2. Dispute game 시작 (온체인)
        ↓
3. 챌린저가 op-challenger 실행
        ↓
4. op-challenger가 Cannon VM + op-program 실행
        ↓
5. op-program이 L1 데이터만으로 L2 상태 재계산
        ↓
6. Cannon이 각 실행 단계를 기록 (단계별 증명)
        ↓
7. 실행 증명(execution proof) 생성
        ↓
8. 온체인에 proof 제출하여 dispute 해결
```

### 🔗 Cannon과의 협력 구조

1. **Cannon**: MIPS VM을 시뮬레이션하는 실행 환경
2. **op-program**: Cannon 내에서 실행되는 프로그램으로 L2 상태 재계산

```bash
# Challenger가 실제로 사용하는 방식
./op-challenger \
    --cannon-bin "./bin/cannon" \
    --cannon-server "./bin/op-program" \  # op-program이 서버로 동작
    --network sepolia \
    [기타 플래그들...]
```

### 📊 실행 흐름 상세

```
Step 1: Initialization
├── op-challenger starts dispute detection
├── Cannon VM environment setup
└── op-program server initialization

Step 2: Data Fetching  
├── op-program fetches L1 block data
├── Parse transaction and receipt data
└── Extract L2 derivation parameters

Step 3: State Transition
├── Execute L2 block construction
├── Process transactions sequentially  
├── Update L2 state incrementally
└── Compute final state root

Step 4: Proof Generation
├── Cannon records each execution step
├── Generate step-by-step execution trace
├── Create cryptographic proofs
└── Submit proof to dispute contract
```

---

## 🖥️ 실행 모드

### 1. **Client Mode** (`client/cmd/main.go`)

**클라이언트 모드**는 Cannon VM 내에서 실행되는 프로그램입니다:

```go
// client/cmd/main.go
package main

import (
    "github.com/ethereum-optimism/optimism/op-program/client"
)

func main() {
    client.Main(false)  // interop=false for standard mode
}
```

**특징:**
- **실행 환경**: MIPS VM (Cannon) 내부
- **용도**: L1 데이터를 읽어서 L2 상태를 재계산
- **입력**: L1 블록 해시, L2 클레임, 블록 번호
- **출력**: 계산된 L2 상태 루트

### 2. **Host Mode** (`host/cmd/main.go`)

**호스트 모드**는 Pre-image oracle 서버로 동작합니다:

```go
// host/cmd/main.go
package main

import (
    "github.com/ethereum-optimism/optimism/op-program/host"
)

func main() {
    args := os.Args
    if err := run(args, host.Main); err != nil {
        log.Crit("Application failed", "err", err)
    }
}
```

**특징:**
- **실행 환경**: 호스트 시스템에서 직접 실행
- **용도**: Client가 필요한 데이터를 제공하는 Oracle
- **기능**: L1 RPC 호출, 데이터 캐싱, Pre-image 제공
- **통신**: Unix socket 또는 네트워크를 통해 Client와 통신

### 3. **Interop Mode** (`client/interopcmd/main.go`)

**Interop 모드**는 다중 체인 상호 운용성을 위한 모드입니다:

```go
// client/interopcmd/main.go
package main

import (
    "github.com/ethereum-optimism/optimism/op-program/client"
)

func main() {
    client.Main(true)  // interop=true for multi-chain mode
}
```

**특징:**
- **용도**: 여러 L2 체인 간의 상호 운용성 검증
- **기능**: 다중 체인 상태 동기화 및 검증
- **실행**: Superchain 환경에서 사용

---

## 🔗 챌린저와의 통합

### 📋 CLI 플래그 통합

op-program은 op-challenger에서 다음과 같은 플래그를 통해 사용됩니다:

```bash
# Cannon 기반 trace type
--cannon-server "./bin/op-program"        # op-program 실행 파일 경로
--cannon-bin "./bin/cannon"               # Cannon VM 실행 파일 경로
--cannon-prestate "./bin/prestate.bin.gz" # 절대 사전 상태 파일

# Asterisc 기반 trace type  
--asterisc-server "./bin/op-program"      # op-program 실행 파일 경로
--asterisc-bin "./bin/asterisc"           # Asterisc VM 실행 파일 경로
--asterisc-prestate "./bin/prestate.bin.gz"
```

### 🔧 설정 통합

```go
// op-challenger/config/config.go
type Config struct {
    // Cannon 설정
    Cannon vm.Config
    CannonAbsolutePreState string
    
    // Asterisc 설정  
    Asterisc vm.Config
    AsteriscAbsolutePreState string
}

// vm.Config 구조
type Config struct {
    VmBin    string  // VM 실행 파일 (cannon/asterisc)
    Server   string  // op-program 실행 파일
    Networks []string
    L1       string  // L1 RPC URL
    L2s      []string // L2 RPC URLs
}
```

### 🚀 실행자 (Executor) 통합

```go
// op-challenger에서 op-program 실행
type OpProgramServerExecutor struct {
    logger log.Logger
}

func (e *OpProgramServerExecutor) ExecuteProgram(
    ctx context.Context,
    cfg vm.Config,
    inputs vm.Inputs,
) (*vm.ExecutionResult, error) {
    // op-program 서버 시작
    cmd := exec.CommandContext(ctx, cfg.Server, 
        "--l1", cfg.L1,
        "--l2", cfg.L2s[0],
        "--network", cfg.Networks[0],
        "--l1.head", inputs.L1Head.Hex(),
        "--l2.head", inputs.L2Head.Hex(),
        "--l2.outputroot", inputs.L2OutputRoot.Hex(),
        "--l2.claim", inputs.L2Claim.Hex(),
        "--l2.blocknumber", fmt.Sprintf("%d", inputs.L2BlockNumber),
    )
    
    return e.runProgram(cmd)
}
```

---

## 💻 실제 사용 예제

### 1. **기본 실행 예제**

#### op-program 빌드
```bash
cd op-program
make op-program
# 결과: ./bin/op-program
```

#### 단독 실행 (Host 모드)
```bash
./bin/op-program \
    --network mainnet \
    --l1 https://ethereum-rpc.publicnode.com \
    --l1.beacon https://ethereum-beacon-api.publicnode.com \
    --l2 https://mainnet.optimism.io \
    --datadir /tmp/op-program-data \
    --l1.head 0x1234... \
    --l2.head 0x5678... \
    --l2.outputroot 0x9abc... \
    --l2.claim 0xdef0... \
    --l2.blocknumber 118900000
```

### 2. **챌린저 통합 실행**

#### Cannon과 함께 사용
```bash
./op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/opt/challenger-data" \
    --cannon-bin "./bin/cannon" \
    --cannon-server "./bin/op-program" \    # op-program 경로
    --cannon-prestate "./bin/prestate.bin.gz" \
    --trace-type "cannon" \
    --game-factory-address "0x..." \
    --log.level "info"
```

#### Asterisc과 함께 사용
```bash
./op-challenger \
    --network "sepolia" \
    --l1-eth-rpc "https://ethereum-sepolia-rpc.publicnode.com" \
    --l1-beacon "https://ethereum-sepolia-beacon-api.publicnode.com" \
    --l2-eth-rpc "https://sepolia.optimism.io" \
    --rollup-rpc "https://sepolia.optimism.io" \
    --datadir "/opt/challenger-data" \
    --asterisc-bin "./bin/asterisc" \
    --asterisc-server "./bin/op-program" \   # op-program 경로
    --asterisc-prestate "./bin/prestate.bin.gz" \
    --trace-type "asterisc" \
    --game-factory-address "0x..." \
    --log.level "info"
```

### 3. **개발 및 테스트**

#### 단위 테스트 실행
```bash
cd op-program
make test
```

#### 절대 사전 상태 생성
```bash
cd op-program
make reproducible-prestate
# 결과: ./bin/prestate-*.bin.gz, ./bin/prestate-proof-*.json
```

#### 호환성 테스트
```bash
cd op-program
make test-compat
```

### 4. **Docker를 사용한 재현 가능한 빌드**

```bash
cd op-program

# 재현 가능한 Docker 빌드
docker build -f Dockerfile.repro -t op-program-repro .

# 컨테이너에서 사전 상태 추출
docker run --rm op-program-repro cat /opt/op-program/bin/prestate.bin.gz > prestate.bin.gz
```

---

## 🔧 기술적 세부사항

### 📊 Pre-image 타입 및 처리

```go
// Pre-image 키 형식
type PreimageKey [32]byte

// Pre-image 요청 처리
func (o *Oracle) GetPreimage(key PreimageKey) ([]byte, error) {
    switch key[0] {
    case byte(LocalPreimage):
        return o.getLocalPreimage(key)
    case byte(Keccak256):
        return o.getKeccakPreimage(key)
    case byte(SHA256):
        return o.getSHA256Preimage(key)
    case byte(Blob):
        return o.getBlobPreimage(key)
    case byte(Precompile):
        return o.getPrecompilePreimage(key)
    default:
        return nil, fmt.Errorf("unknown preimage type: %d", key[0])
    }
}
```

### 🔄 L1 데이터 처리 흐름

```go
// L1 블록 데이터 페칭
type L1Client struct {
    ethClient *ethclient.Client
    beaconClient *beacon.Client
}

func (c *L1Client) FetchBlockData(blockNumber uint64) (*L1BlockData, error) {
    // 1. 실행 계층 블록 데이터
    execBlock, err := c.ethClient.BlockByNumber(ctx, big.NewInt(blockNumber))
    if err != nil {
        return nil, err
    }
    
    // 2. 합의 계층 블록 데이터
    consensusBlock, err := c.beaconClient.GetBlockV2(ctx, blockNumber)
    if err != nil {
        return nil, err
    }
    
    // 3. 트랜잭션 영수증
    receipts, err := c.ethClient.BlockReceipts(ctx, execBlock.Hash())
    if err != nil {
        return nil, err
    }
    
    return &L1BlockData{
        ExecutionBlock: execBlock,
        ConsensusBlock: consensusBlock,
        Receipts: receipts,
    }, nil
}
```

### 🎯 L2 상태 전이 로직

```go
// L2 블록 파생 및 실행
func (e *L2Engine) ProcessL2Block(l1Data *L1BlockData) (*L2Block, error) {
    // 1. L1 attributes 트랜잭션 생성
    l1Attrs := e.deriveL1Attributes(l1Data)
    
    // 2. 사용자 트랜잭션 필터링 및 정렬
    userTxs := e.filterUserTransactions(l1Data)
    
    // 3. L2 블록 구성
    l2Block := &L2Block{
        Header: e.computeL2Header(l1Attrs),
        Transactions: append([]*types.Transaction{l1Attrs}, userTxs...),
    }
    
    // 4. 상태 전이 실행
    newState, err := e.executeStateTransition(l2Block)
    if err != nil {
        return nil, err
    }
    
    // 5. 상태 루트 업데이트
    l2Block.Header.Root = newState.Root()
    
    return l2Block, nil
}
```

### 🗄️ 키-값 저장소 (KV Store)

```go
// KV Store 인터페이스
type KV interface {
    Get(key []byte) ([]byte, error)
    Put(key []byte, value []byte) error
}

// 다양한 KV Store 구현
type KVStore struct {
    // 메모리 기반
    mem *MemKV
    
    // 파일 기반
    dir *DirectoryKV
    
    // Pebble DB 기반
    pebble *PebbleKV
}
```

### 📡 통신 프로토콜

```go
// Host와 Client 간 통신
type PreimageChannel struct {
    requests  chan PreimageRequest
    responses chan PreimageResponse
}

type PreimageRequest struct {
    Key [32]byte `json:"key"`
}

type PreimageResponse struct {
    Data []byte `json:"data"`
    Err  string `json:"error,omitempty"`
}
```

---

## 🔍 문제 해결

### 1. **일반적인 문제**

#### 문제: "failed to connect to L1 RPC"
```bash
# 해결방법:
# 1. L1 RPC URL 확인
--l1 https://ethereum-rpc.publicnode.com

# 2. 네트워크 연결 테스트
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  https://ethereum-rpc.publicnode.com

# 3. 방화벽 및 프록시 설정 확인
```

#### 문제: "prestate file not found"
```bash
# 해결방법:
# 1. 사전 상태 파일 생성
cd op-program
make reproducible-prestate

# 2. 파일 경로 확인
ls -la ./bin/prestate*.bin.gz

# 3. 올바른 경로 지정
--cannon-prestate "./bin/prestate.bin.gz"
```

#### 문제: "invalid L2 output root"
```bash
# 해결방법:
# 1. L2 RPC 연결 확인
--l2 https://sepolia.optimism.io

# 2. 블록 번호 및 해시 검증
# 3. 네트워크 설정 확인 (mainnet/sepolia)
--network sepolia
```

### 2. **성능 문제**

#### 메모리 사용량 최적화
```bash
# 환경변수 설정
export GOMEMLIMIT=4GB
export GOGC=100

# KV Store 설정
--datadir.type "pebble"  # 디스크 기반 저장소 사용
--datadir "/fast-ssd/op-program-data"
```

#### 실행 시간 최적화
```bash
# 병렬 처리 활성화
export GOMAXPROCS=8

# 캐시 설정
--cache.size 1000000     # 캐시 크기 증가
--prefetch.enabled true  # 데이터 프리페칭 활성화
```

### 3. **디버깅**

#### 상세 로그 활성화
```bash
# op-program 디버그 로그
./bin/op-program \
    --log.level debug \
    --log.format json \
    [기타 플래그...] \
    2>&1 | tee op-program-debug.log
```

#### 실행 경로 추적
```bash
# 시스템 호출 추적 (Linux)
strace -o op-program-trace.log ./bin/op-program [플래그들...]

# 프로파일링 (개발 시)
go tool pprof http://localhost:6060/debug/pprof/profile
```

### 4. **개발 환경 설정**

#### 로컬 테스트 환경
```bash
# 로컬 L1/L2 노드 설정
git clone https://github.com/ethereum-optimism/optimism.git
cd optimism
make devnet-up

# op-program 테스트
./bin/op-program \
    --network "devnet" \
    --l1 http://localhost:8545 \
    --l2 http://localhost:9545 \
    --datadir /tmp/op-program-devnet
```

#### 코드 수정 및 테스트
```bash
# 소스 수정 후 빌드
make clean
make op-program

# 단위 테스트
make test

# 통합 테스트
make test-integration
```

---

## 📚 참고 자료

### 🔗 관련 문서
- [Optimism Fault Proof 문서](https://docs.optimism.io/stack/protocol/fault-proofs)
- [Cannon VM 사양](https://github.com/ethereum-optimism/cannon)
- [op-challenger 가이드](./Phase-1-Operations-Guide.md)

### 🛠️ 개발 도구
- [op-program 소스코드](https://github.com/ethereum-optimism/optimism/tree/develop/op-program)
- [Optimism 모노리포](https://github.com/ethereum-optimism/optimism)
- [디버깅 도구](https://github.com/ethereum-optimism/optimism/tree/develop/op-program/scripts)

### 📊 성능 벤치마크
- [호환성 테스트 결과](https://github.com/ethereum-optimism/optimism/tree/develop/op-program/compatibility-test)
- [VM 프로파일](https://github.com/ethereum-optimism/optimism/tree/develop/op-program/vm-profiles)

---

## 📝 결론

**op-program**은 Optimism Fault Proof 시스템의 핵심 컴포넌트로서:

### 🎯 핵심 역할
- **L2 상태 검증**: L1 데이터만으로 L2 상태를 재계산
- **Dispute 해결**: 잘못된 L2 output에 대한 반박 증명 제공  
- **Oracle 서비스**: VM 실행 시 필요한 데이터 제공
- **결정적 실행**: 재현 가능한 실행 경로 보장

### 🔗 통합 생태계
- **op-challenger**: Dispute 감지 및 대응
- **Cannon/Asterisc**: VM 실행 환경 제공
- **L1/L2 노드**: 블록체인 데이터 소스

### 🚀 운영 준비성
- **프로덕션 검증**: 메인넷에서 안정적 운영 중
- **성능 최적화**: 대규모 블록 처리 가능
- **확장성**: 다양한 네트워크 지원
- **모니터링**: 완전한 로깅 및 메트릭 지원

op-program을 통해 Optimism은 **탈중앙화된 fault proof 시스템**을 구현하여 L2의 보안성과 신뢰성을 크게 향상시켰습니다! 🎉