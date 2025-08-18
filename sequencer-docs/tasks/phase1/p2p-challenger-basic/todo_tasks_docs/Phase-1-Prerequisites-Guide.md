# 🔧 P2P 챌린저 사전 작업 가이드

## 📋 개요

이 문서는 P2P 챌린저를 실행하기 전에 필요한 **바이너리 빌드**와 **필수 파일 준비** 방법을 설명합니다.

---

## 🛠️ 1. 바이너리 빌드

### 1.1 저장소 클론 및 브랜치 체크아웃

```bash
# Optimism 저장소 클론
git clone https://github.com/Zena-park/optimism.git
cd optimism

# P2P 챌린저 브랜치로 체크아웃
git checkout feature/p2p-challenger-attention-test
```

### 1.2 바이너리 확인

```bash
# 기존 바이너리 확인
ls -la cannon/bin/cannon
ls -la op-program/bin/op-program
ls -la op-challenger/bin/op-challenger
```

### 1.3 바이너리 빌드 (필요한 경우)

```bash
# Cannon 빌드
cd cannon
make cannon
cd ..

# op-program 빌드
cd op-program
make op-program
cd ..

# op-challenger 빌드
cd op-challenger
make op-challenger
cd ..
```

---

## 📁 2. 필수 파일 준비

### 2.1 prestate 파일 생성

#### 📋 prestate 파일이란?
- **Cannon fault proof 시스템의 절대 초기 상태**를 담고 있는 파일
- 챌린저가 L2 상태의 유효성을 검증할 때 사용하는 **기준점**
- **19MB 크기**의 압축된 바이너리 파일

#### 🔧 생성 방법

**방법 1: op-program 빌드 (권장)**
```bash
cd op-program
make op-program
cd ..
ls -la op-program/bin/prestate*.bin.gz
```
**장점**: 최신 버전, 신뢰성 높음
**단점**: 빌드 시간 소요 (5-10분)

**방법 2: 다운로드**
```bash
# 최신 prestate 파일 다운로드
wget https://raw.githubusercontent.com/ethereum-optimism/optimism/develop/op-program/bin/prestate-mt64Next.bin.gz -O op-program/bin/prestate-mt64Next.bin.gz

# 또는 대체 파일
wget https://raw.githubusercontent.com/ethereum-optimism/optimism/develop/op-program/bin/prestate-interopNext.bin.gz -O op-program/bin/prestate-interopNext.bin.gz
```
**장점**: 빠름, 간단함
**단점**: 네트워크 의존성, 버전 불일치 가능성

**방법 3: Docker로 생성 (고급)**
```bash
cd op-program
make reproducible-prestate
cd ..
```
**장점**: 재현 가능한 빌드
**단점**: Docker 필요, 시간 소요

### 2.2 L2 설정 파일 준비

#### 설정 파일 확인
```bash
# 모니터링하려는 L2 체인의 설정 파일 확인
ls -la packages/contracts-bedrock/deploy-config/sepolia.json
```

#### 설정 파일 다운로드 (필요한 경우)
```bash
# Optimism Sepolia
wget https://raw.githubusercontent.com/ethereum-optimism/optimism/develop/packages/contracts-bedrock/deploy-config/sepolia.json -O packages/contracts-bedrock/deploy-config/sepolia.json

# Base Sepolia
wget https://raw.githubusercontent.com/base-org/optimism/develop/packages/contracts-bedrock/deploy-config/sepolia.json -O packages/contracts-bedrock/deploy-config/base-sepolia.json

# Arbitrum Sepolia
wget https://raw.githubusercontent.com/OffchainLabs/nitro/develop/contracts/deploy-config/sepolia.json -O packages/contracts-bedrock/deploy-config/arbitrum-sepolia.json
```

---

## ✅ 3. 완료 확인

### 3.1 필수 파일 체크리스트

```bash
echo "=== 필수 파일 확인 ==="
echo "1. Cannon 바이너리:"
ls -la cannon/bin/cannon
echo ""
echo "2. op-program 바이너리:"
ls -la op-program/bin/op-program
echo ""
echo "3. prestate 파일:"
ls -la op-program/bin/prestate*.bin.gz
echo ""
echo "4. op-challenger 바이너리:"
ls -la op-challenger/bin/op-challenger
echo ""
echo "5. L2 설정 파일:"
ls -la packages/contracts-bedrock/deploy-config/sepolia.json
echo "======================"
```

### 3.2 자동 확인 스크립트

```bash
#!/bin/bash
echo "🔍 P2P 챌린저 사전 작업 확인 중..."

# 바이너리 확인
if [ -f "cannon/bin/cannon" ]; then
    echo "✅ Cannon 바이너리: OK"
else
    echo "❌ Cannon 바이너리: MISSING"
fi

if [ -f "op-program/bin/op-program" ]; then
    echo "✅ op-program 바이너리: OK"
else
    echo "❌ op-program 바이너리: MISSING"
fi

if [ -f "op-challenger/bin/op-challenger" ]; then
    echo "✅ op-challenger 바이너리: OK"
else
    echo "❌ op-challenger 바이너리: MISSING"
fi

# prestate 파일 확인
if [ -f "op-program/bin/prestate-mt64Next.bin.gz" ] || [ -f "op-program/bin/prestate-interopNext.bin.gz" ]; then
    echo "✅ prestate 파일: OK"
else
    echo "❌ prestate 파일: MISSING"
fi

# L2 설정 파일 확인
if [ -f "packages/contracts-bedrock/deploy-config/sepolia.json" ]; then
    echo "✅ L2 설정 파일: OK"
else
    echo "❌ L2 설정 파일: MISSING"
fi

echo ""
echo "🎯 모든 파일이 준비되면 [Phase-1-Simple-Operations-Guide.md](./Phase-1-Simple-Operations-Guide.md)로 이동하세요!"
```

---

## ⚠️ 주의사항

- prestate 파일은 **Optimism 버전과 일치**해야 함
- 파일이 손상되면 다운로드나 재빌드 필요
- **19MB** 정도의 용량 필요
- 빌드 시 충분한 메모리와 디스크 공간 확보

---

## 🔗 다음 단계

모든 사전 작업이 완료되면 다음 문서로 이동하세요:

**[🚀 P2P 챌린저 간단 운영 가이드](./Phase-1-Simple-Operations-Guide.md)**
