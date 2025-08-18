package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/getlantern/systray"
)

const (
	AppName     = "Optimism Challenger"
	AppVersion  = "1.0.0"
	AppID       = "io.optimism.challenger"
)

type ChallengerTray struct {
	app              fyne.App
	window           fyne.Window
	config           *ChallengerConfig
	process          *os.Process
	statusLabel      *widget.Label
	logTextArea      *widget.Entry
	isRunning        bool
	configFilePath   string
}

type ChallengerConfig struct {
	Network        string   `yaml:"network"`
	L1EthRPC       string   `yaml:"l1_eth_rpc"`
	L1Beacon       string   `yaml:"l1_beacon"`
	L2EthRPC       string   `yaml:"l2_eth_rpc"`
	RollupRPC      string   `yaml:"rollup_rpc"`
	DataDir        string   `yaml:"datadir"`
	P2PEnabled     bool     `yaml:"p2p_enabled"`
	P2PListenAddr  string   `yaml:"p2p_listen_addr"`
	P2PNetworkID   string   `yaml:"p2p_network_id"`
	P2PMaxPeers    int      `yaml:"p2p_max_peers"`
	P2PBootnodes   []string `yaml:"p2p_bootnodes"`
	CannonBin      string   `yaml:"cannon_bin"`
	CannonServer   string   `yaml:"cannon_server"`
	LogLevel       string   `yaml:"log_level"`
	MetricsEnabled bool     `yaml:"metrics_enabled"`
	MetricsAddr    string   `yaml:"metrics_addr"`
	MetricsPort    int      `yaml:"metrics_port"`
}

func NewChallengerTray() *ChallengerTray {
	myApp := app.NewWithID(AppID)

	window := myApp.NewWindow(AppName)
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(800, 600))

	// 기본 설정 파일 경로
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

	return &ChallengerTray{
		app:            myApp,
		window:         window,
		config:         config,
		configFilePath: configFilePath,
		isRunning:      false,
	}
}

func (ct *ChallengerTray) setupUI() {
	// 상태 표시
	ct.statusLabel = widget.NewLabel("상태: 정지됨")
	ct.statusLabel.TextStyle.Bold = true

	// 설정 폼
	networkEntry := widget.NewEntry()
	networkEntry.SetText(ct.config.Network)
	
	l1RpcEntry := widget.NewEntry()
	l1RpcEntry.SetText(ct.config.L1EthRPC)
	
	l2RpcEntry := widget.NewEntry()
	l2RpcEntry.SetText(ct.config.L2EthRPC)
	
	dataDirEntry := widget.NewEntry()
	dataDirEntry.SetText(ct.config.DataDir)
	
	p2pEnabledCheck := widget.NewCheck("P2P 네트워킹 활성화", nil)
	p2pEnabledCheck.SetChecked(ct.config.P2PEnabled)
	
	p2pListenAddrEntry := widget.NewEntry()
	p2pListenAddrEntry.SetText(ct.config.P2PListenAddr)
	
	p2pNetworkIDEntry := widget.NewEntry()
	p2pNetworkIDEntry.SetText(ct.config.P2PNetworkID)

	// 버튼들
	startBtn := widget.NewButton("챌린저 시작", func() {
		ct.updateConfigFromForm(networkEntry, l1RpcEntry, l2RpcEntry, dataDirEntry, 
			p2pEnabledCheck, p2pListenAddrEntry, p2pNetworkIDEntry)
		ct.startChallenger()
	})
	startBtn.Importance = widget.HighImportance

	stopBtn := widget.NewButton("챌린저 정지", func() {
		ct.stopChallenger()
	})
	stopBtn.Importance = widget.MediumImportance

	saveConfigBtn := widget.NewButton("설정 저장", func() {
		ct.updateConfigFromForm(networkEntry, l1RpcEntry, l2RpcEntry, dataDirEntry,
			p2pEnabledCheck, p2pListenAddrEntry, p2pNetworkIDEntry)
		ct.saveConfig()
	})

	loadConfigBtn := widget.NewButton("설정 로드", func() {
		ct.loadConfig()
		ct.updateFormFromConfig(networkEntry, l1RpcEntry, l2RpcEntry, dataDirEntry,
			p2pEnabledCheck, p2pListenAddrEntry, p2pNetworkIDEntry)
	})

	// 로그 영역
	ct.logTextArea = widget.NewMultiLineEntry()
	ct.logTextArea.SetText("챌린저 로그가 여기에 표시됩니다...\n")
	ct.logTextArea.Wrapping = fyne.TextWrapWord
	
	logScroll := container.NewScroll(ct.logTextArea)
	logScroll.SetMinSize(fyne.NewSize(750, 200))

	// 설정 폼 레이아웃
	configForm := container.NewBorder(
		widget.NewCard("기본 설정", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("네트워크:"), networkEntry,
				widget.NewLabel("L1 RPC:"), l1RpcEntry,
				widget.NewLabel("L2 RPC:"), l2RpcEntry,
				widget.NewLabel("데이터 디렉토리:"), dataDirEntry,
			),
		)),
		widget.NewCard("P2P 설정", "", container.NewVBox(
			p2pEnabledCheck,
			container.NewGridWithColumns(2,
				widget.NewLabel("P2P 리슨 주소:"), p2pListenAddrEntry,
				widget.NewLabel("P2P 네트워크 ID:"), p2pNetworkIDEntry,
			),
		)),
		nil, nil,
		nil,
	)

	// 컨트롤 버튼
	controlButtons := container.NewHBox(
		startBtn,
		stopBtn,
		widget.NewSeparator(),
		saveConfigBtn,
		loadConfigBtn,
	)

	// 메인 레이아웃
	content := container.NewBorder(
		container.NewVBox(
			ct.statusLabel,
			widget.NewSeparator(),
		),
		container.NewVBox(
			widget.NewSeparator(),
			controlButtons,
		),
		nil, nil,
		container.NewVSplit(
			configForm,
			widget.NewCard("로그", "", logScroll),
		),
	)

	ct.window.SetContent(content)
}

