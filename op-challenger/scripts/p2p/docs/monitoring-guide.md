# Phase 1 P2P 챌린저 네트워크 - 모니터링 가이드

## 📋 목차

1. [📊 모니터링 개요](#-모니터링-개요)
2. [🚀 자동화된 모니터링](#-자동화된-모니터링)
3. [📈 실시간 모니터링](#-실시간-모니터링)
4. [🔍 로그 모니터링](#-로그-모니터링)
5. [⚡ 성능 모니터링](#-성능-모니터링)
6. [🛠️ 고급 모니터링](#-고급-모니터링)
7. [📱 알림 설정](#-알림-설정)

---

## 📊 모니터링 개요

### 모니터링 목표
- **가용성**: Devnet과 P2P 챌린저 네트워크의 지속적인 가동
- **성능**: 네트워크 성능 및 응답 시간 최적화
- **안정성**: 오류 및 장애 조기 발견 및 대응
- **리소스**: 시스템 리소스 사용량 최적화

### 모니터링 대상
- **Devnet 서비스**: op-node, op-batcher, op-proposer, op-challenger, geth
- **P2P 네트워크**: 피어 연결, 메시지 전송, 네트워크 토폴로지
- **시스템 리소스**: CPU, 메모리, 디스크, 네트워크
- **애플리케이션 성능**: 응답 시간, 처리량, 오류율

---

## 🚀 자동화된 모니터링

### 자동 모니터링 스크립트
```bash
# Devnet 상태 자동 모니터링
cd op-challenger/scripts/p2p
./monitor-devnet.sh
```

**자동 모니터링 기능:**
- ✅ Devnet 서비스 상태 확인
- ✅ 포트 연결 상태 확인
- ✅ 시스템 리소스 사용량 확인
- ✅ P2P 네트워크 상태 확인
- ✅ 자동 알림 및 경고

### 시스템 상태 확인
```bash
# 시스템 요구사항 확인
./check-system.sh

# 포트 연결 상태 확인
./check-ports.sh
```

---

## 📈 실시간 모니터링

### Devnet 상태 모니터링
```bash
# Devnet 실행 상태 확인
kurtosis enclave inspect simple-devnet

# Devnet 서비스 목록 확인
kurtosis enclave inspect simple-devnet --format json | jq '.services'

# 특정 서비스 상태 확인
kurtosis enclave inspect simple-devnet --service op-node
kurtosis enclave inspect simple-devnet --service op-challenger
```

### 실시간 로그 모니터링
```bash
# 전체 Devnet 로그 실시간 모니터링
kurtosis enclave logs simple-devnet --follow

# 특정 서비스 로그 실시간 모니터링
kurtosis enclave logs simple-devnet --service op-node --follow
kurtosis enclave logs simple-devnet --service op-challenger --follow

# 오류 로그만 실시간 모니터링
kurtosis enclave logs simple-devnet --follow | grep -i error
```

### 시스템 리소스 실시간 모니터링
```bash
# Docker 컨테이너 리소스 사용량
docker stats

# 시스템 리소스 사용량 (Linux)
htop
iotop
iftop

# 시스템 리소스 사용량 (macOS)
top
iostat 1
netstat -i 1
```

---

## 🔍 로그 모니터링

### 로그 수집 및 분석
```bash
# 전체 로그 수집
kurtosis enclave logs simple-devnet > devnet.log 2>&1

# 특정 서비스 로그 수집
kurtosis enclave logs simple-devnet --service op-node > op-node.log 2>&1
kurtosis enclave logs simple-devnet --service op-challenger > op-challenger.log 2>&1

# 시간별 로그 필터링
kurtosis enclave logs simple-devnet --since "2024-01-01T00:00:00Z"
kurtosis enclave logs simple-devnet --until "2024-01-01T23:59:59Z"
```

### 로그 분석 도구
```bash
# 오류 로그 분석
grep -i error devnet.log | wc -l
grep -i error devnet.log | tail -20

# 경고 로그 분석
grep -i warn devnet.log | wc -l
grep -i warn devnet.log | tail -20

# 특정 패턴 검색
grep -i "connection refused" devnet.log
grep -i "timeout" devnet.log
grep -i "out of memory" devnet.log

# 로그 통계
wc -l devnet.log
grep -o "2024-[0-9][0-9]-[0-9][0-9]" devnet.log | sort | uniq -c
```

### P2P 네트워크 로그 분석
```bash
# P2P 관련 로그 필터링
grep -i "p2p" op-challenger.log
grep -i "peer" op-challenger.log
grep -i "connection" op-challenger.log

# 피어 연결 상태 확인
grep -i "connected to peer" op-challenger.log
grep -i "disconnected from peer" op-challenger.log

# 메시지 전송 로그
grep -i "sent message" op-challenger.log
grep -i "received message" op-challenger.log
```

---

## ⚡ 성능 모니터링

### 네트워크 성능 모니터링
```bash
# 네트워크 대역폭 사용량
iftop -i eth0

# 네트워크 연결 상태
netstat -i
ss -tuln

# 포트별 연결 수
netstat -an | grep :8545 | wc -l
netstat -an | grep :9545 | wc -l
netstat -an | grep :9876 | wc -l
```

### 애플리케이션 성능 모니터링
```bash
# RPC 응답 시간 테스트
time curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545

# 블록 생성 속도 확인
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq '.result'

# 트랜잭션 풀 상태 확인
curl -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}' \
  http://localhost:9545
```

### 메모리 및 CPU 성능 모니터링
```bash
# 프로세스별 리소스 사용량
ps aux | grep -E "(op-node|op-challenger|geth)" | sort -k3 -nr

# 메모리 사용량 상세 분석
cat /proc/meminfo | grep -E "(MemTotal|MemFree|MemAvailable)"

# CPU 사용률 모니터링
top -p $(pgrep -d',' op-node op-challenger geth)
```

---

## 🛠️ 고급 모니터링

### 커스텀 모니터링 스크립트
```bash
#!/bin/bash
# custom-monitor.sh

# Devnet 상태 확인
echo "=== Devnet Status ==="
kurtosis enclave inspect simple-devnet --format json | jq '.status'

# 서비스별 상태 확인
echo "=== Service Status ==="
services=("op-node" "op-batcher" "op-proposer" "op-challenger" "geth")
for service in "${services[@]}"; do
    status=$(kurtosis enclave inspect simple-devnet --service "$service" --format json | jq -r '.status')
    echo "$service: $status"
done

# 시스템 리소스 확인
echo "=== System Resources ==="
echo "CPU Usage: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
echo "Memory Usage: $(free | grep Mem | awk '{printf "%.2f%%", $3/$2 * 100.0}')"
echo "Disk Usage: $(df / | tail -1 | awk '{print $5}')"

# 네트워크 연결 확인
echo "=== Network Connections ==="
echo "L1 RPC: $(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545 | jq -r '.result' 2>/dev/null || echo "Not available")"
echo "L2 RPC: $(curl -s -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:9545 | jq -r '.result' 2>/dev/null || echo "Not available")"
```

### 지속적인 모니터링
```bash
# 주기적 모니터링 (5분마다)
watch -n 300 ./custom-monitor.sh

# 백그라운드 모니터링
nohup ./custom-monitor.sh > monitoring.log 2>&1 &

# 로그 로테이션
logrotate -f /etc/logrotate.d/devnet-monitoring
```

### 메트릭 수집
```bash
# 블록 높이 추적
while true; do
    l1_block=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      http://localhost:8545 | jq -r '.result' 2>/dev/null || echo "0")
    l2_block=$(curl -s -X POST -H "Content-Type: application/json" \
      --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
      http://localhost:9545 | jq -r '.result' 2>/dev/null || echo "0")
    echo "$(date): L1=$l1_block, L2=$l2_block" >> block-heights.log
    sleep 60
done
```

---

## 📱 알림 설정

### 기본 알림 설정
```bash
#!/bin/bash
# alert-monitor.sh

# 임계값 설정
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
DISK_THRESHOLD=90

# CPU 사용률 확인
cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
if (( $(echo "$cpu_usage > $CPU_THRESHOLD" | bc -l) )); then
    echo "ALERT: High CPU usage: ${cpu_usage}%" | mail -s "Devnet Alert" admin@example.com
fi

# 메모리 사용률 확인
memory_usage=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
if (( $(echo "$memory_usage > $MEMORY_THRESHOLD" | bc -l) )); then
    echo "ALERT: High memory usage: ${memory_usage}%" | mail -s "Devnet Alert" admin@example.com
fi

# 디스크 사용률 확인
disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$disk_usage" -gt "$DISK_THRESHOLD" ]; then
    echo "ALERT: High disk usage: ${disk_usage}%" | mail -s "Devnet Alert" admin@example.com
fi
```

### 서비스 상태 알림
```bash
#!/bin/bash
# service-alert.sh

# Devnet 서비스 상태 확인
services=("op-node" "op-batcher" "op-proposer" "op-challenger" "geth")
for service in "${services[@]}"; do
    status=$(kurtosis enclave inspect simple-devnet --service "$service" --format json | jq -r '.status' 2>/dev/null)
    if [ "$status" != "RUNNING" ]; then
        echo "ALERT: Service $service is not running (Status: $status)" | mail -s "Devnet Service Alert" admin@example.com
    fi
done

# 네트워크 연결 확인
l1_response=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 2>/dev/null)
if [ -z "$l1_response" ]; then
    echo "ALERT: L1 RPC is not responding" | mail -s "Devnet Network Alert" admin@example.com
fi

l2_response=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:9545 2>/dev/null)
if [ -z "$l2_response" ]; then
    echo "ALERT: L2 RPC is not responding" | mail -s "Devnet Network Alert" admin@example.com
fi
```

### Slack/Discord 알림 설정
```bash
#!/bin/bash
# slack-alert.sh

SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

send_slack_alert() {
    local message="$1"
    curl -X POST -H 'Content-type: application/json' \
      --data "{\"text\":\"$message\"}" \
      "$SLACK_WEBHOOK_URL"
}

# 서비스 상태 확인 및 알림
services=("op-node" "op-batcher" "op-proposer" "op-challenger" "geth")
for service in "${services[@]}"; do
    status=$(kurtosis enclave inspect simple-devnet --service "$service" --format json | jq -r '.status' 2>/dev/null)
    if [ "$status" != "RUNNING" ]; then
        send_slack_alert "🚨 Devnet Alert: Service $service is not running (Status: $status)"
    fi
done
```

---

## 📊 모니터링 대시보드

### 간단한 웹 대시보드
```bash
#!/bin/bash
# dashboard.sh

cat << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Devnet Monitoring Dashboard</title>
    <meta http-equiv="refresh" content="30">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .running { background-color: #d4edda; color: #155724; }
        .stopped { background-color: #f8d7da; color: #721c24; }
        .warning { background-color: #fff3cd; color: #856404; }
    </style>
</head>
<body>
    <h1>Devnet Monitoring Dashboard</h1>
    <div id="status">
EOF

# 서비스 상태 출력
services=("op-node" "op-batcher" "op-proposer" "op-challenger" "geth")
for service in "${services[@]}"; do
    status=$(kurtosis enclave inspect simple-devnet --service "$service" --format json | jq -r '.status' 2>/dev/null || echo "UNKNOWN")
    if [ "$status" = "RUNNING" ]; then
        echo "        <div class='status running'>✅ $service: $status</div>"
    else
        echo "        <div class='status stopped'>❌ $service: $status</div>"
    fi
done

cat << 'EOF'
    </div>
    <div id="resources">
        <h2>System Resources</h2>
        <p>CPU Usage: <span id="cpu">Loading...</span></p>
        <p>Memory Usage: <span id="memory">Loading...</span></p>
        <p>Disk Usage: <span id="disk">Loading...</span></p>
    </div>
    <script>
        // 30초마다 페이지 새로고침
        setTimeout(function() {
            location.reload();
        }, 30000);
    </script>
</body>
</html>
EOF
```

---

## 📚 추가 리소스

- [Devnet 가이드](devnet-guide.md): 상세한 Devnet 설정 및 운영 가이드
- [설치 가이드](installation-guide.md): 시스템 요구사항 및 설치 방법
- [문제 해결 가이드](troubleshooting-guide.md): 일반적인 문제 해결 방법
- [보안 가이드](security-guide.md): 보안 고려사항 및 모범 사례

---

## 🎯 모니터링 체크리스트

### 기본 모니터링
- [ ] Devnet 서비스 상태 확인
- [ ] 시스템 리소스 사용량 확인
- [ ] 네트워크 연결 상태 확인
- [ ] 로그 오류 및 경고 확인

### 성능 모니터링
- [ ] 응답 시간 측정
- [ ] 처리량 모니터링
- [ ] 리소스 사용률 추적
- [ ] 성능 병목 지점 식별

### 알림 설정
- [ ] 임계값 기반 알림 설정
- [ ] 서비스 상태 알림 설정
- [ ] 네트워크 연결 알림 설정
- [ ] 알림 채널 구성 (이메일, Slack, Discord)

### 대시보드 구성
- [ ] 실시간 상태 대시보드 구성
- [ ] 성능 메트릭 시각화
- [ ] 알림 히스토리 추적
- [ ] 모니터링 데이터 백업
