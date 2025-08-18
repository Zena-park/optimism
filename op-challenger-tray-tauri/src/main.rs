#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::process::{Child, Command, Stdio};
use std::sync::{Arc, Mutex};
use tauri::{
    AppHandle, CustomMenuItem, Manager, State, SystemTray, SystemTrayEvent, SystemTrayMenu,
    SystemTrayMenuItem, WindowEvent,
};

#[derive(Debug, Serialize, Deserialize, Clone)]
struct ChallengerConfig {
    network: String,
    l1_eth_rpc: String,
    l1_beacon: String,
    l2_eth_rpc: String,
    rollup_rpc: String,
    datadir: String,
    p2p_enabled: bool,
    p2p_listen_addr: String,
    p2p_network_id: String,
    p2p_bootnodes: Vec<String>,
    p2p_max_peers: u32,
    cannon_bin: String,
    cannon_server: String,
    log_level: String,
}

impl Default for ChallengerConfig {
    fn default() -> Self {
        Self {
            network: "sepolia".to_string(),
            l1_eth_rpc: "https://ethereum-sepolia-rpc.publicnode.com".to_string(),
            l1_beacon: "https://ethereum-sepolia-beacon-api.publicnode.com".to_string(),
            l2_eth_rpc: "https://sepolia.optimism.io".to_string(),
            rollup_rpc: "https://sepolia.optimism.io".to_string(),
            datadir: "/tmp/challenger-data".to_string(),
            p2p_enabled: true,
            p2p_listen_addr: "/ip4/0.0.0.0/tcp/9876".to_string(),
            p2p_network_id: "optimism-challenger-sepolia".to_string(),
            p2p_bootnodes: Vec::new(),
            p2p_max_peers: 50,
            cannon_bin: "./bin/cannon".to_string(),
            cannon_server: "./bin/op-program".to_string(),
            log_level: "info".to_string(),
        }
    }
}

#[derive(Debug)]
struct AppState {
    challenger_process: Arc<Mutex<Option<Child>>>,
    config: Arc<Mutex<ChallengerConfig>>,
    logs: Arc<Mutex<Vec<String>>>,
    is_running: Arc<Mutex<bool>>,
}

impl AppState {
    fn new() -> Self {
        Self {
            challenger_process: Arc::new(Mutex::new(None)),
            config: Arc::new(Mutex::new(ChallengerConfig::default())),
            logs: Arc::new(Mutex::new(Vec::new())),
            is_running: Arc::new(Mutex::new(false)),
        }
    }

    fn add_log(&self, message: String) {
        let mut logs = self.logs.lock().unwrap();
        let timestamp = chrono::Utc::now().format("%H:%M:%S").to_string();
        logs.push(format!("[{}] {}", timestamp, message));
        
        // Keep maximum 1000 logs
        if logs.len() > 1000 {
            let len = logs.len();
            logs.drain(0..len - 1000);
        }
    }
}

#[tauri::command]
async fn get_status(state: State<'_, AppState>) -> Result<HashMap<String, serde_json::Value>, String> {
    let is_running = *state.is_running.lock().unwrap();
    let mut result = HashMap::new();
    result.insert("running".to_string(), serde_json::Value::Bool(is_running));
    Ok(result)
}

#[tauri::command]
async fn get_logs(state: State<'_, AppState>) -> Result<HashMap<String, serde_json::Value>, String> {
    let logs = state.logs.lock().unwrap().clone();
    let mut result = HashMap::new();
    result.insert("logs".to_string(), serde_json::Value::Array(
        logs.into_iter().map(serde_json::Value::String).collect()
    ));
    Ok(result)
}

#[tauri::command]
async fn get_config(state: State<'_, AppState>) -> Result<ChallengerConfig, String> {
    let config = state.config.lock().unwrap().clone();
    Ok(config)
}

