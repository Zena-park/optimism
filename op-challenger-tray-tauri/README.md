# 🎉 Optimism Challenger Tray (Tauri)

Optimism 챌린저를 위한 크로스 플랫폼 GUI 트레이 애플리케이션입니다. Tauri를 기반으로 제작되어 가볍고 빠르며 안정적입니다.

## ✨ 주요 기능

- 🚀 **챌린저 프로세스 관리**: 시작/정지 원클릭 제어
- 📊 **실시간 로그 모니터링**: 챌린저 상태 실시간 확인
- ⚙️ **설정 관리**: P2P, 네트워크, 메트릭 설정
- 🌐 **웹 기반 UI**: 직관적이고 반응형 인터페이스
- 💾 **설정 저장/로드**: YAML 기반 설정 관리
- 🔧 **크로스 플랫폼**: Windows, macOS, Linux 지원

## 🚀 빠른 시작

### 필수 요구사항

- [Rust](https://rustup.rs/) (1.70+)
- [Node.js](https://nodejs.org/) (선택사항, 개발용)
- Optimism op-challenger 바이너리

### 설치 및 실행

1. **저장소 클론**
   ```bash
   git clone <repository-url>
   cd op-challenger-tray-tauri
   ```

2. **의존성 설치**
   ```bash
   cargo build
   ```

3. **개발 모드 실행**
   ```bash
   cargo run
   ```

4. **릴리즈 빌드**
   ```bash
   cargo build --release
   ```

## 📁 프로젝트 구조

```
op-challenger-tray-tauri/
├── src/
│   └── main.rs          # Rust 백엔드 (Tauri 명령어들)
├── dist/
│   └── index.html       # 프론트엔드 UI
├── src-tauri/
│   └── icons/           # 애플리케이션 아이콘들
├── Cargo.toml           # Rust 의존성
├── tauri.conf.json      # Tauri 설정
└── README.md
```

## 🔧 설정

### 챌린저 설정

애플리케이션은 다음 설정들을 지원합니다:

#### 기본 설정
- **네트워크**: `sepolia` (테스트넷) 또는 `mainnet`
- **L1 RPC**: Ethereum L1 RPC 엔드포인트
- **L2 RPC**: Optimism L2 RPC 엔드포인트
- **데이터 디렉토리**: 챌린저 데이터 저장 경로

#### P2P 설정
- **P2P 활성화**: 피어-투-피어 네트워킹 사용 여부
- **리슨 주소**: P2P 리슨 주소 (예: `/ip4/0.0.0.0/tcp/9876`)
- **네트워크 ID**: P2P 네트워크 식별자
- **최대 피어 수**: 연결할 최대 피어 수

#### 고급 설정
- **Cannon 바이너리**: cannon 실행 파일 경로
- **OP 프로그램**: op-program 서버 경로
- **로그 레벨**: 로그 상세 수준
- **메트릭**: 메트릭 수집 및 노출 설정

### 설정 파일

설정은 자동으로 `~/.optimism/challenger/config.yaml`에 저장됩니다.

## 🖥️ 사용법

### 1. 애플리케이션 시작
```bash
./target/release/op-challenger-tray-tauri
```

### 2. 챌린저 설정
1. 웹 인터페이스에서 네트워크 및 RPC 설정
2. P2P 설정 구성 (선택사항)
3. "설정 저장" 버튼 클릭

### 3. 챌린저 실행
1. "🚀 챌린저 시작" 버튼 클릭
2. 실시간 로그에서 상태 확인
3. 필요시 "⏹️ 챌린저 정지" 버튼으로 중지

### 4. 모니터링
- **상태 표시**: 실행/정지 상태 실시간 확인
- **로그 모니터링**: 챌린저 출력 실시간 표시
- **자동 업데이트**: 1초마다 상태 및 로그 갱신

## 🔧 개발

### 개발 환경 설정

1. **Tauri CLI 설치**
   ```bash
   cargo install tauri-cli
   ```

2. **개발 서버 실행**
   ```bash
   cargo tauri dev
   ```

3. **빌드**
   ```bash
   cargo tauri build
   ```

### 코드 구조

#### Rust 백엔드 (`src/main.rs`)
- **AppState**: 애플리케이션 전역 상태 관리
- **ChallengerConfig**: 챌린저 설정 구조체
- **Tauri 명령어**: 프론트엔드-백엔드 통신 API

주요 명령어들:
- `get_status()`: 챌린저 실행 상태 조회
- `get_logs()`: 실시간 로그 조회
- `start_challenger()`: 챌린저 프로세스 시작
- `stop_challenger()`: 챌린저 프로세스 정지
- `save_config()`: 설정 저장
- `get_config()`: 설정 로드

#### 프론트엔드 (`dist/index.html`)
- **HTML/CSS**: 반응형 웹 인터페이스
- **JavaScript**: Tauri API를 통한 백엔드 통신
- **실시간 업데이트**: 상태 및 로그 자동 갱신

## 🛠️ 기술 스택

- **백엔드**: Rust + Tauri
- **프론트엔드**: HTML/CSS/JavaScript
- **프로세스 관리**: std::process
- **설정 관리**: YAML (serde)
- **비동기 처리**: Tokio
- **로깅**: chrono

## 📊 성능 특징

- **바이너리 크기**: ~10MB (Electron 대비 90% 절약)
- **메모리 사용량**: ~20-30MB
- **시작 시간**: < 1초
- **CPU 사용량**: 최소한 (유휴 시 0%)

## 🐛 문제 해결

### 일반적인 문제들

1. **챌린저 바이너리를 찾을 수 없음**
   - `./bin/op-challenger` 또는 `../op-challenger/bin/op-challenger` 경로 확인
   - 실행 권한 확인: `chmod +x ./bin/op-challenger`

2. **설정이 저장되지 않음**
   - `~/.optimism/challenger/` 디렉토리 권한 확인
   - 디스크 공간 확인

3. **P2P 연결 실패**
   - 방화벽 설정 확인
   - 네트워크 포트(9876) 사용 가능 여부 확인

4. **로그가 표시되지 않음**
   - 챌린저 프로세스 실행 상태 확인
   - 브라우저 개발자 도구에서 오류 확인

### 로그 파일

- 챌린저 로그: 애플리케이션 내 실시간 표시
- 애플리케이션 로그: 터미널/콘솔 출력

## 🤝 기여하기

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 `LICENSE` 파일을 참조하세요.

## 🔗 관련 링크

- [Optimism 공식 문서](https://docs.optimism.io/)
- [Tauri 문서](https://tauri.app/)
- [op-challenger 가이드](https://docs.optimism.io/builders/node-operators/tutorials/mainnet)

## 📞 지원

문제가 발생하거나 질문이 있으시면:
- GitHub Issues 생성
- [Optimism Discord](https://discord.optimism.io/) 참여

---

**Made with ❤️ using Tauri and Rust**