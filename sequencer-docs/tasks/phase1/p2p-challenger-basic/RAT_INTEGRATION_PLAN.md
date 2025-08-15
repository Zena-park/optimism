# RAT (Randomized Attention Test) 통합 계획

## 📋 개요

기존 P2P 챌린저 기본 구현에 RAT (Randomized Attention Test) 프로토콜을 확장하여 챌린저들이 서로에게 지속적으로 어텐션하고 있는지 확인하고, Optimistic Rollups의 검증자 무임승차 문제를 해결합니다.

## 🎯 RAT 프로토콜 기본 플로우 (컨트랙트 없이, RAT Controller 없이)

### **1단계: 챌린저 네트워크 참여**
- 챌린저들이 P2P 네트워크에 참여
- 서로를 발견하고 연결 (Discovery Service)
- 챌린저 등록 및 상태 공유

### **2단계: 어텐션 테스트 트리거**
- **모든 챌린저**가 동일한 결정적 로직으로 확률적 트리거 결정 (πa = 0.28% per epoch)
- **결정적 무작위성**: 블록 해시 + 상태 루트 + 챌린저 ID 기반으로 특정 챌린저 선택
- 선택된 챌린저에게 **다른 챌린저들의 어텐션 확인 테스트** 출제
- 테스트에는 **다른 챌린저들의 상태, 평판, 활동성**에 대한 질문 포함
- **다른 챌린저들에게 어텐션하고 있지 않으면 답할 수 없는 문제**로 무임승차 방지

### **3단계: 챌린저 응답**
- 선택된 챌린저가 **다른 챌린저들 어텐션 테스트**에 응답
- **다른 챌린저들의 상태, 평판, 활동성 정보**를 기반으로 답변
- **다른 챌린저들에게 어텐션하고 있다면** 올바른 답변 제출 가능
- **다른 챌린저들에게 어텐션하지 않았다면** 답할 수 없어 평판 하락

### **4단계: 결과 처리 및 양방향 평판 관리**
- **응답자 평판**: 응답하지 않거나 잘못된 응답 시 점수 하락, 정확한 응답 시 점수 상승
- **질문자 평판**: 적극적으로 테스트를 수행하고 질문을 할수록 점수 상승
- **종합 평판**: 질문자와 응답자 점수를 종합하여 네트워크 내 지위 결정
- **공정성 보장**: 모든 챌린저가 동등한 기회로 테스트를 받고 수행

### 추후 고려사항.
- 챌린저들의 주소정보와 평판점수가 주기적으로 기록이 되어야 함.
- 모니터링하고 있는 L2에 기록하는 것이 어떤지.. (가격저렴)
- 누가 기록할것이면. 그것이 정확한지에 대한 합의 필요.
- 온체인에 기록된 정보를 바탕으로 추후 보상 지급에 대한 이코노미 설계 (L2에 트랜잭션 실행해서 등록한 등록자에게 추가 보상고려)


## 🔧 구현해야 할 모듈들

### **1. Distributed Attention Trigger** (`p2p/attention/trigger.go`)
       ```go
       // 핵심 기능:
       - ShouldTriggerTest() // 모든 챌린저가 동일한 결정적 로직으로 트리거 판단
       - SelectTargetChallenger() // 블록 해시 + 상태 루트 + 챌린저 ID 기반 무작위 선택
       - BroadcastAttentionTest() // P2P 네트워크로 어텐션 테스트 전파
       - CollectTriggerConsensus() // 트리거 합의 수집 (51% 이상 동의 필요)
       ```

### **2. Hybrid Reputation Store** (`p2p/reputation/hybrid_store.go`)
       ```go
       // 핵심 기능:
       - StoreLocalReputation() // 로컬 LevelDB에 평판 데이터 저장
       - SyncPeerReputations() // P2P 네트워크를 통한 실시간 평판 동기화
       - BackupToL2Node() // L2 노드에 백업 (저비용, 높은 가용성)
       - BackupToIPFS() // 주기적 IPFS 백업 (장기 보관)
       - ProposeReputationUpdate() // 평판 변경 제안
       - CollectReputationConsensus() // 평판 변경 합의 수집
       - ValidateReputationBlock() // 평판 블록 검증
       - GetConsensusReputation() // 합의된 평판 데이터 조회
       - RestoreFromL2Backup() // L2 노드 백업에서 복원
       - RestoreFromIPFSBackup() // IPFS 백업에서 복원
       - RecordToL2Chain() // L2 체인에 평판 정보 기록 (추후 보상용)
       ```

