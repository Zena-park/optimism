# RAT (Randomized Attention Test) 구현 TODO 리스트

## 📋 개요
이 문서는 RAT_INTEGRATION_PLAN.md를 기반으로 한 구체적인 구현 TODO 리스트입니다.

### 🔄 **최근 변경사항**
- **개인화된 트리거 시스템**: 기존 0.28% 고정 확률에서 각 챌린저별 트리거 비율 설정으로 변경
- **배치 기준 동적 비율**: 배치 번호를 기준으로 한 정확한 트리거 비율 시스템
- **개인 최적화**: 각 챌린저가 자신의 상황과 목표에 맞게 트리거 비율 조정 가능

### 🎯 **핵심 설계 개념**

#### **1. 개인화 (Personalization)**
- 각 챌린저가 자신만의 테스트 전략 설정
- 평판점수, 리소스 상황, 네트워크 상태에 따른 개별 최적화
- 예: 평판이 낮은 챌린저는 더 자주 테스트 (10-15%), 높은 챌린저는 적게 테스트 (3-5%)

#### **2. 배치 번호 기준 (Batch Number-based)**
- 배치 번호를 직접 사용한 트리거 비율 계산
- 배치 번호 % 100 < 트리거 비율 → 트리거 실행
- 챌린저 설정: "10% 트리거 비율" = 배치 번호 끝자리가 0-9일 때 트리거

#### **3. 트리거 실행 방법 (Trigger Execution Methods)**
- **L2 배치 이벤트 모니터링 (주된 동작)**: L2 배치가 L1에 제출될 때 이벤트를 모니터링하여 자동 트리거
- **운영자 수동 트리거 (보조 동작)**: 운영자가 특정 챌린저에게 수동으로 트리거 발생

#### **4. 동적 조정 (Dynamic Adjustment)**
- 실시간으로 남은 배치와 목표에 맞게 트리거 비율 조정
- 배치 번호를 기반으로 정확한 비율 달성
- 목표 달성을 보장하는 배치 기반 알고리즘

#### **5. 최적화 (Optimization)**
- 리소스와 목표에 맞는 효율적인 테스트 빈도
- 평판점수 획득 기회 극대화
- 네트워크 부하와 개인 리소스 고려한 균형잡힌 접근

## 🎯 Phase 1: 핵심 기능 구현 (1-2개월)

### **1. Distributed Attention Trigger** (`p2p/attention/trigger.go`)

#### **TODO 1.1: 기본 구조 설계**
- [x] `AttentionTrigger` 구조체 정의
- [x] `TriggerConfig` 설정 구조체 정의
- [x] `TriggerResult` 결과 구조체 정의
- [x] 기본 인터페이스 정의

#### **TODO 1.2: 개인화된 트리거 로직 구현**
- [x] `ShouldTriggerAttentionTest()` 함수 구현 (함수명 개선됨)
  - [x] 배치 번호 기반 간단한 비율 계산 (배치 번호 % 100 < 트리거 비율)
  - [x] 개인별 트리거 비율 설정 (기존 0.28% 고정 확률에서 변경)
  - [x] 복잡한 해시 계산 제거, 단순한 모듈로 연산으로 단순화
- [x] `SelectRandomChallenger()` 함수 구현 (함수명 개선됨)
  - [x] 결정적 무작위성 기반 챌린저 선택 (블록 해시 + 상태 루트 + 챌린저 ID + 블록 번호 + 배치 번호)
  - [x] SHA256 해시 기반 시드 생성 및 모듈로 연산으로 균등 분배
  - [x] 챌린저별 독립적 선택 보장 (같은 블록/배치에서도 다른 결과)
  - [x] 선택된 챌린저 검증 로직
  - [x] 배치 번호 매개변수 추가 완료

#### **TODO 1.3: 트리거 실행 방법 구현**
- [ ] L2 배치 이벤트 모니터링 시스템 구현
  - [ ] L1 컨트랙트 이벤트 구독 로직
  - [ ] 배치 번호 추출 및 트리거 판단
  - [ ] 자동 트리거 실행
- [x] 운영자 수동 트리거 시스템 구현 (구현 완료)
  - [x] `ShouldTriggerManualTest()` 함수 구현
  - [x] `TriggerManualTest()` 함수 구현
  - [ ] 수동 트리거 API 엔드포인트 (구현 필요)
  - [ ] 권한 관리 및 보안 (구현 필요)

#### **TODO 1.4: P2P 네트워크 통합**
- [ ] `BroadcastAttentionTest()` 함수 구현
  - [ ] P2P 네트워크로 테스트 전파
  - [ ] 메시지 형식 정의 및 직렬화
- [ ] `CollectTriggerConsensus()` 함수 구현
  - [ ] 51% 이상 동의 수집 로직
  - [ ] 합의 타임아웃 처리

