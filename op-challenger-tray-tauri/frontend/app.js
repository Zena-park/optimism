// 6탭 사이드바 네비게이션 - 시스템 애플리케이션 스타일
// Tauri API 초기화
let invoke;

// DOM 로드 완료 시 앱 초기화
document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM 로드 완료 - 앱 초기화 시작');
    
    try {
        // Tauri API 가져오기 (선택적)
        if (window.__TAURI__ && window.__TAURI__.tauri) {
            const { invoke: tauriInvoke } = window.__TAURI__.tauri;
            invoke = tauriInvoke;
            console.log('Tauri API 초기화 성공');
        } else {
            console.log('Tauri API 없음 - 브라우저 모드로 실행');
            invoke = async (cmd, args) => {
                console.log(`Mock invoke: ${cmd}`, args);
                // Mock 데이터 반환
                if (cmd === 'get_config') {
                    return {
                        network: 'sepolia',
                        monitor_only: true,
                        p2p_enabled: true
                    };
                }
                if (cmd === 'get_status') {
                    return { running: false };
                }
                return { success: true, message: 'Mock success' };
            };
        }

        // 앱 초기화
        await initializeApp();
    } catch (error) {
        console.error('앱 초기화 실패:', error);
        showNotification('앱 초기화 실패: ' + error.message, 'error');
    }
});

// 전역 변수
let statusInterval;
let logsInterval;
let currentConfig = {};
let currentTab = 'dashboard';

// 네트워크별 기본값
const networkDefaults = {
    sepolia: {
        l1_eth_rpc: "https://ethereum-sepolia-rpc.publicnode.com",
        l2_eth_rpc: "https://sepolia.optimism.io",
        rollup_rpc: "https://sepolia.optimism.io"
    },
    mainnet: {
        l1_eth_rpc: "https://ethereum-rpc.publicnode.com",
        l2_eth_rpc: "https://mainnet.optimism.io",
        rollup_rpc: "https://mainnet.optimism.io"
    },
    local: {
        l1_eth_rpc: "http://localhost:8545",
        l2_eth_rpc: "http://localhost:8546",
        rollup_rpc: "http://localhost:8546"
    }
};

// 앱 초기화
async function initializeApp() {
    try {
        console.log('앱 초기화 시작...');
        
        // 이벤트 리스너 설정
        setupEventListeners();

        // 설정 로드
        await loadSettings();

        // 기본 탭 로드 (Dashboard)
        console.log('기본 탭 로드 시도...');
        switchTab('dashboard');

        // 상태 업데이트
        await updateSystemStatus();
        await updateDashboard();

        // 주기적 상태 업데이트 시작
        startStatusPolling();

        console.log('⚡ Optimism Challenger 초기화 완료');
        showNotification('애플리케이션이 초기화되었습니다', 'success');

        // 디버깅: 모든 탭 패널 확인
        console.log('=== 탭 패널 디버깅 ===');
        document.querySelectorAll('.tab-panel').forEach(panel => {
            console.log(`탭 패널: ${panel.id}, 표시: ${getComputedStyle(panel).display}, 클래스: ${panel.className}`);
        });

    } catch (error) {
        console.error('앱 초기화 실패:', error);
        showNotification('초기화 중 오류가 발생했습니다: ' + error, 'error');
    }
}

// 이벤트 리스너 설정
function setupEventListeners() {
    console.log('이벤트 리스너 설정 시작');
    
    // 사이드바 네비게이션
    const navItems = document.querySelectorAll('.nav-item');
    console.log(`네비게이션 아이템 개수: ${navItems.length}`);
    
    navItems.forEach((item, index) => {
        const tabName = item.dataset.tab;
        console.log(`네비게이션 아이템 ${index}: ${tabName}`);
        
        // 기존 이벤트 리스너 제거 (중복 방지)
        item.removeEventListener('click', handleTabClick);
        
        // 새 이벤트 리스너 추가
        item.addEventListener('click', handleTabClick);
        
        // 시각적 피드백을 위한 hover 효과도 추가
        item.addEventListener('mouseenter', () => {
            console.log(`Mouse enter: ${tabName}`);
        });
    });

    // 윈도우 종료 시 정리
    window.addEventListener('beforeunload', cleanup);
    
    console.log('이벤트 리스너 설정 완료');
}

// 탭 클릭 핸들러 함수
function handleTabClick(event) {
    const item = event.currentTarget;
    const tabName = item.dataset.tab;
    console.log(`탭 클릭: ${tabName}`);
    
    // 클릭 효과
    item.style.transform = 'scale(0.95)';
    setTimeout(() => {
        item.style.transform = '';
    }, 150);
    
    switchTab(tabName);
}

