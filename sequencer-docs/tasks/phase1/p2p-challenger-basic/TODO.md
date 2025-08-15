# RAT (Randomized Attention Test) 구현 TODO 리스트

## 📋 개요
이 문서는 RAT_INTEGRATION_PLAN.md를 기반으로 한 구체적인 구현 TODO 리스트입니다.


### 🔄 **최근 변경사항**
- **개인화된 트리거 시스템**: 기존 0.28% 고정 확률에서 각 챌린저별 트리거 비율 설정으로 변경
- **배치 기준 동적 비율**: 배치 번호를 기준으로 한 정확한 트리거 비율 시스템
- **개인 최적화**: 각 챌린저가 자신의 상황과 목표에 맞게 트리거 비율 조정 가능
- **챌린저 개선 제안**: 직접 배치 데이터 비교 기능 추가 (op-challenger-analysis.md 기반)

### 🎯 **핵심 설계 개념**

#### **1. 개인화 (Personalization)**
- 각 챌린저가 자신만의 테스트 전략 설정
- 평판점수, 리소스 상황, 네트워크 상태에 따른 개별 최적화
- 예: 평판이 낮은 챌린저는 더 자주 테스트 (10-15%), 높은 챌린저는 적게 테스트 (3-5%)

#### **2.Game 번호 기준 (Game Number-based)**
- Game 번호를 직접 사용한 트리거 비율 계산
- Game 번호 % 100 < 트리거 비율 → 트리거 실행
- 챌린저 설정: "10% 트리거 비율" = Game 번호 끝자리가 0-9일 때 트리거

#### **3. 트리거 실행 방법 (Trigger Execution Methods)**
- **Game 이벤트 모니터링 (주된 동작)**: Game Create 이벤트를 모니터링하여 자동 트리거
- **운영자 수동 트리거 (보조 동작)**: 운영자가 특정 챌린저에게 수동으로 트리거 발생

#### **4. 직접 배치 데이터 검증 (Direct Batch Data Verification)**
- L1 배치정보를 직접 확인하여 L1 BatchInbox와 L2 체인의 배치 데이터를 직접 비교
- 기존 상태 루트 검증과 함께 이중 검증 구조 구현
- 단순한 데이터 비교로 배치 오류를 즉시 감지

## 🎯 Phase 1: 핵심 기능 구현

### **1. Distributed Attention Question System** (`p2p/attention/question.go`)

#### **TODO 1.1: 기본 구조 설계**
- [ ] `AttentionQuestioner` 구조체 정의
- [ ] `QuestionConfig` 설정 구조체 정의
- [ ] `QuestionResult` 결과 구조체 정의
- [ ] 기본 인터페이스 정의


#### **TODO 1.2: P2P 챌린저 네트워크 구축**
- [ ] 챌린저 리스트 관리 시스템
  - [ ] P2P 네트워크 기반 챌린저 발견
  - [ ] 활성 챌린저 목록 유지
  - [ ] 챌린저 상태 동기화
- [ ] P2P 통신 프로토콜
  - [ ] 챌린저 간 직접 메시지 전송
  - [ ] Attention 질문/응답 메시지 형식
  - [ ] 연결 관리 및 재연결 로직

#### **TODO 1.3: 개인화된 질문 로직 구현**
- [ ] `ShouldAskAttentionQuestion()` 함수 구현 (함수명 개선됨)
  - [ ] Dispute Game 번호 기반 간단한 비율 계산 (게임 번호 % 100 < 질문 비율)
  - [ ] 개인별 질문 비율 설정 (기존 0.28% 고정 확률에서 변경)
  - [ ] 복잡한 해시 계산 제거, 단순한 모듈로 연산으로 단순화
- [ ] `SelectRandomChallenger()` 함수 구현 (함수명 개선됨)
  - [ ] 결정적 무작위성 기반 챌린저 선택 (블록 해시 + 상태 루트 + 챌린저 ID + 블록 번호 + 게임 번호)
  - [ ] SHA256 해시 기반 시드 생성 및 모듈로 연산으로 균등 분배
  - [ ] 챌린저별 독립적 선택 보장 (같은 블록/게임에서도 다른 결과)
  - [ ] 선택된 챌린저 검증 로직
  - [ ] 게임 번호 매개변수 추가 완료

#### **TODO 1.4: 챌린저 간 Attention 질문 시스템**
- [ ] Dispute Game 이벤트 발생 시 처리
  - [ ] 새로운 Dispute Game 감지 시 자동 질문 실행
  - [ ] 어텐션 비율에 따른 다른 챌린저 선택
  - [ ] 선택된 챌린저에게 attention 질문 전송
- [ ] Attention 질문 생성 시스템
  - [ ] Dispute Game 관련 질문 생성 (게임 상태, 챌린지 진행 등)
  - [ ] 챌린저 활동 관련 질문 생성 (최근 챌린지, 응답 시간 등)
  - [ ] 기타 검증 가능한 질문들
- [ ] 응답 검증 및 평판 관리
  - [ ] `ValidateAttentionResponse()` 함수 구현
  - [ ] 응답 정확성 및 속도 평가
  - [ ] 평판 점수 업데이트 시스템
- [ ] P2P 기반 질문/응답 시스템
  - [ ] P2P 네트워크를 통한 질문 전송
  - [ ] 응답 수집 및 타임아웃 처리
  - [ ] 챌린저 간 직접 통신