func (ct *ChallengerTray) updateConfigFromForm(networkEntry, l1RpcEntry, l2RpcEntry, dataDirEntry *widget.Entry,
	p2pEnabledCheck *widget.Check, p2pListenAddrEntry, p2pNetworkIDEntry *widget.Entry) {
	ct.config.Network = networkEntry.Text
	ct.config.L1EthRPC = l1RpcEntry.Text
	ct.config.L2EthRPC = l2RpcEntry.Text
	ct.config.DataDir = dataDirEntry.Text
	ct.config.P2PEnabled = p2pEnabledCheck.Checked
	ct.config.P2PListenAddr = p2pListenAddrEntry.Text
	ct.config.P2PNetworkID = p2pNetworkIDEntry.Text
}

func (ct *ChallengerTray) updateFormFromConfig(networkEntry, l1RpcEntry, l2RpcEntry, dataDirEntry *widget.Entry,
	p2pEnabledCheck *widget.Check, p2pListenAddrEntry, p2pNetworkIDEntry *widget.Entry) {
	networkEntry.SetText(ct.config.Network)
	l1RpcEntry.SetText(ct.config.L1EthRPC)
	l2RpcEntry.SetText(ct.config.L2EthRPC)
	dataDirEntry.SetText(ct.config.DataDir)
	p2pEnabledCheck.SetChecked(ct.config.P2PEnabled)
	p2pListenAddrEntry.SetText(ct.config.P2PListenAddr)
	p2pNetworkIDEntry.SetText(ct.config.P2PNetworkID)
}

func (ct *ChallengerTray) startChallenger() {
	if ct.isRunning {
		ct.appendLog("챌린저가 이미 실행 중입니다.\n")
		return
	}

	// 데이터 디렉토리 생성
	os.MkdirAll(ct.config.DataDir, 0755)

	// op-challenger 명령어 구성
	args := []string{
		"--network", ct.config.Network,
		"--l1-eth-rpc", ct.config.L1EthRPC,
		"--l1-beacon", ct.config.L1Beacon,
		"--l2-eth-rpc", ct.config.L2EthRPC,
		"--rollup-rpc", ct.config.RollupRPC,
		"--datadir", ct.config.DataDir,
		"--cannon-bin", ct.config.CannonBin,
		"--cannon-server", ct.config.CannonServer,
		"--log.level", ct.config.LogLevel,
	}

	if ct.config.P2PEnabled {
		args = append(args,
			"--p2p-enabled",
			"--p2p-listen-addr", ct.config.P2PListenAddr,
			"--p2p-network-id", ct.config.P2PNetworkID,
			"--p2p-max-peers", fmt.Sprintf("%d", ct.config.P2PMaxPeers),
		)
	}

	if ct.config.MetricsEnabled {
		args = append(args,
			"--metrics-enabled",
			"--metrics-addr", ct.config.MetricsAddr,
			"--metrics-port", fmt.Sprintf("%d", ct.config.MetricsPort),
		)
	}

	// op-challenger 실행 파일 경로 찾기
	binPath := "./bin/op-challenger"
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		binPath = "../op-challenger/bin/op-challenger"
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			ct.appendLog("오류: op-challenger 실행 파일을 찾을 수 없습니다.\n")
			dialog.ShowError(fmt.Errorf("op-challenger 실행 파일을 찾을 수 없습니다"), ct.window)
			return
		}
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binPath, args...)
	
	// 표준 출력과 오류를 로그로 리다이렉트
	cmd.Stdout = &logWriter{ct: ct, prefix: "[OUT] "}
	cmd.Stderr = &logWriter{ct: ct, prefix: "[ERR] "}

	ct.appendLog(fmt.Sprintf("챌린저 시작 중... 명령어: %s %v\n", binPath, args))
	
	if err := cmd.Start(); err != nil {
		ct.appendLog(fmt.Sprintf("챌린저 시작 실패: %v\n", err))
		dialog.ShowError(err, ct.window)
		return
	}

	ct.process = cmd.Process
	ct.isRunning = true
	ct.statusLabel.SetText("상태: 실행 중")
	ct.appendLog("챌린저가 성공적으로 시작되었습니다.\n")

	// 프로세스 모니터링
	go func() {
		err := cmd.Wait()
		ct.isRunning = false
		ct.statusLabel.SetText("상태: 정지됨")
		if err != nil {
			ct.appendLog(fmt.Sprintf("챌린저가 오류로 종료되었습니다: %v\n", err))
		} else {
			ct.appendLog("챌린저가 정상적으로 종료되었습니다.\n")
		}
	}()
}

