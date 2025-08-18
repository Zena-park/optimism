# Optimism Challenger GUI 트레이 애플리케이션 🎉

**Optimism 챌린저를 누구나 쉽게 관리할 수 있는 혁신적인 GUI 트레이 애플리케이션**

> 복잡한 터미널 CLI 대신 직관적인 GUI 인터페이스로 Optimism 챌린저를 실행하고 관리하세요!

## 🌟 프로젝트 개요

이 애플리케이션은 **"아무나 쉽게 컨트롤 할 수 있도록"** 하는 목표로 개발된 Optimism 챌린저 GUI 관리 도구입니다. 시스템 트레이에 상주하면서 챌린저의 상태를 실시간으로 모니터링하고, 복잡한 CLI 명령어를 기억할 필요 없이 버튼 클릭만으로 모든 챌린저 기능을 제어할 수 있습니다.

### 🚀 왜 혁신적인가?

#### **📱 사용성 혁신**
- **"아무나 쉽게 컨트롤"**: 복잡한 CLI 명령어 대신 버튼 클릭으로 모든 제어
- **직관적 인터페이스**: 기술적 배경 없이도 챌린저 운영 가능
- **트레이 통합**: 백그라운드에서 상시 실행하며 필요시 즉시 접근

#### **🔧 운영 편의성**
- **설정 관리**: YAML 파일 수동 편집 대신 GUI 폼으로 간편 설정
- **실시간 피드백**: 터미널 로그 대신 통합된 로그 뷰어
- **상태 모니터링**: 챌린저 상태를 트레이와 메인 창에서 즉시 확인

#### **🚀 기술적 우수성**
- **Go 네이티브**: 단일 바이너리로 배포 가능
- **크로스플랫폼**: Windows, macOS, Linux 지원
- **경량화**: 외부 런타임 의존성 없이 독립 실행

## ✨ 주요 기능

### 🖥️ GUI 인터페이스
- **직관적인 설정 폼**: 모든 챌린저 설정을 GUI로 편리하게 구성
- **실시간 로그 뷰어**: 챌린저 실행 로그를 실시간으로 확인
- **상태 모니터링**: 챌린저 실행 상태를 한눈에 파악
- **설정 저장/로드**: YAML 기반 설정 파일 관리

### 🔧 시스템 트레이 기능
- **백그라운드 실행**: 시스템 트레이에 상주하며 백그라운드에서 동작
- **빠른 액세스**: 트레이 아이콘을 통한 빠른 시작/정지
- **컨텍스트 메뉴**: 창 보기, 시작/정지, 종료 메뉴
- **알림 기능**: 챌린저 상태 변화 시 시스템 알림

### ⚙️ 설정 관리
- **YAML 설정 파일**: 설정을 파일로 저장하고 로드
- **P2P 네트워킹 지원**: P2P 설정을 GUI로 쉽게 활성화
- **네트워크 프리셋**: 메인넷/테스트넷 설정 프리셋
- **실시간 설정 변경**: 폼을 통한 즉시 설정 변경

### 🛡️ 프로세스 관리
- **안전한 시작/정지**: SIGTERM을 통한 안전한 프로세스 종료
- **프로세스 모니터링**: 챌린저 프로세스 상태 실시간 추적
- **로그 리다이렉션**: stdout/stderr을 GUI 로그 창으로 리다이렉트
- **자동 재시작**: 옵션을 통한 자동 재시작 기능

## 🏗️ 기술 스택

- **GUI 프레임워크**: Fyne v2 (Go 네이티브 GUI)
- **시스템 트레이**: getlantern/systray
- **설정 관리**: YAML 기반 설정 파일
- **프로세스 관리**: Go 표준 라이브러리 exec 패키지
- **언어**: Go 1.23+

## 📋 시스템 요구사항

### 최소 요구사항
- **OS**: Windows 10+, macOS 10.14+, Linux (X11/Wayland)
- **메모리**: 512MB RAM
- **디스크**: 50MB 여유 공간

### 필수 의존성
- **Go**: 1.23+ (빌드 시에만 필요)
- **op-challenger**: Optimism 챌린저 바이너리
- **cannon**: Fault proof 시스템 바이너리
- **op-program**: 프로그램 서버 바이너리

## 🚀 설치 및 실행

### 1. 빌드

```bash
# Optimism 저장소 내 트레이 애플리케이션 디렉토리로 이동
cd optimism/op-challenger-tray

# 의존성 설치 및 빌드
./build.sh
```

### 2. 실행

```bash
# GUI 애플리케이션 실행
./bin/op-challenger-tray
```

### 3. 첫 실행 설정

1. **기본 설정 구성**:
   - 네트워크: sepolia (테스트넷) 또는 mainnet
   - L1/L2 RPC 엔드포인트 설정
   - 데이터 디렉토리 경로 지정

