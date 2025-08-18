package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type WebChallengerTray struct {
	config    *ChallengerConfig
	process   *exec.Cmd
	isRunning bool
	logs      []string
	configFilePath string
}

type ChallengerConfig struct {
	Network        string   `yaml:"network" json:"network"`
	L1EthRPC       string   `yaml:"l1_eth_rpc" json:"l1_eth_rpc"`
	L1Beacon       string   `yaml:"l1_beacon" json:"l1_beacon"`
	L2EthRPC       string   `yaml:"l2_eth_rpc" json:"l2_eth_rpc"`
	RollupRPC      string   `yaml:"rollup_rpc" json:"rollup_rpc"`
	DataDir        string   `yaml:"datadir" json:"datadir"`
	P2PEnabled     bool     `yaml:"p2p_enabled" json:"p2p_enabled"`
	P2PListenAddr  string   `yaml:"p2p_listen_addr" json:"p2p_listen_addr"`
	P2PNetworkID   string   `yaml:"p2p_network_id" json:"p2p_network_id"`
	P2PMaxPeers    int      `yaml:"p2p_max_peers" json:"p2p_max_peers"`
	P2PBootnodes   []string `yaml:"p2p_bootnodes" json:"p2p_bootnodes"`
	CannonBin      string   `yaml:"cannon_bin" json:"cannon_bin"`
	CannonServer   string   `yaml:"cannon_server" json:"cannon_server"`
	LogLevel       string   `yaml:"log_level" json:"log_level"`
	MetricsEnabled bool     `yaml:"metrics_enabled" json:"metrics_enabled"`
	MetricsAddr    string   `yaml:"metrics_addr" json:"metrics_addr"`
	MetricsPort    int      `yaml:"metrics_port" json:"metrics_port"`
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>🎉 Optimism Challenger GUI</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #ff0420, #ff6b6b); color: white; padding: 20px; border-radius: 10px; margin-bottom: 20px; }
        .card { background: white; padding: 20px; border-radius: 10px; margin-bottom: 20px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .status { font-size: 18px; font-weight: bold; margin-bottom: 20px; }
        .status.running { color: #28a745; }
        .status.stopped { color: #dc3545; }
        .form-group { margin-bottom: 15px; }
        .form-group label { display: block; margin-bottom: 5px; font-weight: bold; }
        .form-group input, .form-group select { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        .checkbox-group { display: flex; align-items: center; }
        .checkbox-group input { width: auto; margin-right: 10px; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: bold; margin-right: 10px; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .btn:hover { opacity: 0.9; }
        .logs { background: #000; color: #00ff00; padding: 15px; border-radius: 4px; height: 300px; overflow-y: auto; font-family: monospace; font-size: 12px; }
        .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        @media (max-width: 768px) { .grid { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Optimism Challenger GUI Tray</h1>
            <p>웹 기반 챌린저 관리 인터페이스 - 브라우저에서 쉽게 챌린저를 제어하세요!</p>
        </div>

        <div class="card">
            <div id="status" class="status">상태 로딩 중...</div>
            <div>
                <button class="btn btn-success" onclick="startChallenger()">🚀 챌린저 시작</button>
                <button class="btn btn-danger" onclick="stopChallenger()">⏹️ 챌린저 정지</button>
                <button class="btn btn-secondary" onclick="saveConfig()">💾 설정 저장</button>
                <button class="btn btn-secondary" onclick="loadConfig()">📁 설정 로드</button>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>기본 설정</h3>
                <div class="form-group">
                    <label>네트워크:</label>
                    <select id="network">
                        <option value="sepolia">Sepolia (테스트넷)</option>
                        <option value="mainnet">Mainnet</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>L1 RPC:</label>
                    <input type="text" id="l1_eth_rpc" placeholder="https://ethereum-sepolia-rpc.publicnode.com">
                </div>
                <div class="form-group">
                    <label>L2 RPC:</label>
                    <input type="text" id="l2_eth_rpc" placeholder="https://sepolia.optimism.io">
                </div>
                <div class="form-group">
                    <label>데이터 디렉토리:</label>
                    <input type="text" id="datadir" placeholder="/tmp/challenger-data">
                </div>
            </div>

            <div class="card">
                <h3>P2P 설정</h3>
                <div class="form-group">
                    <div class="checkbox-group">
                        <input type="checkbox" id="p2p_enabled">
                        <label>P2P 네트워킹 활성화</label>
                    </div>
                </div>
                <div class="form-group">
                    <label>P2P 리슨 주소:</label>
                    <input type="text" id="p2p_listen_addr" placeholder="/ip4/0.0.0.0/tcp/9876">
                </div>
                <div class="form-group">
                    <label>P2P 네트워크 ID:</label>
                    <input type="text" id="p2p_network_id" placeholder="optimism-challenger-sepolia">
                </div>
                <div class="form-group">
                    <label>최대 피어 수:</label>
                    <input type="number" id="p2p_max_peers" value="50">
                </div>
            </div>
        </div>

        <div class="card">
            <h3>실시간 로그</h3>
            <div id="logs" class="logs">로그가 여기에 표시됩니다...</div>
        </div>
    </div>

    <script>
        let isRunning = false;

        function updateStatus() {
            fetch('/api/status')
                .then(response => response.json())
                .then(data => {
                    const statusEl = document.getElementById('status');
                    isRunning = data.running;
                    statusEl.textContent = data.running ? '상태: ✅ 실행 중' : '상태: ⏹️ 정지됨';
                    statusEl.className = 'status ' + (data.running ? 'running' : 'stopped');
                });
        }

        function updateLogs() {
            fetch('/api/logs')
                .then(response => response.json())
                .then(data => {
                    const logsEl = document.getElementById('logs');
                    logsEl.innerHTML = data.logs.join('<br>');
                    logsEl.scrollTop = logsEl.scrollHeight;
                });
        }

        function loadConfigToForm() {
            fetch('/api/config')
                .then(response => response.json())
                .then(config => {
                    document.getElementById('network').value = config.network || 'sepolia';
                    document.getElementById('l1_eth_rpc').value = config.l1_eth_rpc || '';
                    document.getElementById('l2_eth_rpc').value = config.l2_eth_rpc || '';
                    document.getElementById('datadir').value = config.datadir || '';
                    document.getElementById('p2p_enabled').checked = config.p2p_enabled || false;
                    document.getElementById('p2p_listen_addr').value = config.p2p_listen_addr || '';
                    document.getElementById('p2p_network_id').value = config.p2p_network_id || '';
                    document.getElementById('p2p_max_peers').value = config.p2p_max_peers || 50;
                });
        }

        function startChallenger() {
            const config = getFormConfig();
            fetch('/api/start', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            }).then(response => response.json()).then(data => {
                alert(data.message);
                updateStatus();
            });
        }

        function stopChallenger() {
            fetch('/api/stop', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert(data.message);
                    updateStatus();
                });
        }

        function saveConfig() {
            const config = getFormConfig();
            fetch('/api/config', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(config)
            }).then(response => response.json()).then(data => {
                alert(data.message);
            });
        }

        function loadConfig() {
            loadConfigToForm();
            alert('설정이 로드되었습니다.');
        }

        function getFormConfig() {
            return {
                network: document.getElementById('network').value,
                l1_eth_rpc: document.getElementById('l1_eth_rpc').value,
                l2_eth_rpc: document.getElementById('l2_eth_rpc').value,
                datadir: document.getElementById('datadir').value,
                p2p_enabled: document.getElementById('p2p_enabled').checked,
                p2p_listen_addr: document.getElementById('p2p_listen_addr').value,
                p2p_network_id: document.getElementById('p2p_network_id').value,
                p2p_max_peers: parseInt(document.getElementById('p2p_max_peers').value)
            };
        }

        // 초기화 및 자동 업데이트
        document.addEventListener('DOMContentLoaded', function() {
            loadConfigToForm();
            updateStatus();
            updateLogs();
            
            // 1초마다 상태 및 로그 업데이트
            setInterval(updateStatus, 1000);
            setInterval(updateLogs, 1000);
        });
    </script>
</body>
</html>
`

func NewWebChallengerTray() *WebChallengerTray {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".optimism", "challenger")
	os.MkdirAll(configDir, 0755)
	configFilePath := filepath.Join(configDir, "config.yaml")

	config := &ChallengerConfig{
		Network:        "sepolia",
		L1EthRPC:       "https://ethereum-sepolia-rpc.publicnode.com",
		L1Beacon:       "https://ethereum-sepolia-beacon-api.publicnode.com",
		L2EthRPC:       "https://sepolia.optimism.io",
		RollupRPC:      "https://sepolia.optimism.io",
		DataDir:        filepath.Join(configDir, "data"),
		P2PEnabled:     false,
		P2PListenAddr:  "/ip4/0.0.0.0/tcp/9876",
		P2PNetworkID:   "optimism-challenger-sepolia",
		P2PMaxPeers:    50,
		P2PBootnodes:   []string{},
		CannonBin:      "./bin/cannon",
		CannonServer:   "./bin/op-program",
		LogLevel:       "info",
		MetricsEnabled: true,
		MetricsAddr:    "0.0.0.0",
		MetricsPort:    7300,
	}

	return &WebChallengerTray{
		config:         config,
		isRunning:      false,
		logs:          []string{"웹 GUI 트레이 애플리케이션이 시작되었습니다."},
		configFilePath: configFilePath,
	}
}

func (wct *WebChallengerTray) addLog(message string) {
	timestamp := time.Now().Format("15:04:05")
	logMessage := fmt.Sprintf("[%s] %s", timestamp, message)
	wct.logs = append(wct.logs, logMessage)
	
	// 로그 최대 1000개 유지
	if len(wct.logs) > 1000 {
		wct.logs = wct.logs[len(wct.logs)-1000:]
	}
}

func (wct *WebChallengerTray) startChallenger(config *ChallengerConfig) error {
	if wct.isRunning {
		return fmt.Errorf("챌린저가 이미 실행 중입니다")
	}

	wct.config = config
	os.MkdirAll(config.DataDir, 0755)

	args := []string{
		"--network", config.Network,
		"--l1-eth-rpc", config.L1EthRPC,
		"--l1-beacon", config.L1Beacon,
		"--l2-eth-rpc", config.L2EthRPC,
		"--rollup-rpc", config.RollupRPC,
		"--datadir", config.DataDir,
		"--cannon-bin", config.CannonBin,
		"--cannon-server", config.CannonServer,
		"--log.level", config.LogLevel,
	}

	if config.P2PEnabled {
		args = append(args,
			"--p2p-enabled",
			"--p2p-listen-addr", config.P2PListenAddr,
			"--p2p-network-id", config.P2PNetworkID,
			"--p2p-max-peers", strconv.Itoa(config.P2PMaxPeers),
		)
	}

	if config.MetricsEnabled {
		args = append(args,
			"--metrics-enabled",
			"--metrics-addr", config.MetricsAddr,
			"--metrics-port", strconv.Itoa(config.MetricsPort),
		)
	}

	binPath := "./bin/op-challenger"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		binPath = "../op-challenger/bin/op-challenger"
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			return fmt.Errorf("op-challenger 실행 파일을 찾을 수 없습니다")
		}
	}

	wct.addLog(fmt.Sprintf("챌린저 시작 중... 명령어: %s %s", binPath, strings.Join(args, " ")))

	cmd := exec.Command(binPath, args...)
	
	// 로그 캡처를 위한 파이프 설정
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		wct.addLog(fmt.Sprintf("챌린저 시작 실패: %v", err))
		return err
	}

	wct.process = cmd
	wct.isRunning = true
	wct.addLog("챌린저가 성공적으로 시작되었습니다.")

	// 로그 출력 고루틴
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			wct.addLog(fmt.Sprintf("[OUT] %s", string(buf[:n])))
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			wct.addLog(fmt.Sprintf("[ERR] %s", string(buf[:n])))
		}
	}()

	// 프로세스 모니터링
	go func() {
		err := cmd.Wait()
		wct.isRunning = false
		if err != nil {
			wct.addLog(fmt.Sprintf("챌린저가 오류로 종료되었습니다: %v", err))
		} else {
			wct.addLog("챌린저가 정상적으로 종료되었습니다.")
		}
	}()

	return nil
}

func (wct *WebChallengerTray) stopChallenger() error {
	if !wct.isRunning || wct.process == nil {
		return fmt.Errorf("정지할 챌린저 프로세스가 없습니다")
	}

	wct.addLog("챌린저 정지 중...")
	
	if err := wct.process.Process.Signal(syscall.SIGTERM); err != nil {
		wct.addLog(fmt.Sprintf("SIGTERM 전송 실패: %v", err))
		if err := wct.process.Process.Kill(); err != nil {
			wct.addLog(fmt.Sprintf("강제 종료 실패: %v", err))
			return err
		}
	}

	wct.isRunning = false
	wct.addLog("챌린저 정지 요청을 전송했습니다.")
	return nil
}

func (wct *WebChallengerTray) saveConfig() error {
	data, err := yaml.Marshal(wct.config)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(wct.configFilePath, data, 0644); err != nil {
		return err
	}
	
	wct.addLog("설정이 저장되었습니다.")
	return nil
}

func (wct *WebChallengerTray) loadConfig() error {
	data, err := os.ReadFile(wct.configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			wct.addLog("설정 파일이 없어서 기본 설정을 사용합니다.")
			return nil
		}
		return err
	}
	
	if err := yaml.Unmarshal(data, wct.config); err != nil {
		return err
	}
	
	wct.addLog("설정이 로드되었습니다.")
	return nil
}

func main() {
	wct := NewWebChallengerTray()
	wct.loadConfig()

	// HTML 페이지 제공
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, htmlTemplate)
	})

	// API 엔드포인트들
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"running": wct.isRunning,
		})
	})

	http.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"logs": wct.logs,
		})
	})

	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(wct.config)
		} else if r.Method == "POST" {
			var config ChallengerConfig
			if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			wct.config = &config
			if err := wct.saveConfig(); err != nil {
				json.NewEncoder(w).Encode(map[string]string{"message": "설정 저장 실패: " + err.Error()})
			} else {
				json.NewEncoder(w).Encode(map[string]string{"message": "설정이 저장되었습니다."})
			}
		}
	})

	http.HandleFunc("/api/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var config ChallengerConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"message": "설정 파싱 실패: " + err.Error()})
			return
		}

		if err := wct.startChallenger(&config); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"message": "챌린저 시작 실패: " + err.Error()})
		} else {
			json.NewEncoder(w).Encode(map[string]string{"message": "챌린저가 성공적으로 시작되었습니다."})
		}
	})

	http.HandleFunc("/api/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := wct.stopChallenger(); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"message": "챌린저 정지 실패: " + err.Error()})
		} else {
			json.NewEncoder(w).Encode(map[string]string{"message": "챌린저가 정지되었습니다."})
		}
	})

	fmt.Println("🎉 Optimism Challenger 웹 GUI가 시작되었습니다!")
	fmt.Println("📱 브라우저에서 http://localhost:8080 을 열어주세요")
	fmt.Println("🔧 웹 인터페이스에서 모든 챌린저 기능을 제어할 수 있습니다")
	fmt.Println("")
	fmt.Println("🚀 주요 기능:")
	fmt.Println("   • GUI로 챌린저 시작/정지")
	fmt.Println("   • P2P 네트워킹 설정")
	fmt.Println("   • 실시간 로그 모니터링")
	fmt.Println("   • 설정 저장/로드")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}