### **3. Attention Test Handler** (`p2p/challenger/attention.go`)
       ```go
       // 핵심 기능:
       - HandleIncomingTest() // 수신된 어텐션 테스트 처리
       - GetPeerAttentionInfo() // 다른 챌린저들의 어텐션 정보 수집
       - SubmitTestResponse() // 어텐션 테스트 응답 제출
       - ValidateOwnResponse() // 자신의 응답 검증
       - CollectPeerStatus() // 다른 챌린저들의 상태 정보 수집
       - CollectPeerReputation() // 다른 챌린저들의 평판 정보 수집
       - CollectPeerActivity() // 다른 챌린저들의 활동성 정보 수집
       ```

### **4. Reputation Manager** (`p2p/reputation/manager.go`)
       ```go
       // 핵심 기능:
       - UpdateQuestionerScore() // 질문자 평판 점수 업데이트 (적극성, 질문 빈도)
       - UpdateResponderScore() // 응답자 평판 점수 업데이트 (응답성, 정확성, 속도)
       - GetCombinedScore() // 종합 평판 점수 조회
       - ApplyPenalty() // 페널티 적용 (응답 실패, 부정확한 응답)
       - ApplyReward() // 보상 적용 (정확한 응답, 적극적 참여)
       - CalculateFairnessScore() // 공정성 점수 계산
       - DetectCollusion() // 담합 탐지 및 위험도 계산
       - CalculateNodeAffinity() // 노드 간 밀접성 계산
       - SeparateColludingGroups() // 담합 그룹 분리
       - CalculateHonestyScore() // 정직성 점수 계산
       - DistributeSynergy() // 시뇨리지 분배 (정직성/성과 기반)
       - ValidateWithMultiLayer() // 다중 검증 계층을 통한 검증
       - RecordToL2ForReward() // L2에 기록하여 추후 보상 지급 준비
       ```

### **5. Peer Attention Validator** (`p2p/attention/validator.go`)
       ```go
       // 핵심 기능:
       - ValidatePeerAttention() // 다른 챌린저들의 어텐션 검증
       - CollectPeerInfo() // 다른 챌린저들의 정보 수집
       - VerifyAttentionResponse() // 어텐션 응답 검증
       - GetAttentionData() // 어텐션 데이터 조회
       - ValidatePeerStatus() // 다른 챌린저들의 상태 검증
       - ValidatePeerReputation() // 다른 챌린저들의 평판 검증
       - ValidatePeerActivity() // 다른 챌린저들의 활동성 검증
       ```

### **6. L2 Chain Recorder** (`p2p/l2/recorder.go`)
       ```go
       // 핵심 기능:
       - RecordChallengerInfo() // 챌린저 주소정보와 평판점수를 L2에 기록
       - ValidateRecordAccuracy() // 기록된 정보의 정확성 검증
       - CollectRecordingConsensus() // 기록에 대한 합의 수집
       - PrepareRewardEconomy() // 온체인 기록 기반 보상 이코노미 준비
       - DistributeRecordingReward() // L2 트랜잭션 실행 등록자에게 추가 보상
       ```

## 🚀 구현 순서

1. **Distributed Attention Trigger** (완전 분산화된 어텐션 테스트 트리거)
2. **Hybrid Reputation Store** (P2P + L2 노드 + IPFS 하이브리드 저장 시스템)
3. **Attention Test Handler** (어텐션 테스트 처리)
4. **Reputation Manager** (평판 관리)
5. **Peer Attention Validator** (다른 챌린저 어텐션 검증)
6. **L2 Chain Recorder** (L2 체인 기록 및 보상 이코노미)

## 📊 현재 구현 상태

### ✅ 완료된 것:
- P2P 네트워크 기본 인프라 (`network/`)
- 메시지 타입 및 구조체 (`messages.go`)
- 기본 통신 프로토콜
- 포괄적인 단위 테스트
- RAT 프로토콜 문서

### ❌ 구현 예정:
- Distributed Attention Trigger
- Hybrid Reputation Store
- Attention Test Handler
- Reputation Manager
- Peer Attention Validator
- L2 Chain Recorder