#[tauri::command]
async fn save_config(
    config: ChallengerConfig,
    state: State<'_, AppState>,
) -> Result<HashMap<String, String>, String> {
    *state.config.lock().unwrap() = config;
    state.add_log("Settings saved successfully.".to_string());
    
    let mut result = HashMap::new();
    result.insert("message".to_string(), "Settings saved successfully.".to_string());
    Ok(result)
}

#[tauri::command]
async fn start_challenger(
    config: ChallengerConfig,
    state: State<'_, AppState>,
    app: AppHandle,
) -> Result<HashMap<String, String>, String> {
    let mut is_running = state.is_running.lock().unwrap();
    if *is_running {
        let mut result = HashMap::new();
        result.insert("message".to_string(), "Challenger is already running.".to_string());
        return Ok(result);
    }

    // 설정 업데이트
    *state.config.lock().unwrap() = config.clone();

    // 명령어 구성
    let mut args = vec![
        "--network".to_string(), config.network,
        "--l1-eth-rpc".to_string(), config.l1_eth_rpc,
        "--l1-beacon".to_string(), config.l1_beacon,
        "--l2-eth-rpc".to_string(), config.l2_eth_rpc,
        "--rollup-rpc".to_string(), config.rollup_rpc,
        "--datadir".to_string(), config.datadir,
        "--cannon-bin".to_string(), config.cannon_bin,
        "--cannon-server".to_string(), config.cannon_server,
        "--log.level".to_string(), config.log_level,
    ];

    if config.p2p_enabled {
        args.extend(vec![
            "--p2p-enabled".to_string(),
            "--p2p-listen-addr".to_string(), config.p2p_listen_addr,
            "--p2p-network-id".to_string(), config.p2p_network_id,
            "--p2p-max-peers".to_string(), config.p2p_max_peers.to_string(),
        ]);
        
        // 부트노드 추가
        for bootnode in &config.p2p_bootnodes {
            args.extend(vec![
                "--p2p-bootnodes".to_string(),
                bootnode.clone(),
            ]);
        }
    }


    // 실행 파일 경로 찾기
    let bin_path = if std::path::Path::new("./bin/op-challenger").exists() {
        "./bin/op-challenger"
    } else if std::path::Path::new("../op-challenger/bin/op-challenger").exists() {
        "../op-challenger/bin/op-challenger"
    } else {
        state.add_log("Error: op-challenger executable not found.".to_string());
        let mut result = HashMap::new();
        result.insert("message".to_string(), "op-challenger executable not found.".to_string());
        return Ok(result);
    };

    state.add_log(format!("Starting challenger... Command: {} {}", bin_path, args.join(" ")));

    // 프로세스 시작
    let mut cmd = Command::new(bin_path);
    cmd.args(&args)
        .stdout(Stdio::piped())
        .stderr(Stdio::piped());

    match cmd.spawn() {
        Ok(child) => {
            *state.challenger_process.lock().unwrap() = Some(child);
            *is_running = true;
            state.add_log("Challenger started successfully.".to_string());
            
            // 트레이 아이콘 상태 업데이트
            let tray_handle = app.tray_handle().get_item("status");
            let _ = tray_handle.set_title("✅ Running");

            let mut result = HashMap::new();
            result.insert("message".to_string(), "Challenger started successfully.".to_string());
            Ok(result)
        }
        Err(e) => {
            state.add_log(format!("Failed to start challenger: {}", e));
            let mut result = HashMap::new();
            result.insert("message".to_string(), format!("Failed to start challenger: {}", e));
            Ok(result)
        }
    }
}

