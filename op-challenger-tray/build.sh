#!/bin/bash

# Optimism Challenger Tray Application 빌드 스크립트

set -e

echo "🔨 Optimism Challenger Tray 애플리케이션 빌드 중..."

# 의존성 다운로드
echo "📦 의존성 설치 중..."
go mod tidy

# 빌드
echo "⚙️ 빌드 중..."
go build -o bin/op-challenger-tray .

# 실행 권한 부여
chmod +x bin/op-challenger-tray

echo "✅ 빌드 완료!"
echo ""
echo "🚀 실행 방법:"
echo "   ./bin/op-challenger-tray"
echo ""
echo "📁 실행 파일 위치: $(pwd)/bin/op-challenger-tray"