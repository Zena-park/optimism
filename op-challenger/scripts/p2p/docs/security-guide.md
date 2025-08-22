# Phase 1 P2P 챌린저 네트워크 - 보안 가이드

## 📋 목차

1. [보안 설정](#-보안-설정)
2. [방화벽 구성](#-방화벽-구성)
3. [백업 및 복구](#-백업-및-복구)
4. [키 파일 보안](#-키-파일-보안)

---

## 🔐 보안 설정

### 프라이빗 키 보안
```bash
# 키 파일 권한 제한
chmod 600 /opt/optimism/challenger/keys/*.pem
chown optimism:optimism /opt/optimism/challenger/keys/*.pem

# 키 백업 (암호화)
gpg --symmetric --cipher-algo AES256 p2p-mainnet.pem

# 키 파일 위치 확인
ls -la /opt/optimism/challenger/keys/
```

### 네트워크 보안
```bash
# Rate limiting 활성화
--p2p-rate-limit 1000

# 연결 제한
--p2p-connection-limit 100

# 알려진 악의적 피어 차단 (설정 파일 사용)
```

### 서비스 보안
```bash
# systemd 서비스 보안 설정
[Service]
# 보안 설정
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/optimism/challenger /var/log/optimism/challenger
```

---

## 🛡️ 방화벽 구성

### UFW 방화벽 설정
```bash
# P2P 포트만 열기
sudo ufw allow 9876/tcp

# 메트릭 포트는 로컬만
sudo ufw allow from 127.0.0.1 to any port 7300

# SSH 포트 (필요한 경우)
sudo ufw allow ssh

# 방화벽 활성화
sudo ufw enable

# 방화벽 상태 확인
sudo ufw status
```

### iptables 설정
```bash
# 기본 정책 설정
sudo iptables -P INPUT DROP
sudo iptables -P FORWARD DROP
sudo iptables -P OUTPUT ACCEPT

# 로컬 루프백 허용
sudo iptables -A INPUT -i lo -j ACCEPT

# P2P 포트 허용
sudo iptables -A INPUT -p tcp --dport 9876 -j ACCEPT

# 메트릭 포트 (로컬만)
sudo iptables -A INPUT -p tcp --dport 7300 -s 127.0.0.1 -j ACCEPT

# ESTABLISHED 연결 허용
sudo iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT

# 변경사항 저장
sudo iptables-save > /etc/iptables/rules.v4
```

---

## 💾 백업 및 복구

### 데이터 백업
```bash
# 일일 백업 스크립트
#!/bin/bash
BACKUP_DIR="/backup/optimism/challenger"
DATA_DIR="/opt/optimism/challenger/data"

# 챌린저 정지
systemctl stop op-challenger

# 데이터 백업
rsync -av $DATA_DIR/ $BACKUP_DIR/$(date +%Y%m%d)/

# 챌린저 재시작
systemctl start op-challenger
```

### 복구 절차
```bash
# 서비스 정지
systemctl stop op-challenger

# 데이터 복구
rsync -av /backup/optimism/challenger/20240101/ /opt/optimism/challenger/data/

# 권한 복구
chown -R optimism:optimism /opt/optimism/challenger/

# 서비스 재시작
systemctl start op-challenger
```

### 자동 백업 설정
```bash
# crontab에 백업 작업 추가
0 2 * * * /opt/optimism/challenger/scripts/backup.sh

# 백업 스크립트 생성
cat > /opt/optimism/challenger/scripts/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/backup/optimism/challenger"
DATA_DIR="/opt/optimism/challenger/data"
DATE=$(date +%Y%m%d_%H%M%S)

# 백업 디렉토리 생성
mkdir -p $BACKUP_DIR/$DATE

# 데이터 백업
rsync -av $DATA_DIR/ $BACKUP_DIR/$DATE/

# 7일 이상 된 백업 삭제
find $BACKUP_DIR -type d -mtime +7 -exec rm -rf {} \;
EOF

chmod +x /opt/optimism/challenger/scripts/backup.sh
```

---

## 🔑 키 파일 보안

### 키 파일 생성
```bash
# 안전한 키 생성
openssl genpkey -algorithm ED25519 -out /opt/optimism/challenger/keys/p2p-key.pem

# 권한 설정
chmod 600 /opt/optimism/challenger/keys/p2p-key.pem
chown optimism:optimism /opt/optimism/challenger/keys/p2p-key.pem
```

### 키 파일 백업
```bash
# 암호화된 백업 생성
gpg --symmetric --cipher-algo AES256 /opt/optimism/challenger/keys/p2p-key.pem

# 백업 파일을 안전한 위치로 이동
mv /opt/optimism/challenger/keys/p2p-key.pem.gpg /backup/secure/
```

### 키 파일 복구
```bash
# 암호화된 백업 복구
gpg --decrypt /backup/secure/p2p-key.pem.gpg > /opt/optimism/challenger/keys/p2p-key.pem

# 권한 복구
chmod 600 /opt/optimism/challenger/keys/p2p-key.pem
chown optimism:optimism /opt/optimism/challenger/keys/p2p-key.pem
```

---

## 🔗 관련 문서

- [설치 가이드](installation-guide.md)
- [Devnet 사용법](devnet-guide.md)
- [모니터링 가이드](monitoring-guide.md)
- [트러블슈팅 가이드](troubleshooting-guide.md)

---

**보안을 통해 안전한 P2P 챌린저 네트워크를 운영하세요!** 🔒