#[tauri::command]
async fn stop_challenger(
    state: State<'_, AppState>,
    app: AppHandle,
) -> Result<HashMap<String, String>, String> {
    let mut is_running = state.is_running.lock().unwrap();
    let mut process = state.challenger_process.lock().unwrap();

    if !*is_running || process.is_none() {
        let mut result = HashMap::new();
        result.insert("message".to_string(), "No challenger process to stop.".to_string());
        return Ok(result);
    }

    state.add_log("Stopping challenger...".to_string());

    if let Some(mut child) = process.take() {
        match child.kill() {
            Ok(_) => {
                state.add_log("Challenger stopped.".to_string());
                *is_running = false;
                
                // 트레이 아이콘 상태 업데이트
                let tray_handle = app.tray_handle().get_item("status");
                let _ = tray_handle.set_title("⏹️ Stopped");

                let mut result = HashMap::new();
                result.insert("message".to_string(), "Challenger stopped.".to_string());
                Ok(result)
            }
            Err(e) => {
                state.add_log(format!("Failed to stop challenger: {}", e));
                let mut result = HashMap::new();
                result.insert("message".to_string(), format!("Failed to stop challenger: {}", e));
                Ok(result)
            }
        }
    } else {
        let mut result = HashMap::new();
        result.insert("message".to_string(), "Process not found.".to_string());
        Ok(result)
    }
}

fn create_tray() -> SystemTray {
    let quit = CustomMenuItem::new("quit".to_string(), "Quit");
    let toggle = CustomMenuItem::new("toggle".to_string(), "Show/Hide Window");
    let start = CustomMenuItem::new("start".to_string(), "Start Challenger");
    let stop = CustomMenuItem::new("stop".to_string(), "Stop Challenger");
    let status = CustomMenuItem::new("status".to_string(), "⏹️ Stopped").disabled();
    
    let tray_menu = SystemTrayMenu::new()
        .add_item(status)
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(toggle)
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(start)
        .add_item(stop)
        .add_native_item(SystemTrayMenuItem::Separator)
        .add_item(quit);

    SystemTray::new().with_menu(tray_menu).with_tooltip("Optimism Challenger Tray")
}

fn main() {
    let app_state = AppState::new();
    
    tauri::Builder::default()
        .manage(app_state)
        .system_tray(create_tray())
        .on_system_tray_event(|app, event| match event {
            SystemTrayEvent::LeftClick {
                position: _,
                size: _,
                ..
            } => {
                // Toggle main window visibility on left click
                let window = app.get_window("main").unwrap();
                match window.is_visible() {
                    Ok(true) => {
                        let _ = window.hide();
                    }
                    Ok(false) => {
                        let _ = window.show();
                        let _ = window.set_focus();
                    }
                    Err(_) => {
                        let _ = window.show();
                        let _ = window.set_focus();
                    }
                }
            }
            SystemTrayEvent::MenuItemClick { id, .. } => match id.as_str() {
                "quit" => {
                    std::process::exit(0);
                }
                "toggle" => {
                    let window = app.get_window("main").unwrap();
                    match window.is_visible() {
                        Ok(true) => {
                            let _ = window.hide();
                        }
                        Ok(false) => {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                        Err(_) => {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                }
                "start" => {
                    // Start directly from tray (using default settings)
                    let app_handle = app.app_handle();
                    tauri::async_runtime::spawn(async move {
                        let state: State<AppState> = app_handle.state();
                        let config = state.config.lock().unwrap().clone();
                        let _ = start_challenger(config, state, app_handle.clone()).await;
                    });
                }
                "stop" => {
                    let app_handle = app.app_handle();
                    tauri::async_runtime::spawn(async move {
                        let state: State<AppState> = app_handle.state();
                        let _ = stop_challenger(state, app_handle.clone()).await;
                    });
                }
                _ => {}
            },
            _ => {}
        })
        .on_window_event(|event| match event.event() {
            WindowEvent::CloseRequested { api, .. } => {
                // Hide to tray on window close (not complete exit)
                event.window().hide().unwrap();
                api.prevent_close();
            }
            _ => {}
        })
        .invoke_handler(tauri::generate_handler![
            get_status,
            get_logs,
            get_config,
            save_config,
            start_challenger,
            stop_challenger
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