// 탭별 이벤트 리스너 설정
function setupTabEventListeners(tabName) {
    switch(tabName) {
        case 'dashboard':
            // 챌린저 컨트롤
            document.getElementById('startBtn')?.addEventListener('click', startChallenger);
            document.getElementById('stopBtn')?.addEventListener('click', stopChallenger);
            break;

        case 'settings':
            // 설정 관련
            document.getElementById('network')?.addEventListener('change', updateNetworkDefaults);

            // 운영 모드 라디오 버튼
            document.querySelectorAll('input[name="operation_mode"]').forEach(radio => {
                radio.addEventListener('change', updateOperationMode);
            });
            break;

        case 'logs':
            // 로그 필터
            document.getElementById('logLevel')?.addEventListener('change', filterLogs);
            break;

        case 'history':
            // 히스토리 필터
            document.getElementById('historyFilter')?.addEventListener('change', filterHistory);
            break;
    }
}

// 탭 전환 - CSS 방식
function switchTab(tabName) {
    console.log(`탭 전환: ${currentTab} → ${tabName}`);
    
    if (currentTab === tabName) {
        return;
    }

    // 모든 탭 패널 숨기기
    document.querySelectorAll('.tab-panel').forEach(panel => {
        panel.classList.remove('active');
    });

    // 선택된 탭 패널 표시
    const targetPanel = document.getElementById(tabName);
    if (targetPanel) {
        targetPanel.classList.add('active');
        console.log(`✓ ${tabName} 패널 활성화`);
    } else {
        console.error(`✗ 탭 패널 없음: ${tabName}`);
        showNotification(`탭을 찾을 수 없습니다: ${tabName}`, 'error');
        return;
    }

    // 모든 네비게이션 아이템 비활성화
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });

    // 선택된 탭 활성화
    const targetNavItem = document.querySelector(`[data-tab="${tabName}"]`);
    if (targetNavItem) {
        targetNavItem.classList.add('active');
    }

    currentTab = tabName;

    // 탭별 초기화 작업
    switch (tabName) {
        case 'dashboard':
            updateDashboard();
            break;
        case 'l2-monitor':
            refreshL2Data();
            break;
        case 'p2p-challengers':
            refreshP2PData();
            break;
        case 'history':
            loadHistory();
            break;
        case 'settings':
            break;
        case 'logs':
            startLogsPolling();
            break;
    }

    // 다른 탭으로 이동 시 로그 폴링 중지
    if (tabName !== 'logs') {
        stopLogsPolling();
    }
}

// 나머지 함수들... (기존 app.js의 나머지 부분)
// 설정 관련 함수들
function getFormConfig() {
    const bootnodes = document.getElementById('p2p_bootnodes')?.value
        .split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0) || [];

    const operationMode = document.querySelector('input[name="operation_mode"]:checked')?.value || 'monitor';

    return {
        network: document.getElementById('network')?.value || 'sepolia',
        l1_eth_rpc: document.getElementById('l1_eth_rpc')?.value || '',
        l2_eth_rpc: document.getElementById('l2_eth_rpc')?.value || '',
        rollup_rpc: document.getElementById('rollup_rpc')?.value || '',
        p2p_enabled: document.getElementById('p2p_enabled')?.checked !== false,
        p2p_listen_addr: document.getElementById('p2p_listen_addr')?.value || '/ip4/0.0.0.0/tcp/9876',
        p2p_max_peers: parseInt(document.getElementById('p2p_max_peers')?.value || '50'),
        p2p_bootnodes: bootnodes,
        monitor_only: operationMode === 'monitor',
        notifications_enabled: document.getElementById('notifications_enabled')?.checked === true,
        email_address: document.getElementById('email_address')?.value || null,
        slack_webhook: document.getElementById('slack_webhook')?.value || null,
        log_level: document.getElementById('logLevel')?.value || 'info'
    };
}

// 설정 저장/로드 등 기본 함수들
async function saveSettings() {
    try {
        const config = getFormConfig();
        if (invoke) {
            await invoke('save_config', { config });
            currentConfig = config;
            showNotification('설정이 저장되었습니다', 'success');
        } else {
            console.log('Mock save settings:', config);
            showNotification('설정 저장됨 (Mock)', 'info');
        }
    } catch (error) {
        console.error('설정 저장 실패:', error);
        showNotification('설정 저장 실패: ' + error, 'error');
    }
}