#### **TODO 1.5: 단위 테스트**
- [x] `ShouldTriggerAttentionTest()` 테스트 (함수명 개선됨)
- [x] `SelectRandomChallenger()` 테스트 (함수명 개선됨)
- [ ] `BroadcastAttentionTest()` 테스트
- [ ] `CollectTriggerConsensus()` 테스트
- [x] 기본 통합 테스트 (tests/ 폴더로 분리됨)
- [ ] 배치 번호 기반 트리거 로직 테스트
- [ ] L2 이벤트 모니터링 테스트
- [ ] 수동 트리거 API 테스트

### **2. Hybrid Reputation Store** (`p2p/reputation/hybrid_store.go`)

#### **TODO 2.1: 로컬 저장소 구현**
- [ ] LevelDB 기반 로컬 저장소 구현
- [ ] `StoreLocalReputation()` 함수 구현
- [ ] `GetLocalReputation()` 함수 구현
- [ ] 평판 데이터 직렬화/역직렬화

#### **TODO 2.2: P2P 네트워크 동기화**
- [ ] `SyncPeerReputations()` 함수 구현
- [ ] P2P 메시지 기반 평판 동기화
- [ ] 동기화 충돌 해결 로직
- [ ] 실시간 업데이트 처리

#### **TODO 2.3: 백업 시스템**
- [ ] `BackupToL2Node()` 함수 구현
- [ ] `BackupToIPFS()` 함수 구현
- [ ] 백업 스케줄링 로직
- [ ] 백업 복원 기능

#### **TODO 2.4: 합의 메커니즘**
- [ ] `ProposeReputationUpdate()` 함수 구현
- [ ] `CollectReputationConsensus()` 함수 구현
- [ ] `ValidateReputationBlock()` 함수 구현
- [ ] 합의 실패 처리

#### **TODO 2.5: 단위 테스트**
- [ ] 로컬 저장소 테스트
- [ ] P2P 동기화 테스트
- [ ] 백업 시스템 테스트
- [ ] 합의 메커니즘 테스트

### **3. Attention Test Handler** (`p2p/challenger/attention.go`)

#### **TODO 3.1: 테스트 처리**
- [ ] `HandleIncomingTest()` 함수 구현
- [ ] 테스트 메시지 파싱 및 검증
- [ ] 테스트 실행 로직
- [ ] 응답 준비

#### **TODO 3.2: 피어 정보 수집**
- [ ] `GetPeerAttentionInfo()` 함수 구현
- [ ] `CollectPeerStatus()` 함수 구현
- [ ] `CollectPeerReputation()` 함수 구현
- [ ] `CollectPeerActivity()` 함수 구현

#### **TODO 3.3: 응답 처리**
- [ ] `SubmitTestResponse()` 함수 구현
- [ ] `ValidateOwnResponse()` 함수 구현
- [ ] 응답 검증 로직
- [ ] 응답 전파

#### **TODO 3.4: 단위 테스트**
- [ ] 테스트 처리 테스트
- [ ] 피어 정보 수집 테스트
- [ ] 응답 처리 테스트
- [ ] 통합 테스트

## 🎯 Phase 2: 고급 기능 구현 (2-3개월)

### **4. Reputation Manager** (`p2p/reputation/manager.go`)

#### **TODO 4.1: 기본 평판 관리**
- [ ] `UpdateQuestionerScore()` 함수 구현
- [ ] `UpdateResponderScore()` 함수 구현
- [ ] `GetCombinedScore()` 함수 구현
- [ ] 평판 계산 알고리즘

#### **TODO 4.2: 보상/페널티 시스템**
- [ ] `ApplyPenalty()` 함수 구현
- [ ] `ApplyReward()` 함수 구현
- [ ] 페널티/보상 규칙 정의
- [ ] 점수 조정 로직

#### **TODO 4.3: 고급 기능**
- [ ] `CalculateFairnessScore()` 함수 구현
- [ ] `DetectCollusion()` 함수 구현
- [ ] `CalculateNodeAffinity()` 함수 구현
- [ ] `CalculateHonestyScore()` 함수 구현

#### **TODO 4.4: 단위 테스트**
- [ ] 기본 평판 관리 테스트
- [ ] 보상/페널티 시스템 테스트
- [ ] 고급 기능 테스트

### **5. Peer Attention Validator** (`p2p/attention/validator.go`)

#### **TODO 5.1: 검증 로직**
- [ ] `ValidatePeerAttention()` 함수 구현
- [ ] `ValidatePeerStatus()` 함수 구현
- [ ] `ValidatePeerReputation()` 함수 구현
- [ ] `ValidatePeerActivity()` 함수 구현

#### **TODO 5.2: 정보 수집**
- [ ] `CollectPeerInfo()` 함수 구현
- [ ] `GetAttentionData()` 함수 구현
- [ ] 데이터 검증 로직

#### **TODO 5.3: 응답 검증**
- [ ] `VerifyAttentionResponse()` 함수 구현
- [ ] 응답 정확성 검증
- [ ] 응답 시간 검증