2. **바이너리 경로 설정**:
   - Cannon 바이너리 경로: `./bin/cannon`
   - op-program 서버 경로: `./bin/op-program`

3. **P2P 네트워킹 (선택사항)**:
   - P2P 활성화 체크박스 선택
   - 리슨 주소 및 네트워크 ID 설정

## 📖 사용 방법

### 메인 인터페이스

#### 기본 설정 탭
- **네트워크**: `sepolia`, `mainnet` 등
- **L1 RPC**: Ethereum L1 RPC 엔드포인트
- **L2 RPC**: Optimism L2 RPC 엔드포인트
- **데이터 디렉토리**: 챌린저 데이터 저장 경로

#### P2P 설정 탭
- **P2P 활성화**: P2P 네트워킹 on/off
- **리슨 주소**: LibP2P 멀티어드레스 (예: `/ip4/0.0.0.0/tcp/9876`)
- **네트워크 ID**: P2P 네트워크 식별자

#### 컨트롤 버튼
- **챌린저 시작**: 설정된 구성으로 챌린저 시작
- **챌린저 정지**: 실행 중인 챌린저 안전하게 정지
- **설정 저장**: 현재 설정을 YAML 파일로 저장
- **설정 로드**: 저장된 설정 파일에서 설정 불러오기

### 시스템 트레이 메뉴

- **창 보기**: 메인 GUI 창 표시
- **챌린저 시작**: 빠른 챌린저 시작
- **챌린저 정지**: 빠른 챌린저 정지
- **종료**: 애플리케이션 완전 종료

### 실시간 로그 모니터링

- **표준 출력**: `[OUT]` 접두사로 표시
- **오류 출력**: `[ERR]` 접두사로 표시
- **자동 스크롤**: 새 로그가 추가되면 자동으로 맨 아래로 스크롤

## 📁 설정 파일

설정은 다음 위치에 YAML 형식으로 저장됩니다:

**경로**: `~/.optimism/challenger/config.yaml`

**예시**:
```yaml
network: "sepolia"
l1_eth_rpc: "https://ethereum-sepolia-rpc.publicnode.com"
l1_beacon: "https://ethereum-sepolia-beacon-api.publicnode.com"
l2_eth_rpc: "https://sepolia.optimism.io"
rollup_rpc: "https://sepolia.optimism.io"
datadir: "/Users/username/.optimism/challenger/data"
p2p_enabled: true
p2p_listen_addr: "/ip4/0.0.0.0/tcp/9876"
p2p_network_id: "optimism-challenger-sepolia"
p2p_max_peers: 50
p2p_bootnodes: []
cannon_bin: "./bin/cannon"
cannon_server: "./bin/op-program"
log_level: "info"
metrics_enabled: true
metrics_addr: "0.0.0.0"
metrics_port: 7300
```

## 🔧 고급 설정

### P2P 네트워킹

P2P 네트워킹을 활성화하면 다른 챌린저들과 분산 네트워크를 구성할 수 있습니다:

```yaml
p2p_enabled: true
p2p_listen_addr: "/ip4/0.0.0.0/tcp/9876"
p2p_network_id: "optimism-challenger-sepolia"
p2p_max_peers: 50
p2p_bootnodes:
  - "/ip4/bootnode1.example.com/tcp/9876/p2p/12D3KooW..."
  - "/ip4/bootnode2.example.com/tcp/9876/p2p/12D3KooW..."
```

### 메트릭 모니터링

Prometheus 메트릭을 활성화하면 챌린저 성능을 모니터링할 수 있습니다:

```yaml
metrics_enabled: true
metrics_addr: "0.0.0.0"  # 로컬만: "127.0.0.1"
metrics_port: 7300
```

메트릭 엔드포인트: `http://localhost:7300/metrics`

### 로그 레벨 설정

```yaml
log_level: "debug"  # trace, debug, info, warn, error
```

## 💡 실제 사용 시나리오

### 시나리오 1: 첫 사용자 (초보자)
1. **GUI 실행**: 트레이 애플리케이션 더블클릭
2. **기본 설정**: 네트워크 드롭다운에서 "sepolia" 선택
3. **챌린저 시작**: "챌린저 시작" 버튼 클릭
4. **상태 확인**: 실시간 로그에서 "P2P networking started successfully" 메시지 확인
5. **트레이 최소화**: 창을 닫으면 자동으로 트레이로 최소화

### 시나리오 2: 고급 사용자 (P2P 네트워킹)
1. **P2P 활성화**: "P2P 네트워킹 활성화" 체크박스 선택
2. **네트워크 ID 설정**: "optimism-challenger-mainnet" 입력
3. **설정 저장**: "설정 저장" 버튼으로 구성 저장
4. **트레이 실행**: 트레이 메뉴에서 "챌린저 시작"
5. **분산 네트워크 참여**: 다른 챌린저들과 자동 연결