#### **TODO 1.5: 질문 수신 및 응답 처리 시스템**
- [ ] 메시지 수신 처리기 구현
  - [ ] `handleIncomingMessage()` 함수 구현
  - [ ] 메시지 타입별 라우팅 로직
  - [ ] Attention 질문 메시지 감지
- [ ] 질문 응답 핸들러 구현
  - [ ] `handleAttentionQuestion()` 함수 구현
  - [ ] 질문 내용 파싱 및 검증
  - [ ] 응답 생성 로직
- [ ] 응답 전송 시스템
  - [ ] `sendAttentionResponse()` 함수 구현
  - [ ] 응답 메시지 생성 및 서명
  - [ ] P2P 네트워크를 통한 응답 전송
- [ ] 타임아웃 및 재시도 로직
  - [ ] 응답 타임아웃 설정 (기본 30초)
  - [ ] 응답 실패 시 재시도 로직
  - [ ] 응답 상태 추적


#### **TODO 1.6: 수동 Attention 질문 실행**
- [ ] 운영자 수동 질문 시스템 구현
  - [ ] `ShouldAskManualQuestion()` 함수 구현
  - [ ] `AskManualQuestion()` 함수 구현
  - [ ] 수동 attention 질문 API 엔드포인트
  - [ ] 권한 관리 및 보안
- [ ] 배치 데이터 수동 검증 기능 추가
  - [ ] 특정 배치 번호에 대한 수동 검증 API
  - [ ] 배치 데이터 비교 결과 조회 API
  - [ ] 배치 검증 히스토리 관리


#### **TODO 1.7: 단위 테스트**
- [ ] `ShouldAskAttentionQuestion()` 테스트 (함수명 개선됨)
- [ ] `SelectRandomChallenger()` 테스트 (함수명 개선됨)
- [ ] `BroadcastAttentionQuestion()` 테스트
- [ ] `CollectQuestionConsensus()` 테스트
- [ ] `handleIncomingMessage()` 테스트
- [ ] `handleAttentionQuestion()` 테스트
- [ ] `sendAttentionResponse()` 테스트
- [ ] 기본 통합 테스트 (tests/ 폴더로 분리됨)
- [ ] Dispute Game 번호 기반 질문 로직 테스트
- [ ] Dispute Game 이벤트 모니터링 테스트
- [ ] 수동 질문 API 테스트
- [ ] `ExecuteAttentionQuestion()` 테스트 (TODO 1.6 구현 후)
- [ ] `ValidateAttentionResponse()` 테스트 (TODO 1.5 구현 후)


### **2. EnhancedVerifier 통합** (`op-challenger/verification/enhanced.go`)

#### **TODO 2.1: 통합 검증기 구조**
- [x] `EnhancedVerifier` 구조체 구현
- [x] 기존 상태 루트 검증기와 배치 검증기 통합
- [x] 두 검증 결과를 종합하여 최종 판단
- [x] 배치 체커 기능 추가
  - [x] `BatchChecker` 인터페이스 정의 (`op-challenger/batch/checker.go`)
  - [x] L1/L2 배치 데이터 일치성 확인
  - [x] 배치 해시 검증
  - [x] 배치 순서 검증

#### **TODO 2.2: 검증 프로세스**
- [x] 상태 루트 검증: Dispute Game과 L2 실행 결과 비교
- [x] 배치 데이터 검증: L1과 L2 배치 데이터 직접 비교
- [x] 통합 결과 결정: 두 검증 결과 종합

#### **TODO 2.3: 결과 분류**
- [x] **VALID**: 상태 루트와 배치 데이터 모두 일치
- [x] **INVALID_STATE**: 배치 데이터는 일치하지만 상태 루트 불일치
- [x] **INVALID_BATCH**: 상태 루트는 일치하지만 배치 데이터 불일치
- [x] **INVALID_BOTH**: 상태 루트와 배치 데이터 모두 불일치

#### **TODO 2.4: 단위 테스트**
- [ ] EnhancedVerifier 통합 테스트
- [ ] 검증 프로세스 테스트
- [ ] 결과 분류 테스트

## 🎯 Phase 2: 고급 기능 구현

### **4. 배치 검증 성능 최적화** (`op-challenger/batch/optimization.go`)

#### **TODO 4.1: 병렬 검증**
- [ ] 상태 루트 검증과 배치 데이터 검증을 동시에 수행
- [ ] 전체 검증 시간 단축

#### **TODO 4.2: 캐싱 전략**
- [ ] L1 배치 데이터를 캐싱하여 반복 조회 최소화
- [ ] 네트워크 요청 횟수 감소

#### **TODO 4.3: 단위 테스트**
- [ ] 병렬 검증 테스트
- [ ] 캐싱 전략 테스트

## 🧪 테스트 및 최적화

### **TODO 5.1: 통합 테스트**
- [ ] 전체 RAT 플로우 테스트
- [ ] 배치 데이터 검증 시스템 통합 테스트
- [ ] 성능 테스트

### **TODO 5.2: 성능 최적화**
- [ ] 메모리 사용량 최적화 (< 500MB)
- [ ] CPU 사용률 최적화 (< 5%)
- [ ] 배치 검증 성능 최적화 (< 1초)

## 📚 참고 문서

- [op-challenger-analysis.md](../../../op-challenger-analysis.md) - 챌린저 개선 제안 및 직접 배치 데이터 검증 시스템 설계
- [RAT_INTEGRATION_PLAN.md](../RAT_INTEGRATION_PLAN.md) - RAT 통합 계획