#### **TODO 5.4: 단위 테스트**
- [ ] 검증 로직 테스트
- [ ] 정보 수집 테스트
- [ ] 응답 검증 테스트

## 🎯 Phase 3: 확장 기능 구현 (3-4개월)

### **6. L2 Chain Recorder** (`p2p/l2/recorder.go`)

#### **TODO 6.1: L2 기록**
- [ ] `RecordChallengerInfo()` 함수 구현
- [ ] L2 트랜잭션 생성 및 전송
- [ ] 기록 성공/실패 처리

#### **TODO 6.2: 검증 및 합의**
- [ ] `ValidateRecordAccuracy()` 함수 구현
- [ ] `CollectRecordingConsensus()` 함수 구현
- [ ] 기록 정확성 검증

#### **TODO 6.3: 보상 이코노미**
- [ ] `PrepareRewardEconomy()` 함수 구현
- [ ] `DistributeRecordingReward()` 함수 구현
- [ ] 보상 계산 및 분배

#### **TODO 6.4: 단위 테스트**
- [ ] L2 기록 테스트
- [ ] 검증 및 합의 테스트
- [ ] 보상 이코노미 테스트

## 🧪 테스트 및 최적화

### **TODO 7.1: 통합 테스트**
- [ ] 전체 RAT 플로우 테스트
- [ ] 개인화된 트리거 시스템 통합 테스트
- [ ] 성능 테스트
- [ ] 부하 테스트
- [ ] 장애 복구 테스트

### **TODO 7.2: 성능 최적화**
- [ ] 메모리 사용량 최적화 (< 500MB)
- [ ] CPU 사용률 최적화 (< 5%)
- [ ] 네트워크 지연 최적화 (< 100ms)
- [ ] 응답 시간 최적화 (< 30초)

### **TODO 7.3: 보안 강화**
- [ ] 담합 탐지 시스템 강화
- [ ] 데이터 무결성 검증 강화
- [ ] 암호화 및 서명 검증 강화
- [ ] 공격 방지 메커니즘 구현

## 📊 모니터링 및 메트릭

### **TODO 8.1: 모니터링 시스템**
- [ ] 테스트 관련 메트릭 수집
- [ ] 평판 관련 메트릭 수집
- [ ] 네트워크 관련 메트릭 수집
- [ ] 성능 관련 메트릭 수집

### **TODO 8.2: 대시보드**
- [ ] 실시간 모니터링 대시보드
- [ ] 성능 지표 시각화
- [ ] 알림 시스템
- [ ] 로그 분석 도구

## 📚 문서화

### **TODO 9.1: 기술 문서**
- [ ] API 문서 작성
- [ ] 아키텍처 문서 작성
- [ ] 배포 가이드 작성
- [ ] 트러블슈팅 가이드 작성

### **TODO 9.2: 사용자 문서**
- [ ] 사용자 가이드 작성
- [ ] 운영 가이드 작성
- [ ] FAQ 작성
- [ ] 예제 코드 작성

## 🚀 배포 및 운영

### **TODO 10.1: 배포 준비**
- [ ] Docker 컨테이너화
- [ ] CI/CD 파이프라인 구축
- [ ] 환경별 설정 관리
- [ ] 배포 스크립트 작성

### **TODO 10.2: 운영 도구**
- [ ] 로그 수집 시스템
- [ ] 백업 및 복구 도구
- [ ] 모니터링 및 알림 시스템
- [ ] 운영 대시보드

## 📅 마일스톤

### **Week 1-2:**
- [x] Distributed Attention Trigger 기본 구조
- [ ] Hybrid Reputation Store 로컬 저장소

### **Week 3-4:**
- [ ] P2P 네트워크 통합
- [x] 기본 테스트 구현 (함수명 개선 및 테스트 폴더 분리 완료)
- [ ] 배치 기반 트리거 시스템 구현 (배치 번호 기반 비율 계산)

### **Week 5-6:**
- [ ] Attention Test Handler
- [ ] 기본 평판 관리

### **Week 7-8:**
- [ ] 통합 테스트
- [ ] 성능 최적화

### **Month 2:**
- [ ] 고급 평판 기능
- [ ] Peer Attention Validator

### **Month 3:**
- [ ] L2 Chain Recorder
- [ ] 보상 이코노미

### **Month 4:**
- [ ] 보안 강화
- [ ] 운영 도구
- [ ] 문서화 완료

## 🎯 성공 기준

### **기능적 기준:**
- [ ] 모든 핵심 기능이 정상 동작
- [ ] 성능 목표 달성
- [ ] 보안 요구사항 충족

### **품질 기준:**
- [ ] 테스트 커버리지 80% 이상
- [ ] 코드 리뷰 완료
- [ ] 문서화 완료

### **운영 기준:**
- [ ] 모니터링 시스템 구축
- [ ] 배포 파이프라인 구축
- [ ] 운영 가이드 완성