async function loadSettings() {
    try {
        if (invoke) {
            const config = await invoke('get_config');
            currentConfig = config;
            console.log('설정 로드 완료:', config);
        } else {
            console.log('Mock load settings - 기본값 사용');
            currentConfig = {
                network: 'sepolia',
                monitor_only: true,
                p2p_enabled: true
            };
        }
    } catch (error) {
        console.error('설정 로드 실패:', error);
        // 오류가 발생해도 앱은 계속 실행되도록 함
        currentConfig = {
            network: 'sepolia', 
            monitor_only: true,
            p2p_enabled: true
        };
        console.log('기본 설정으로 계속 진행');
    }
}

// 챌린저 제어
async function startChallenger() {
    try {
        const result = await invoke('start_challenger', { config: currentConfig });
        console.log('챌린저 시작 결과:', result);
        showNotification('챌린저가 시작되었습니다', 'success');
        await updateSystemStatus();
    } catch (error) {
        console.error('챌린저 시작 실패:', error);
        showNotification('챌린저 시작 실패: ' + error, 'error');
    }
}

async function stopChallenger() {
    try {
        const result = await invoke('stop_challenger');
        console.log('챌린저 중지 결과:', result);
        showNotification('챌린저가 중지되었습니다', 'success');
        await updateSystemStatus();
    } catch (error) {
        console.error('챌린저 중지 실패:', error);
        showNotification('챌린저 중지 실패: ' + error, 'error');
    }
}

// 상태 업데이트 함수들
async function updateSystemStatus() {
    try {
        const status = await invoke('get_status');
        console.log('시스템 상태:', status);
        
        // DOM 요소 업데이트
        const systemStatusElement = document.getElementById('systemStatus');
        if (systemStatusElement) {
            if (status && status.running) {
                systemStatusElement.textContent = 'Running';
                systemStatusElement.className = 'status-badge running';
            } else {
                systemStatusElement.textContent = 'Stopped';
                systemStatusElement.className = 'status-badge stopped';
            }
        }
    } catch (error) {
        console.error('상태 업데이트 실패:', error);
    }
}

function updateDashboard() {
    // 대시보드 업데이트 로직
}

function refreshL2Data() {
    showNotification('L2 데이터 새로고침', 'info');
}

function refreshP2PData() {
    showNotification('P2P 데이터 새로고침', 'info');
}

function loadHistory() {
    // 히스토리 로드 로직
}

function updateLogs() {
    // 로그 업데이트 로직
}

// 유틸리티 함수들
function filterLogs() {}
function filterHistory() {}
function updateNetworkDefaults() {}
function updateOperationMode() {}
function resetSettings() {}
function clearLogs() {}
function exportLogs() {}
function exportHistory() {}

function startStatusPolling() {
    if (statusInterval) return;
    statusInterval = setInterval(updateSystemStatus, 3000);
}

function startLogsPolling() {
    if (logsInterval) return;
    logsInterval = setInterval(updateLogs, 1000);
}

function stopLogsPolling() {
    if (logsInterval) {
        clearInterval(logsInterval);
        logsInterval = null;
    }
}

// 알림 표시
function showNotification(message, type = 'info') {
    console.log(`[${type.toUpperCase()}] ${message}`);
    
    // 간단한 알림 시스템
    const notification = document.createElement('div');
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 20px;
        border-radius: 6px;
        color: white;
        font-size: 13px;
        font-weight: 500;
        z-index: 10000;
        max-width: 300px;
        opacity: 0.9;
    `;

    const colors = {
        success: '#34c759',
        error: '#ff3b30',
        warning: '#ff9500',
        info: '#007aff'
    };

    notification.style.backgroundColor = colors[type] || colors.info;
    notification.textContent = message;

    document.body.appendChild(notification);

    // 3초 후 자동 제거
    setTimeout(() => {
        if (notification.parentNode) {
            notification.remove();
        }
    }, 3000);
}

// 정리
function cleanup() {
    if (statusInterval) {
        clearInterval(statusInterval);
        statusInterval = null;
    }
    if (logsInterval) {
        clearInterval(logsInterval);
        logsInterval = null;
    }
}

// 글로벌 함수들 (HTML에서 직접 호출되는 함수들)
window.startChallenger = startChallenger;
window.stopChallenger = stopChallenger;
window.saveSettings = saveSettings;
window.resetSettings = resetSettings;
window.updateNetworkDefaults = updateNetworkDefaults;
window.refreshL2Data = refreshL2Data;
window.refreshP2PData = refreshP2PData;
window.exportHistory = exportHistory;
window.clearLogs = clearLogs;
window.exportLogs = exportLogs;