# P2P 챌린저 네트워크 스크립트

## 📁 스크립트 위치 변경

**P2P 챌린저 네트워크 스크립트들이 `op-challenger/scripts/p2p/` 디렉토리로 이동되었습니다.**

### 새로운 위치
```
op-challenger/scripts/p2p/
├── install-and-run.sh    # 완전 자동화 설치 스크립트
├── quick-test.sh         # 빠른 테스트 스크립트
├── manage.sh             # 일상 관리 스크립트
└── README.md             # 사용법 가이드
```

### 사용법

#### 1. 완전 자동화 설치
```bash
# optimism 루트 디렉토리에서 실행
cd op-challenger/scripts/p2p
./install-and-run.sh
```

#### 2. 빠른 테스트
```bash
cd op-challenger/scripts/p2p
./quick-test.sh
```

#### 3. 일상 관리
```bash
cd op-challenger/scripts/p2p
./manage.sh [명령어]

# 예시
./manage.sh status      # 서비스 상태 확인
./manage.sh monitor     # 실시간 로그 모니터링
./manage.sh metrics     # 메트릭 확인
./manage.sh p2p         # P2P 네트워크 상태 확인
```

### 이동 이유
- **실제 사용**: op-challenger 바이너리와 함께 관리되어 실제 사용하기 편함
- **논리적 구조**: P2P 챌린저 관련 스크립트가 op-challenger 디렉토리에 있는 것이 논리적
- **유지보수**: 바이너리와 스크립트가 같은 위치에 있어 유지보수가 용이

### 상세 사용법
자세한 사용법은 `op-challenger/scripts/p2p/README.md` 파일을 참조하세요.

---

**이제 `op-challenger/scripts/p2p/` 디렉토리에서 스크립트를 사용하시면 됩니다!** 🎉