## 🎯 성능 목표

### **설정 기준:**
- **RAT 논문 기반**: "Looking for Attention: Randomized Attention Test Design for Validator Monitoring in Optimistic Rollups"
- **P2P 네트워크 표준**: Bitcoin, Ethereum P2P 네트워크 성능 기준
- **경량 노드 요구사항**: 리소스 효율적인 P2P 노드 설계
- **실시간 통신**: P2P 네트워크 환경에서의 실시간 메시지 처리

### **목표 지표:**
- **테스트 트리거 확률**: 0.28% per block (약 4-6 tests/hour, 블록당 12초 기준)
- **응답 시간**: < 30초 (P2P 네트워크 내 실시간 응답)
- **네트워크 지연**: < 100ms (P2P 네트워크 지연 시간)
- **메모리 사용량**: < 500MB (1000 챌린저 기준, P2P 노드 최적화)
- **CPU 사용률**: < 5% (정상 상태, 백그라운드 동작)
- **등록 비용**: $0 (P2P 네트워크 내 무료 참여)
- **운영 비용**: 월 $20-100 (최적화된 서버 운영 비용)
- **L2 가스비**: L1 대비 1/500 수준 ($0.01-1 per transaction)
- **평판 업데이트 비용**: $0.1-0.4 per update
- **동시 처리**: 100+ 챌린저 동시 테스트 처리 가능

## 🔒 보안 고려사항

- **무작위성**: 블록 해시 + 상태 루트 + 챌린저 ID 기반 결정적 무작위성
- **평판 조작 방지**: 다중 검증자 합의
- **재전송 공격 방지**: Nonce 기반 메시지 검증
- **서명 검증**: 모든 중요 메시지에 디지털 서명
- **담합 방지**: 노드 간 밀접성 모니터링 및 동적 그룹 분리
- **패턴 분석**: 응답 패턴 일치도 및 시간적 연관성 분석
- **위험도 평가**: 실시간 담합 위험도 계산 및 대응
- **분산화**: 모든 챌린저가 동등한 권한으로 트리거 결정
- **데이터 무결성**: IPFS 해시 기반 백업 검증
- **백업 보안**: 암호화된 평판 데이터 백업

## 📈 모니터링 및 메트릭

- **테스트 관련**: 테스트 트리거 빈도, 응답 성공률, 평균 응답 시간
- **평판 관련**: 질문자/응답자 평판 점수 분포, 평판 변화 추이
- **네트워크 관련**: 참여도, 공정성 지수, 신뢰도 지수
- **성능 관련**: 시스템 성능 지표, 메모리/CPU 사용량
- **보안 관련**: 부정행위 탐지, 이상 패턴 분석, 담합 위험도 모니터링

## 🔗 기존 P2P 챌린저 시스템과의 통합

### **Discovery Service 확장**
- 기존 노드 발견 기능에 챌린저 등록 기능 추가
- 챌린저 상태 모니터링 및 업데이트

### **메시지 시스템 확장**
- RAT 관련 메시지 타입 추가 완료
- 어텐션 테스트 메시지 처리 로직 구현 필요

### **네트워크 관리 확장**
- 챌린저 평판 기반 네트워크 관리
- 비활성 챌린저 자동 제거

### **L2 가스비 최적화**
- 낮은 가스비로 빈번한 평판 업데이트 가능
- 실시간 평판 동기화 및 검증
- 비용 효율적인 RAT 테스트 실행

## 📝 다음 단계

1. **Distributed Attention Trigger 구현** - 완전 분산화된 어텐션 테스트 트리거 시스템
2. **Hybrid Reputation Store 구현** - P2P + L2 노드 + IPFS 하이브리드 저장 시스템
3. **Attention Test Handler 구현** - 어텐션 테스트 수신 및 응답 처리
4. **Reputation Manager 구현** - 평판 점수 관리 시스템
5. **Peer Attention Validator 구현** - 다른 챌린저 어텐션 검증 시스템
6. **L2 Chain Recorder 구현** - L2 체인 기록 및 보상 이코노미 시스템
7. **통합 테스트** - 전체 RAT 플로우 테스트
8. **성능 최적화** - 메모리 및 CPU 사용량 최적화