### 시나리오 3: 운영팀 (모니터링)
1. **메트릭 활성화**: 기본으로 활성화된 메트릭 확인
2. **실시간 로그**: GUI에서 챌린저 활동 모니터링
3. **Prometheus 연동**: `http://localhost:7300/metrics`에서 성능 지표 수집
4. **문제 시 대응**: 로그에서 오류 발생 시 즉시 재시작

## 🐛 문제 해결

### 자주 발생하는 문제

#### 1. "op-challenger 실행 파일을 찾을 수 없습니다"
**해결방법**:
- op-challenger 바이너리가 올바른 경로에 있는지 확인
- 실행 권한이 있는지 확인: `chmod +x ./bin/op-challenger`

#### 2. "P2P 연결 실패"
**해결방법**:
- 포트가 다른 프로세스에 의해 사용되고 있는지 확인
- 방화벽 설정에서 포트가 차단되었는지 확인
- 부트스트랩 피어 주소가 올바른지 확인

#### 3. "RPC 연결 오류"
**해결방법**:
- RPC 엔드포인트 URL이 올바른지 확인
- 네트워크 연결 상태 확인
- 공개 RPC의 경우 속도 제한에 걸렸는지 확인

### 로그 확인

GUI 애플리케이션 내의 로그 탭에서 실시간으로 로그를 확인할 수 있습니다. 추가적인 디버깅이 필요한 경우:

```bash
# 터미널에서 직접 실행하여 상세 로그 확인
./bin/op-challenger-tray
```

## 📁 프로젝트 구조

```
op-challenger-tray/
├── main.go           ← 메인 GUI 애플리케이션
├── config.go         ← YAML 설정 파일 관리
├── icon.go           ← 트레이 아이콘 리소스
├── go.mod            ← Go 모듈 정의
├── build.sh          ← 빌드 스크립트
└── README.md         ← 이 문서
```

### 핵심 컴포넌트

#### **main.go**
- GUI 인터페이스 구성
- 시스템 트레이 통합
- 프로세스 관리 및 모니터링
- 이벤트 핸들링

#### **config.go**
- YAML 설정 파일 읽기/쓰기
- 설정 구조체 정의
- 기본값 관리

#### **icon.go**
- 트레이 아이콘 리소스
- PNG 형식 아이콘 데이터

## 🚧 알려진 제한사항

1. **단일 인스턴스**: 한 번에 하나의 챌린저만 실행 가능
2. **플랫폼 의존성**: 시스템 트레이는 데스크톱 환경이 필요
3. **권한 요구사항**: 일부 포트는 관리자 권한이 필요할 수 있음

## 🔮 향후 계획 (Phase 2)

이 GUI 트레이 애플리케이션은 향후 다음과 같은 고급 기능들을 위한 견고한 기반을 제공합니다:

- [ ] **다중 챌린저 지원**: 여러 네트워크의 챌린저 동시 실행
- [ ] **설정 프리셋**: 메인넷/테스트넷 원클릭 설정
- [ ] **성능 대시보드**: 내장된 메트릭 대시보드
- [ ] **자동 업데이트**: 챌린저 바이너리 자동 업데이트
- [ ] **네트워크 진단**: 연결 상태 진단 도구
- [ ] **대시보드 통합**: 메트릭 및 성능 모니터링 대시보드
- [ ] **알림 시스템**: 중요 이벤트 시 시스템 알림

## 🎯 관련 프로젝트

이 GUI 트레이 애플리케이션은 **Phase 1 P2P 챌린저 네트워크** 프로젝트의 일부입니다:

- **P2P 인프라**: LibP2P 기반 분산 네트워크 (100% 완료)
- **CLI 통합**: 9개 P2P 플래그 완전 지원
- **성능 최적화**: 8.7M msg/sec 처리량 달성
- **완전한 테스트**: 165개 테스트 (100% 성공률)

자세한 내용은 다음 문서를 참조하세요:
- [Phase 1 구현 보고서](../sequencer-docs/tasks/phase1/p2p-challenger-basic/todo_tasks_docs/Phase-1-Implementation.md)
- [운영 가이드](../sequencer-docs/tasks/phase1/p2p-challenger-basic/todo_tasks_docs/Phase-1-Operations-Guide.md)

## 📞 지원

문제가 발생하거나 기능 요청이 있는 경우:

- **GitHub Issues**: [optimism/issues](https://github.com/ethereum-optimism/optimism/issues)
- **Discord**: Optimism 개발자 커뮤니티

## 📄 라이선스

이 프로젝트는 Optimism 저장소의 라이선스를 따릅니다.

---

**🎉 Optimism Challenger GUI Tray로 더욱 편리하게 Fault Proof 시스템에 참여하세요!**

> 이제 복잡한 터미널 명령어 대신 직관적인 GUI로 누구나 쉽게 Optimism 챌린저를 운영할 수 있습니다!