func (ct *ChallengerTray) stopChallenger() {
	if !ct.isRunning || ct.process == nil {
		ct.appendLog("정지할 챌린저 프로세스가 없습니다.\n")
		return
	}

	ct.appendLog("챌린저 정지 중...\n")
	
	// SIGTERM 신호 전송
	if err := ct.process.Signal(syscall.SIGTERM); err != nil {
		ct.appendLog(fmt.Sprintf("SIGTERM 전송 실패: %v\n", err))
		// 강제 종료 시도
		if err := ct.process.Kill(); err != nil {
			ct.appendLog(fmt.Sprintf("강제 종료 실패: %v\n", err))
		}
	}

	ct.isRunning = false
	ct.statusLabel.SetText("상태: 정지됨")
	ct.appendLog("챌린저 정지 요청을 전송했습니다.\n")
}

func (ct *ChallengerTray) appendLog(text string) {
	current := ct.logTextArea.Text
	ct.logTextArea.SetText(current + text)
	
	// 자동 스크롤 (맨 아래로)
	ct.logTextArea.CursorRow = len(ct.logTextArea.Text)
}

func (ct *ChallengerTray) saveConfig() {
	if err := saveConfigToFile(ct.config, ct.configFilePath); err != nil {
		ct.appendLog(fmt.Sprintf("설정 저장 실패: %v\n", err))
		dialog.ShowError(err, ct.window)
	} else {
		ct.appendLog("설정이 저장되었습니다.\n")
		dialog.ShowInformation("저장 완료", "설정이 성공적으로 저장되었습니다.", ct.window)
	}
}

func (ct *ChallengerTray) loadConfig() {
	if err := loadConfigFromFile(ct.configFilePath, ct.config); err != nil {
		ct.appendLog(fmt.Sprintf("설정 로드 실패: %v\n", err))
		dialog.ShowError(err, ct.window)
	} else {
		ct.appendLog("설정이 로드되었습니다.\n")
	}
}

func (ct *ChallengerTray) setupSystemTray() {
	go func() {
		systray.Run(ct.onSystrayReady, ct.onSystrayExit)
	}()
}

func (ct *ChallengerTray) onSystrayReady() {
	systray.SetIcon(getIconData())
	systray.SetTitle(AppName)
	systray.SetTooltip(AppName + " - Optimism 챌린저 관리")

	mShow := systray.AddMenuItem("창 보기", "메인 창을 표시합니다")
	mStart := systray.AddMenuItem("챌린저 시작", "챌린저를 시작합니다")
	mStop := systray.AddMenuItem("챌린저 정지", "챌린저를 정지합니다")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("종료", "애플리케이션을 종료합니다")

	for {
		select {
		case <-mShow.ClickedCh:
			ct.window.Show()
		case <-mStart.ClickedCh:
			ct.startChallenger()
		case <-mStop.ClickedCh:
			ct.stopChallenger()
		case <-mQuit.ClickedCh:
			ct.app.Quit()
			return
		}
	}
}

func (ct *ChallengerTray) onSystrayExit() {
	// 정리 작업
	if ct.isRunning && ct.process != nil {
		ct.stopChallenger()
	}
}

func (ct *ChallengerTray) Run() {
	ct.setupUI()
	ct.loadConfig() // 시작 시 설정 로드
	ct.setupSystemTray()
	ct.window.ShowAndRun()
}

type logWriter struct {
	ct     *ChallengerTray
	prefix string
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.ct.appendLog(lw.prefix + string(p))
	return len(p), nil
}

func main() {
	tray := NewChallengerTray()
	tray.Run()
}